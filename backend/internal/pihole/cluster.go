package pihole

import (
	"errors"
	"fmt"
	"sync"

	"github.com/auto-dns/pihole-cluster-admin/internal/util"
	"github.com/rs/zerolog"
)

type Cluster struct {
	clients       map[int64]ClientInterface
	cursorManager *CursorManager[FetchQueryLogFilters]
	logger        zerolog.Logger
}

func NewCluster(clients map[int64]ClientInterface, logger zerolog.Logger) ClusterInterface {
	return &Cluster{
		clients: clients,
		cursorManager: &CursorManager[FetchQueryLogFilters]{
			cursors: make(map[string]*CursorState[FetchQueryLogFilters]),
		},
		logger: logger,
	}
}

func (c *Cluster) AddClient(client ClientInterface) error {
	id := client.GetId()
	logger := c.logger.With().Int64("id", id).Str("name", client.GetName()).Str("scheme", client.GetScheme()).Str("host", client.GetHost()).Int("port", client.GetPort()).Logger()

	if _, exists := c.clients[id]; exists {
		err := errors.New("client id already exists")
		logger.Error().Err(err).Msg("client id conflict")
		return err
	}

	c.clients[id] = client
	c.logger.Debug().Msg("client added to cluster")

	return nil
}

func (c *Cluster) RemoveClient(id int64) error {
	logger := c.logger.With().Int64("id", id).Logger()

	client, exists := c.clients[id]
	if !exists {
		err := errors.New("client id not found")
		logger.Error().Err(err).Msg("client id not found")
		return err
	}

	delete(c.clients, id)
	c.logger.Debug().Int64("id", client.GetId()).Str("name", client.GetName()).Str("scheme", client.GetScheme()).Str("host", client.GetHost()).Int("port", client.GetPort()).Msg("client added to cluster")

	return nil
}

func (c *Cluster) UpdateClient(id int64, cfg *ClientConfig) error {
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

	c.clients[id].Update(cfg)
	logger.Debug().Msg("updated client")
	return nil
}

func (c *Cluster) HasClient(id int64) bool {
	_, has := c.clients[id]
	return has
}

func (c *Cluster) forEachClient(f func(id int64, client ClientInterface)) {
	var wg sync.WaitGroup
	for id, client := range c.clients {
		wg.Add(1)
		go func(id int64, client ClientInterface) {
			defer wg.Done()
			defer func() {
				if r := recover(); r != nil {
					c.logger.Error().Interface("panic", r).Int64("id", id).Msg("worker panic recovered")
				}
			}()
			f(id, client)
		}(id, client)
	}
	wg.Wait()
}

