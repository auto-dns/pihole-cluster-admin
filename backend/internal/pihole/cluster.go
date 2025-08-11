package pihole

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/auto-dns/pihole-cluster-admin/internal/util"
	"github.com/rs/zerolog"
	"golang.org/x/sync/errgroup"
)

type Cluster struct {
	clients       map[int64]ClientInterface
	cursorManager CursorManagerInterface[FetchQueryLogFilters]
	logger        zerolog.Logger
	rw            sync.RWMutex
}

func NewCluster(clients map[int64]ClientInterface, cursorManager CursorManagerInterface[FetchQueryLogFilters], logger zerolog.Logger) ClusterInterface {
	return &Cluster{
		clients:       clients,
		cursorManager: cursorManager,
		logger:        logger,
	}
}

func (c *Cluster) AddClient(ctx context.Context, client ClientInterface) error {
	c.rw.Lock()
	defer c.rw.Unlock()

	id := client.GetId(ctx)
	logger := c.logger.With().Int64("id", id).Str("name", client.GetName(ctx)).Str("scheme", client.GetScheme(ctx)).Str("host", client.GetHost(ctx)).Int("port", client.GetPort(ctx)).Logger()

	if _, exists := c.clients[id]; exists {
		err := errors.New("client id already exists")
		logger.Error().Err(err).Msg("client id conflict")
		return err
	}

	c.clients[id] = client
	c.logger.Debug().Msg("client added to cluster")

	return nil
}

func (c *Cluster) RemoveClient(ctx context.Context, id int64) error {
	c.rw.Lock()
	defer c.rw.Unlock()

	logger := c.logger.With().Int64("id", id).Logger()

	client, exists := c.clients[id]
	if !exists {
		err := errors.New("client id not found")
		logger.Error().Err(err).Msg("client id not found")
		return err
	}

	delete(c.clients, id)
	c.logger.Debug().Int64("id", client.GetId(ctx)).Str("name", client.GetName(ctx)).Str("scheme", client.GetScheme(ctx)).Str("host", client.GetHost(ctx)).Int("port", client.GetPort(ctx)).Msg("client removed from cluster")

	return nil
}

func (c *Cluster) UpdateClient(ctx context.Context, id int64, cfg *ClientConfig) error {
	c.rw.Lock()
	defer c.rw.Unlock()

	logger := c.logger.With().Int64("id", id).Str("name", cfg.Name).Str("scheme", cfg.Scheme).Str("host", cfg.Host).Int("port", cfg.Port).Logger()

	if id != cfg.Id {
		err := errors.New("id must match cfg.Id")
		logger.Error().Err(err).Msg("id must match cfg.Id")
		return err
	}

	if _, exists := c.clients[id]; !exists {
		err := errors.New("client id not found")
		logger.Error().Err(err).Msg("client id not found")
		return err
	}

	c.clients[id].Update(ctx, cfg)
	logger.Debug().Msg("updated client")
	return nil
}

func (c *Cluster) HasClient(ctx context.Context, id int64) bool {
	c.rw.RLock()
	defer c.rw.RUnlock()
	_, has := c.clients[id]
	return has
}

func (c *Cluster) forEachClient(ctx context.Context, limit int, f func(ctx context.Context, id int64, client ClientInterface) error) error {
	c.rw.RLock()
	clients := make(map[int64]ClientInterface, len(c.clients))
	for id, client := range c.clients {
		clients[id] = client
	}
	c.rw.RUnlock()

	var semaphore chan struct{}
	if limit > 0 {
		semaphore = make(chan struct{}, limit)
	}

	g, ctx := errgroup.WithContext(ctx)
	for id, client := range clients {
		id, client := id, client
		g.Go(func() error {
			defer func() {
				if r := recover(); r != nil {
					c.logger.Error().Interface("panic", r).Int64("id", id).Msg("worker panic recovered")
				}
			}()

			if semaphore != nil {
                select {
                case semaphore <- struct{}{}:
                case <-ctx.Done():
                    return ctx.Err()
                }
                defer func() { <-semaphore }()
            }

			nodeTimeout := 3 * time.Second
			if deadline, ok := ctx.Deadline(); ok {
				nodeTimeout = time.Until(deadline) / 2
                if nodeTimeout > 3*time.Second {
                    nodeTimeout = 3 * time.Second
                }
			}
			var cancel context.CancelFunc
			nodeCtx, cancel := context.WithTimeout(ctx, nodeTimeout)
			defer cancel()

			return f(nodeCtx, id, client)
		})
	}
	return g.Wait()
}

