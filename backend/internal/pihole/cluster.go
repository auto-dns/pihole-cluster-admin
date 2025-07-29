package pihole

import (
	"fmt"
	"sync"

	"github.com/auto-dns/pihole-cluster-admin/internal/util"
)

type Cluster struct {
	clients       []ClientInterface
	cursorManager *CursorManager[FetchQueryLogFilters]
}

func NewCluster(clients ...ClientInterface) *Cluster {
	return &Cluster{
		clients: clients,
		cursorManager: &CursorManager[FetchQueryLogFilters]{
			cursors: make(map[string]*CursorState[FetchQueryLogFilters]),
		},
	}
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
	var wg sync.WaitGroup

	for i, client := range c.clients {
		wg.Add(1)
		go func(i int, ci ClientInterface) {
			defer wg.Done()

			nodeReq := FetchQueryLogRequest{
				Filters:  filters, // use either user-provided filters (no cursor) or cursor snapshot
				Length:   req.Length,
				Start:    req.Start,
				CursorID: nil,
			}

			// If we already have a node-specific cursor, use it
			if cursor, ok := nodeCursors[ci.GetNodeInfo().ID]; ok {
				nodeReq.CursorID = &cursor
				nodeReq.Start = nil // offset is ignored when using cursor
			}

			res, err := ci.FetchQueryLogs(nodeReq)
			node := ci.GetNodeInfo()
			results[i] = &NodeResult[FetchQueryLogResponse]{
				PiholeNode: node,
				Success:    err == nil,
				Error:      util.ErrorString(err),
				Response:   res,
			}

			// Save the node cursor for future pagination
			if err == nil && res != nil {
				nodeCursors[node.ID] = fmt.Sprintf("%d", res.Cursor)
			}
		}(i, client)
	}

	wg.Wait()

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
	var wg sync.WaitGroup
	results := make([]*NodeResult[GetDomainRulesResponse], len(c.clients))

	for i, client := range c.clients {
		wg.Add(1)
		go func(i int, ci ClientInterface) {
			defer wg.Done()
			res, err := ci.GetDomainRules(opts)
			node := ci.GetNodeInfo()
			results[i] = &NodeResult[GetDomainRulesResponse]{
				PiholeNode: node,
				Success:    err == nil,
				Error:      util.ErrorString(err),
				Response:   res,
			}
		}(i, client)
	}

	wg.Wait()
	return results

}

func (c *Cluster) AddDomainRule(opts AddDomainRuleOptions) []*NodeResult[AddDomainRuleResponse] {
	var wg sync.WaitGroup
	results := make([]*NodeResult[AddDomainRuleResponse], len(c.clients))

	for i, client := range c.clients {
		wg.Add(1)
		go func(i int, ci ClientInterface) {
			defer wg.Done()
			r, err := ci.AddDomainRule(opts)
			node := ci.GetNodeInfo()
			results[i] = &NodeResult[AddDomainRuleResponse]{
				PiholeNode: node,
				Success:    err == nil,
				Error:      util.ErrorString(err),
				Response:   r,
			}
		}(i, client)
	}

	wg.Wait()
	return results
}

func (c *Cluster) RemoveDomainRule(opts RemoveDomainRuleOptions) []*NodeResult[RemoveDomainRuleResponse] {
	var wg sync.WaitGroup
	results := make([]*NodeResult[RemoveDomainRuleResponse], len(c.clients))

	for i, client := range c.clients {
		wg.Add(1)
		go func(i int, ci ClientInterface) {
			defer wg.Done()
			err := ci.RemoveDomainRule(opts)
			node := ci.GetNodeInfo()

			results[i] = &NodeResult[RemoveDomainRuleResponse]{
				PiholeNode: node,
				Success:    err == nil,
				Error:      util.ErrorString(err),
				// Response is nil because 204 has no body
			}
		}(i, client)
	}

	wg.Wait()
	return results
}

func (c *Cluster) Logout() []error {
	var wg sync.WaitGroup
	errs := make([]error, len(c.clients))

	for i, client := range c.clients {
		wg.Add(1)
		go func(i int, ci ClientInterface) {
			defer wg.Done()
			err := ci.Logout()
			errs[i] = err
		}(i, client)
	}

	wg.Wait()
	return errs
}