func (c *Cluster) FetchQueryLogs(req FetchQueryLogClusterRequest) (*FetchQueryLogsClusterResponse, error) {
	c.logger.Debug().Msg("fetching query logs from all pihole nodes")

	var nodeCursors map[int64]int
	filters := req.Filters

	// --- App cursor is passed in from the browser
	if req.Cursor != nil && *req.Cursor != "" {
		state, ok := c.cursorManager.GetCursor(*req.Cursor)
		if !ok {
			return nil, fmt.Errorf("cursor expired or not found")
		}
		// Grab the cached filters used during the original request from cursor snapshot to ensure consistency
		filters = state.Options
		nodeCursors = state.NodeCursors
	} else {
		// --- Cursor is not passed in - initialize empty cursor list
		nodeCursors = make(map[int64]int)
	}

	var resultsMutex sync.Mutex
	// Prepare results list
	results := make(map[int64]*NodeResult[FetchQueryLogResponse], len(c.clients))

	// Used to collect the cursors from each pihole for later synthesis back into the cursor storage
	newClientCursors := make(map[int64]int)

	// Make a call to each pihole in parallel
	c.forEachClient(func(id int64, client ClientInterface) {
		nodeReq := FetchQueryLogClientRequest{
			Filters: filters, // use either user-provided filters (no cursor) or cursor snapshot
			Length:  req.Length,
			Start:   req.Start,
			Cursor:  nil,
		}

		// If we already have a node-specific cursor, use it
		if cursor, ok := nodeCursors[id]; ok {
			nodeReq.Cursor = &cursor
			nodeReq.Start = nil // offset is ignored when using cursor
		}

		response, err := client.FetchQueryLogs(nodeReq)

		if err != nil {
			c.logger.Warn().Int64("id", id).Str("error", util.ErrorString(err)).Msg("node operation failed")
		}

		node := client.GetNodeInfo()
		resultsMutex.Lock()
		results[id] = &NodeResult[FetchQueryLogResponse]{
			PiholeNode: node,
			Success:    err == nil,
			Error:      util.ErrorString(err),
			Response:   response,
		}

		// Save the node cursor for future pagination
		if err == nil && response != nil {
			newClientCursors[id] = response.Cursor
		}
		resultsMutex.Unlock()
	})

	// Merge local cursors into the shared map
	for id, cursor := range newClientCursors {
		if cursor != 0 {
			nodeCursors[id] = cursor
		}
	}

	// Determine if node cursors changed
	var changed bool
	if req.Cursor == nil || *req.Cursor == "" {
		changed = true // first call always creates new cursor
	} else {
		state, _ := c.cursorManager.GetCursor(*req.Cursor)
		if state != nil {
			for nodeID, cursor := range nodeCursors {
				if state.NodeCursors[nodeID] != cursor {
					changed = true
					break
				}
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

	// Create a new cursor snapshot when data advanced
	newCursor := c.cursorManager.NewCursor(filters, nodeCursors)

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

func (c *Cluster) GetDomainRules(opts GetDomainRulesOptions) map[int64]*NodeResult[GetDomainRulesResponse] {
	c.logger.Debug().Msg("getting domain rules from all pihole nodes")

	results := make(map[int64]*NodeResult[GetDomainRulesResponse], len(c.clients))
	c.forEachClient(func(id int64, client ClientInterface) {
		res, err := client.GetDomainRules(opts)

		if err != nil {
			c.logger.Warn().Int64("id", id).Str("error", util.ErrorString(err)).Msg("node operation failed")
		}

		node := client.GetNodeInfo()
		results[id] = &NodeResult[GetDomainRulesResponse]{
			PiholeNode: node,
			Success:    err == nil,
			Error:      util.ErrorString(err),
			Response:   res,
		}
	})
	return results
}

func (c *Cluster) AddDomainRule(opts AddDomainRuleOptions) map[int64]*NodeResult[AddDomainRuleResponse] {
	c.logger.Debug().Msg("adding domain rule to all pihole nodes")

	results := make(map[int64]*NodeResult[AddDomainRuleResponse], len(c.clients))
	c.forEachClient(func(id int64, client ClientInterface) {
		r, err := client.AddDomainRule(opts)

		if err != nil {
			c.logger.Warn().Int64("id", id).Str("error", util.ErrorString(err)).Msg("node operation failed")
		}

		node := client.GetNodeInfo()
		results[id] = &NodeResult[AddDomainRuleResponse]{
			PiholeNode: node,
			Success:    err == nil,
			Error:      util.ErrorString(err),
			Response:   r,
		}
	})
	return results
}

func (c *Cluster) RemoveDomainRule(opts RemoveDomainRuleOptions) map[int64]*NodeResult[RemoveDomainRuleResponse] {
	c.logger.Debug().Msg("removing domain rule from all pihole nodes")

	results := make(map[int64]*NodeResult[RemoveDomainRuleResponse], len(c.clients))
	c.forEachClient(func(id int64, client ClientInterface) {
		err := client.RemoveDomainRule(opts)

		if err != nil {
			c.logger.Warn().Int64("id", id).Str("error", util.ErrorString(err)).Msg("node operation failed")
		}

		node := client.GetNodeInfo()
		results[id] = &NodeResult[RemoveDomainRuleResponse]{
			PiholeNode: node,
			Success:    err == nil,
			Error:      util.ErrorString(err),
			// Response is nil because 204 has no body
		}
	})
	return results
}

func (c *Cluster) Logout() map[int64]error {
	c.logger.Debug().Msg("logging out all pihole nodes")

	errs := make(map[int64]error, len(c.clients))
	c.forEachClient(func(id int64, client ClientInterface) {
		err := client.Logout()

		if err != nil {
			c.logger.Warn().Int64("id", id).Str("error", util.ErrorString(err)).Msg("node operation failed")
		}

		errs[id] = err
	})
	return errs
}
