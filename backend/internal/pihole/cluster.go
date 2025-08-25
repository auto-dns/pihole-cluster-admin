package pihole

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/auto-dns/pihole-cluster-admin/internal/domain"
	"github.com/auto-dns/pihole-cluster-admin/internal/util"
	"github.com/rs/zerolog"
	"golang.org/x/sync/errgroup"
)

type Cluster struct {
	clients       map[int64]clientPort
	cursorManager cursorManagerPort[FetchQueryLogFilters]
	logger        zerolog.Logger
	rw            sync.RWMutex
}

func NewCluster(clientMap map[int64]*Client, cursorManager cursorManagerPort[FetchQueryLogFilters], logger zerolog.Logger) *Cluster {
	clients := make(map[int64]clientPort, len(clientMap))
	for id, c := range clientMap {
		clients[id] = c
	}
	return &Cluster{
		clients:       clients,
		cursorManager: cursorManager,
		logger:        logger,
	}
}

func (c *Cluster) AddClient(ctx context.Context, client *Client) error {
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
	logger.Debug().Msg("client added to cluster")

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

func (c *Cluster) GetBlockingState(ctx context.Context) map[int64]*domain.NodeResult[*domain.BlockingState] {
	c.logger.Debug().Msg("getting block status from all pihole nodes")
	out, _ := fanout[*domain.BlockingState](c, ctx, 0, func(nodeCtx context.Context, _ int64, client clientPort) (*domain.BlockingState, error) {
		return client.GetBlockingState(nodeCtx)
	})
	return out
}

func (c *Cluster) FetchQueryLogs(ctx context.Context, req FetchQueryLogClusterRequest) (*FetchQueryLogsClusterResponse, error) {
	c.logger.Debug().Msg("fetching query logs from all pihole nodes")

	// Get search state (from cursor, if present - nil if not)
	var state searchStatePort[FetchQueryLogFilters]
	if req.Cursor != nil && *req.Cursor != "" {
		var ok bool
		if state, ok = c.cursorManager.GetSearchState(*req.Cursor); !ok {
			return nil, fmt.Errorf("cursor expired or not found")
		}
	}

	// Create result map
	responses, err := fanout[*FetchQueryLogResponse](c, ctx, 0, func(nodeCtx context.Context, id int64, client clientPort) (*FetchQueryLogResponse, error) {
		// Build request
		nodeReq := fetchQueryLogClientRequest{
			Filters: req.Filters, // use either user-provided filters (no cursor) or cursor snapshot
			Length:  req.Length,
			Start:   req.Start,
			Cursor:  nil,
		}

		// Add cursor to individual pihole request if one exists
		if state != nil {
			if cursor, ok := state.GetPiholeCursor(id); ok {
				nodeReq.Filters = state.GetRequestParams()
				nodeReq.Start = nil
				nodeReq.Cursor = &cursor
			} else {
				c.logger.Warn().Int64("id", id).Msg("pihole cursor not found")
			}
		}

		// Make pihole client request
		return client.FetchQueryLogs(nodeCtx, nodeReq)
	})
	if err != nil && errors.Is(err, context.Canceled) {
		c.logger.Warn().Err(err).Msg("fan-out aborted")
	}

	// Determine if node cursors changed
	nextPiholeCursors := make(map[int64]int, len(responses))
	changed := req.Cursor == nil || *req.Cursor == ""
	for id, response := range responses {
		if response.Success && response.Response != nil {
			nextPiholeCursors[id] = response.Response.Cursor
			if state != nil && !changed {
				if old, ok := state.GetPiholeCursor(id); !ok || old != (*(*response).Response).Cursor {
					changed = true
				}
			}
		}
	}

	// If no changes in cursors, reuse existing cursor and mark end of results
	if !changed && req.Cursor != nil {
		return &FetchQueryLogsClusterResponse{
			Cursor:       *req.Cursor,
			Results:      responses,
			EndOfResults: true,
		}, nil
	}

	// Create a new cursor snapshot
	newCursor := c.cursorManager.CreateCursor(req.Filters, nextPiholeCursors)
	return &FetchQueryLogsClusterResponse{
		Cursor:       newCursor,
		Results:      responses,
		EndOfResults: false,
	}, nil
}

func (c *Cluster) GetAllDomainRules(ctx context.Context) map[int64]*domain.NodeResult[*GetDomainRulesResponse] {
	c.logger.Debug().Msg("getting domain rules from all pihole nodes")
	out, _ := fanout[*GetDomainRulesResponse](c, ctx, 0, func(nodeCtx context.Context, _ int64, client clientPort) (*GetDomainRulesResponse, error) {
		return client.GetAllDomainRules(nodeCtx)
	})
	return out
}

func (c *Cluster) GetDomainRulesByType(ctx context.Context, opts GetDomainRulesByTypeOptions) map[int64]*domain.NodeResult[*GetDomainRulesResponse] {
	c.logger.Debug().Msg("getting domain rules from all pihole nodes")
	out, _ := fanout[*GetDomainRulesResponse](c, ctx, 0, func(nodeCtx context.Context, _ int64, client clientPort) (*GetDomainRulesResponse, error) {
		return client.GetDomainRulesByType(nodeCtx, opts)
	})
	return out
}

func (c *Cluster) GetDomainRulesByKind(ctx context.Context, opts GetDomainRulesByKindOptions) map[int64]*domain.NodeResult[*GetDomainRulesResponse] {
	c.logger.Debug().Msg("getting domain rules from all pihole nodes")
	out, _ := fanout[*GetDomainRulesResponse](c, ctx, 0, func(nodeCtx context.Context, _ int64, client clientPort) (*GetDomainRulesResponse, error) {
		return client.GetDomainRulesByKind(nodeCtx, opts)
	})
	return out
}

func (c *Cluster) GetDomainRulesByDomain(ctx context.Context, opts GetDomainRulesByDomainOptions) map[int64]*domain.NodeResult[*GetDomainRulesResponse] {
	c.logger.Debug().Msg("getting domain rules from all pihole nodes")
	out, _ := fanout[*GetDomainRulesResponse](c, ctx, 0, func(nodeCtx context.Context, _ int64, client clientPort) (*GetDomainRulesResponse, error) {
		return client.GetDomainRulesByDomain(nodeCtx, opts)
	})
	return out
}

func (c *Cluster) GetDomainRulesByTypeKind(ctx context.Context, opts GetDomainRulesByTypeKindOptions) map[int64]*domain.NodeResult[*GetDomainRulesResponse] {
	c.logger.Debug().Msg("getting domain rules from all pihole nodes")
	out, _ := fanout[*GetDomainRulesResponse](c, ctx, 0, func(nodeCtx context.Context, _ int64, client clientPort) (*GetDomainRulesResponse, error) {
		return client.GetDomainRulesByTypeKind(nodeCtx, opts)
	})
	return out
}

func (c *Cluster) GetDomainRulesByTypeKindDomain(ctx context.Context, opts GetDomainRulesByTypeKindDomainOptions) map[int64]*domain.NodeResult[*GetDomainRulesResponse] {
	c.logger.Debug().Msg("getting domain rules from all pihole nodes")
	out, _ := fanout[*GetDomainRulesResponse](c, ctx, 0, func(nodeCtx context.Context, _ int64, client clientPort) (*GetDomainRulesResponse, error) {
		return client.GetDomainRulesByTypeKindDomain(nodeCtx, opts)
	})
	return out
}

func (c *Cluster) AddDomainRule(ctx context.Context, opts AddDomainRuleOptions) map[int64]*domain.NodeResult[*AddDomainRuleResponse] {
	c.logger.Debug().Msg("adding domain rule to all pihole nodes")
	out, _ := fanout[*AddDomainRuleResponse](c, ctx, 0, func(nodeCtx context.Context, _ int64, client clientPort) (*AddDomainRuleResponse, error) {
		return client.AddDomainRule(nodeCtx, opts)
	})
	return out
}

func (c *Cluster) RemoveDomainRule(ctx context.Context, opts RemoveDomainRuleOptions) map[int64]*domain.NodeResult[struct{}] {
	c.logger.Debug().Msg("removing domain rule from all pihole nodes")
	out, _ := fanout[struct{}](c, ctx, 0, func(nodeCtx context.Context, _ int64, client clientPort) (struct{}, error) {
		return struct{}{}, client.RemoveDomainRule(nodeCtx, opts)
	})
	return out
}

func (c *Cluster) AuthStatus(ctx context.Context) map[int64]*domain.NodeResult[*domain.AuthStatus] {
	c.logger.Trace().Msg("getting auth status for cluster")
	out, _ := fanout[*domain.AuthStatus](c, ctx, 0, func(nodeCtx context.Context, _ int64, client clientPort) (*domain.AuthStatus, error) {
		return client.AuthStatus(nodeCtx)
	})
	return out
}

func (c *Cluster) Logout(ctx context.Context) map[int64]*domain.NodeResult[struct{}] {
	c.logger.Debug().Msg("logging out all pihole nodes")
	out, _ := fanout[struct{}](c, ctx, 0, func(nodeCtx context.Context, _ int64, client clientPort) (struct{}, error) {
		return struct{}{}, client.Logout(nodeCtx)
	})
	return out
}

func fanout[T any](
	c *Cluster,
	ctx context.Context,
	limit int,
	op func(ctx context.Context, id int64, client clientPort) (T, error),
) (NodeResults[T], error) {
	c.logger.Debug().Msg("fanout: starting operation on all pihole nodes")

	c.rw.RLock()
	clients := make(map[int64]clientPort, len(c.clients))
	for id, cl := range c.clients {
		clients[id] = cl
	}
	c.rw.RUnlock()

	results := make(NodeResults[T], len(clients))
	var mu sync.Mutex

	var sem chan struct{}
	if limit > 0 {
		sem = make(chan struct{}, limit)
	}

	g, gctx := errgroup.WithContext(ctx)
	for id, cl := range clients {
		id, cl := id, cl
		g.Go(func() error {
			defer func() {
				if r := recover(); r != nil {
					c.logger.Error().Interface("panic", r).Int64("id", id).Msg("worker panic recovered")
				}
			}()

			if sem != nil {
				select {
				case sem <- struct{}{}:
				case <-gctx.Done():
					return gctx.Err()
				}
				defer func() { <-sem }()
			}

			nodeTimeout := 3 * time.Second
			if dl, ok := gctx.Deadline(); ok {
				nodeTimeout = time.Until(dl) / 2
				if nodeTimeout > 3*time.Second {
					nodeTimeout = 3 * time.Second
				}
			}
			nodeCtx, cancel := context.WithTimeout(gctx, nodeTimeout)
			defer cancel()

			resp, err := op(nodeCtx, id, cl)

			var nerr *domain.NodeError
			if err != nil {
				nerr = mapClientErr(err)
			}

			node := cl.GetNodeInfo(nodeCtx)

			mu.Lock()
			results[id] = &domain.NodeResult[T]{
				PiholeNode: node,
				Success:    err == nil,
				Error:      err,
				NodeErr:    nerr,
				Response:   resp,
			}
			mu.Unlock()

			if err != nil && !errors.Is(err, context.Canceled) {
				c.logger.Warn().Int64("id", id).Str("error", util.ErrorString(err)).Msg("fanout: node operation failed")
			}
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		if errors.Is(err, context.Canceled) {
			c.logger.Warn().Err(err).Msg("fanout aborted")
		}
		return results, err
	}
	return results, nil
}