func (c *Cluster) FetchQueryLogs(ctx context.Context, req FetchQueryLogClusterRequest) (*FetchQueryLogsClusterResponse, error) {
	c.logger.Debug().Msg("fetching query logs from all pihole nodes")

	var searchState SearchStateInterface[FetchQueryLogFilters]
	// -- Cluster cursor is passed in from the browser
	if req.Cursor != nil && *req.Cursor != "" {
		var ok bool
		searchState, ok = c.cursorManager.GetSearchState(*req.Cursor)
		if !ok {
			return nil, fmt.Errorf("cursor expired or not found")
		}
	}

	results := make(map[int64]*NodeResult[FetchQueryLogResponse], len(c.clients))
	nextPiholeCursors := make(map[int64]int)
	var mu sync.Mutex

	err := c.forEachClient(ctx, 0, func(nodeCtx context.Context, id int64, client ClientInterface) error {
		nodeReq := FetchQueryLogClientRequest{
			Filters: req.Filters, // use either user-provided filters (no cursor) or cursor snapshot
			Length:  req.Length,
			Start:   req.Start,
			Cursor:  nil,
		}

		if searchState != nil {
			if cursor, ok := searchState.GetPiholeCursor(id); ok {
				nodeReq.Filters = searchState.GetRequestParams()
				nodeReq.Start = nil
				cur := cursor
				nodeReq.Cursor = &cur
			} else {
				c.logger.Warn().Int64("id", id).Msg("pihole cursor not found")
			}
		}

		response, err := client.FetchQueryLogs(nodeCtx, nodeReq)

		node := client.GetNodeInfo(nodeCtx)
		mu.Lock()
		{
			results[id] = &NodeResult[FetchQueryLogResponse]{
				PiholeNode: node,
				Success:    err == nil,
				Error:      util.ErrorString(err),
				Response:   response,
			}

			// Save the node cursor for future pagination
			if err == nil && response != nil {
				nextPiholeCursors[id] = response.Cursor
			}	
		}
		mu.Unlock()

		if err != nil {
            c.logger.Warn().Int64("id", id).Str("error", util.ErrorString(err)).Msg("node operation failed")
        }
		return nil
	})
	if err != nil && errors.Is(err, context.Canceled) {
		c.logger.Warn().Err(err).Msg("fan-out aborted")
	}

	// Determine if node cursors changed
	var changed bool
	if req.Cursor == nil || *req.Cursor == "" {
		changed = true // first call always creates new cursor
	} else if searchState != nil {
		for piholeId, newCursor := range nextPiholeCursors {
			oldCursor, ok := searchState.GetPiholeCursor(piholeId)
			if !ok || oldCursor != newCursor {
				// !ok means the old cursor didn't exist
				// oldCursor != newCursor - we got a new cursor back
				changed = true
				break
			}
		}
	}

	// If no changes in cursors, reuse existing cursor and mark end of results
	if !changed && req.Cursor != nil {
		return &FetchQueryLogsClusterResponse{
			Cursor:       *req.Cursor,
			Results:      results,
			EndOfResults: true,
		}, nil
	}

	// Create a new cursor snapshot
	newCursor := c.cursorManager.CreateCursor(req.Filters, nextPiholeCursors)
	if !changed && req.Cursor != nil {
		c.logger.Debug().Str("cursor_id", *req.Cursor).Msg("no new data, reusing cursor")
	} else {
		c.logger.Debug().Str("cursor_id", newCursor).Msg("new cursor created for query logs")
	}

	return &FetchQueryLogsClusterResponse{
		Cursor:       newCursor,
		Results:      results,
		EndOfResults: false,
	}, nil
}

