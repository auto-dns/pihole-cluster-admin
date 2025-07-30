package pihole

import (
	"fmt"
	"sync"

	"github.com/auto-dns/pihole-cluster-admin/internal/util"
	"github.com/rs/zerolog"
)

type Cluster struct {
	clients       []ClientInterface
	cursorManager *CursorManager[FetchQueryLogFilters]
	logger        zerolog.Logger
}

func NewCluster(logger zerolog.Logger, clients ...ClientInterface) *Cluster {
	return &Cluster{
		clients: clients,
		cursorManager: &CursorManager[FetchQueryLogFilters]{
			cursors: make(map[string]*CursorState[FetchQueryLogFilters]),
		},
		logger: logger,
	}
}

func (c *Cluster) forEachClient(f func(i int, client ClientInterface)) {
	var wg sync.WaitGroup
	for i, client := range c.clients {
		wg.Add(1)
		go func(i int, client ClientInterface) {
			defer wg.Done()
			defer func() {
				if r := recover(); r != nil {
					c.logger.Error().Interface("panic", r).Msg("worker panic recovered")
				}
			}()
			f(i, client)
		}(i, client)
	}
	wg.Wait()
}

func (c *Cluster) FetchQueryLogs(req FetchQueryLogRequest) (FetchQueryLogsClusterResponse, error) {
	var nodeCursors map[string]string
	filters := req.Filters

	// --- Handle cursor reuse
	if req.CursorID != nil && *req.CursorID != "" {
		state, ok := c.cursorManager.GetCursor(*req.CursorID)
		if !ok {
			return FetchQueryLogsClusterResponse{}, fmt.Errorf("cursor expired or not found")
		}
		// Reuse filters from cursor snapshot to ensure consistency
		filters = state.Options
		nodeCursors = state.NodeCursors
	} else {
		nodeCursors = make(map[string]string)
	}

	results := make([]*NodeResult[FetchQueryLogResponse], len(c.clients))
	type cursorResult struct {
		nodeId string
		cursor string
	}
	cursorsOut := make([]cursorResult, len(c.clients))
	c.forEachClient(func(i int, client ClientInterface) {
		nodeReq := FetchQueryLogRequest{
			Filters:  filters, // use either user-provided filters (no cursor) or cursor snapshot
			Length:   req.Length,
			Start:    req.Start,
			CursorID: nil,
		}

		// If we already have a node-specific cursor, use it
		if cursor, ok := nodeCursors[client.GetNodeInfo().ID]; ok {
			nodeReq.CursorID = &cursor
			nodeReq.Start = nil // offset is ignored when using cursor
		}

		res, err := client.FetchQueryLogs(nodeReq)
		node := client.GetNodeInfo()
		results[i] = &NodeResult[FetchQueryLogResponse]{
			PiholeNode: node,
			Success:    err == nil,
			Error:      util.ErrorString(err),
			Response:   res,
		}

		// Save the node cursor for future pagination
		if err == nil && res != nil {
			cursorsOut[i] = cursorResult{
				nodeId: node.ID,
				cursor: fmt.Sprintf("%d", res.Cursor),
			}
		}
	})

	// Merge local cursors into the shared map
	for _, c := range cursorsOut {
		if c.nodeId != "" {
			nodeCursors[c.nodeId] = c.cursor
		}
	}

	// Determine if node cursors changed
	var changed bool
	if req.CursorID == nil || *req.CursorID == "" {
		changed = true // first call always creates new cursor
	} else {
		state, _ := c.cursorManager.GetCursor(*req.CursorID)
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
	if !changed && req.CursorID != nil {
		return FetchQueryLogsClusterResponse{
			CursorID:     *req.CursorID,
			Results:      results,
			EndOfResults: true,
		}, nil
	}

	// Create a new cursor snapshot when data advanced
	newCursor := c.cursorManager.NewCursor(filters, nodeCursors)
	return FetchQueryLogsClusterResponse{
		CursorID:     newCursor,
		Results:      results,
		EndOfResults: false,
	}, nil
}

func (c *Cluster) GetDomainRules(opts GetDomainRulesOptions) []*NodeResult[GetDomainRulesResponse] {
	results := make([]*NodeResult[GetDomainRulesResponse], len(c.clients))
	c.forEachClient(func(i int, client ClientInterface) {
		res, err := client.GetDomainRules(opts)
		node := client.GetNodeInfo()
		results[i] = &NodeResult[GetDomainRulesResponse]{
			PiholeNode: node,
			Success:    err == nil,
			Error:      util.ErrorString(err),
			Response:   res,
		}
	})
	return results
}

func (c *Cluster) AddDomainRule(opts AddDomainRuleOptions) []*NodeResult[AddDomainRuleResponse] {
	results := make([]*NodeResult[AddDomainRuleResponse], len(c.clients))
	c.forEachClient(func(i int, client ClientInterface) {
		r, err := client.AddDomainRule(opts)
		node := client.GetNodeInfo()
		results[i] = &NodeResult[AddDomainRuleResponse]{
			PiholeNode: node,
			Success:    err == nil,
			Error:      util.ErrorString(err),
			Response:   r,
		}
	})
	return results
}

func (c *Cluster) RemoveDomainRule(opts RemoveDomainRuleOptions) []*NodeResult[RemoveDomainRuleResponse] {
	results := make([]*NodeResult[RemoveDomainRuleResponse], len(c.clients))
	c.forEachClient(func(i int, client ClientInterface) {
		err := client.RemoveDomainRule(opts)
		node := client.GetNodeInfo()

		results[i] = &NodeResult[RemoveDomainRuleResponse]{
			PiholeNode: node,
			Success:    err == nil,
			Error:      util.ErrorString(err),
			// Response is nil because 204 has no body
		}
	})
	return results
}

func (c *Cluster) Logout() []error {
	errs := make([]error, len(c.clients))
	c.forEachClient(func(i int, client ClientInterface) {
		err := client.Logout()
		errs[i] = err
	})
	return errs
}