func (c *Cluster) GetDomainRules(ctx context.Context, opts GetDomainRulesOptions) map[int64]*NodeResult[GetDomainRulesResponse] {
	c.logger.Debug().Msg("getting domain rules from all pihole nodes")

	results := make(map[int64]*NodeResult[GetDomainRulesResponse], len(c.clients))
	var mu sync.Mutex
	err := c.forEachClient(ctx, 0, func(nodeCtx context.Context, id int64, client ClientInterface) error {
		res, err := client.GetDomainRules(nodeCtx, opts)
		node := client.GetNodeInfo(nodeCtx)
		mu.Lock()
		results[id] = &NodeResult[GetDomainRulesResponse]{
			PiholeNode: node,
			Success:    err == nil,
			Error:      util.ErrorString(err),
			Response:   res,
		}
		mu.Unlock()

		if err != nil {
			c.logger.Warn().Int64("id", id).Str("error", util.ErrorString(err)).Msg("node operation failed")
		}
		return nil
	})
	if err != nil && errors.Is(err, context.Canceled) {
		c.logger.Warn().Err(err).Msg("fan-out aborted")
	}

	return results
}

func (c *Cluster) AddDomainRule(ctx context.Context, opts AddDomainRuleOptions) map[int64]*NodeResult[AddDomainRuleResponse] {
	c.logger.Debug().Msg("adding domain rule to all pihole nodes")

	results := make(map[int64]*NodeResult[AddDomainRuleResponse], len(c.clients))
	var mu sync.Mutex
	err := c.forEachClient(ctx, 0, func(nodeCtx context.Context, id int64, client ClientInterface) error {
		r, err := client.AddDomainRule(nodeCtx, opts)
		node := client.GetNodeInfo(nodeCtx)
		mu.Lock()
		results[id] = &NodeResult[AddDomainRuleResponse]{
			PiholeNode: node,
			Success:    err == nil,
			Error:      util.ErrorString(err),
			Response:   r,
		}
		mu.Unlock()
		if err != nil {
			c.logger.Warn().Int64("id", id).Str("error", util.ErrorString(err)).Msg("node operation failed")
		}
		return nil
	})
	if err != nil && errors.Is(err, context.Canceled) {
		c.logger.Warn().Err(err).Msg("fan-out aborted")
	}

	return results
}

func (c *Cluster) RemoveDomainRule(ctx context.Context, opts RemoveDomainRuleOptions) map[int64]*NodeResult[RemoveDomainRuleResponse] {
	c.logger.Debug().Msg("removing domain rule from all pihole nodes")

	results := make(map[int64]*NodeResult[RemoveDomainRuleResponse], len(c.clients))
	var mu sync.Mutex
	err := c.forEachClient(ctx, 0, func(nodeCtx context.Context, id int64, client ClientInterface) error {
		err := client.RemoveDomainRule(nodeCtx, opts)
		node := client.GetNodeInfo(nodeCtx)
		mu.Lock()
		results[id] = &NodeResult[RemoveDomainRuleResponse]{
			PiholeNode: node,
			Success:    err == nil,
			Error:      util.ErrorString(err),
			// Response is nil because 204 has no body
		}
		mu.Unlock()
		if err != nil {
			c.logger.Warn().Int64("id", id).Str("error", util.ErrorString(err)).Msg("node operation failed")
		}
		return nil
	})
	if err != nil && errors.Is(err, context.Canceled) {
		c.logger.Warn().Err(err).Msg("fan-out aborted")
	}

	return results
}

func (c *Cluster) Logout(ctx context.Context) map[int64]error {
	c.logger.Debug().Msg("logging out all pihole nodes")

	errs := make(map[int64]error, len(c.clients))
	var mu sync.Mutex
	err := c.forEachClient(ctx, 0, func(nodeCtx context.Context, id int64, client ClientInterface) error {
		err := client.Logout(nodeCtx)
		mu.Lock()
		errs[id] = err
		mu.Unlock()
		if err != nil {
			c.logger.Warn().Int64("id", id).Str("error", util.ErrorString(err)).Msg("node operation failed")
		}
		return nil
	})
	if err != nil && errors.Is(err, context.Canceled) {
		c.logger.Warn().Err(err).Msg("fan-out aborted")
	}
	
	return errs
}
