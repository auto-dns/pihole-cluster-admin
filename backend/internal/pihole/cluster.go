package pihole

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/auto-dns/pihole-cluster-admin/internal/domain"
	"github.com/auto-dns/pihole-cluster-admin/internal/errs"
	logs "github.com/auto-dns/pihole-cluster-admin/internal/logger"
	"github.com/auto-dns/pihole-cluster-admin/internal/util"
	"github.com/rs/zerolog"
	"golang.org/x/sync/errgroup"
)

type Cluster struct {
	clients       map[int64]clientPort
	cursorManager cursorManagerPort[domain.QueryLogFilters]
	logger        zerolog.Logger
	rw            sync.RWMutex
}

func NewCluster(clientMap map[int64]*Client, cursorManager cursorManagerPort[domain.QueryLogFilters], logger zerolog.Logger) *Cluster {
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
	logs.Event(ctx, c.logger).Msg("getting block status from all pihole nodes")
	out, _ := fanout(c, ctx, 0, 3*time.Second, func(nodeCtx context.Context, _ int64, client clientPort) (*domain.BlockingState, error) {
		return client.GetBlockingState(nodeCtx)
	})
	return out
}

func (c *Cluster) SetBlockingState(ctx context.Context, blocking bool, timer *int) map[int64]*domain.NodeResult[*domain.BlockingState] {
	c.logger.Debug().Msg("setting block status on all pihole nodes")
	out, _ := fanout(c, ctx, 0, 3*time.Second, func(nodeCtx context.Context, _ int64, client clientPort) (*domain.BlockingState, error) {
		return client.SetBlockingState(nodeCtx, blocking, timer)
	})
	return out
}

func (c *Cluster) SetBlockingStateForNode(ctx context.Context, nodeID int64, blocking bool, timer *int) (*domain.NodeResult[*domain.BlockingState], error) {
	c.rw.RLock()
	client, exists := c.clients[nodeID]
	c.rw.RUnlock()
	if !exists {
		return nil, errs.NotFound("node not found", fmt.Errorf("node %d not found", nodeID))
	}
	nodeTimeout := 3 * time.Second
	if dl, ok := ctx.Deadline(); ok {
		nodeTimeout = time.Until(dl) / 2
		if nodeTimeout > 3*time.Second {
			nodeTimeout = 3 * time.Second
		}
	}
	nodeCtx, cancel := context.WithTimeout(ctx, nodeTimeout)
	defer cancel()
	resp, err := client.SetBlockingState(nodeCtx, blocking, timer)
	if err != nil {
		err = mapClientErr(err)
	}
	node := client.GetNodeInfo(nodeCtx)
	return &domain.NodeResult[*domain.BlockingState]{
		PiholeNode: node,
		Success:    err == nil,
		Error:      err,
		Response:   resp,
	}, nil
}

// Query Logs

func domainFiltersToWire(f domain.QueryLogFilters) queriesWireFilters {
	var from, until *int64
	if f.From != nil {
		v := f.From.Unix()
		from = &v
	}
	if f.Until != nil {
		v := f.Until.Unix()
		until = &v
	}
	return queriesWireFilters{
		From: from, Until: until,
		Domain: f.Domain, ClientIP: f.ClientIP, ClientName: f.ClientName,
		Upstream: f.Upstream, Type: f.Type, Status: f.Status,
		Reply: f.Reply, DNSSEC: f.DNSSEC, Disk: f.Disk,
	}
}

func (c *Cluster) FetchQueryLogs(ctx context.Context, req domain.QueryLogQuery) (*domain.ClusterQueryLogResponse, error) {
	c.logger.Debug().Msg("fetching query logs from all pihole nodes")

	// Get search state (from cursor, if present - nil if not)
	var state searchStatePort[domain.QueryLogFilters]
	if req.Cursor != nil && *req.Cursor != "" {
		var ok bool
		if state, ok = c.cursorManager.GetSearchState(*req.Cursor); !ok {
			return nil, fmt.Errorf("cursor expired or not found")
		}
	}

	// Create result map
	responses, err := fanout(c, ctx, 0, 12*time.Second, func(nodeCtx context.Context, id int64, client clientPort) (*domain.QueryLogPage, error) {
		// Build request
		wireReq := queriesWireRequest{
			Filters: domainFiltersToWire(req.Filters),
			Length:  req.Length,
			Start:   req.Start,
		}
		// Add cursor to individual pihole request if one exists
		if state != nil {
			if cursor, ok := state.GetPiholeCursor(id); ok {
				wireReq.Filters = domainFiltersToWire(state.GetRequestParams())
				wireReq.Start = nil
				wireReq.Cursor = &cursor
			} else {
				c.logger.Warn().Int64("id", id).Msg("pihole cursor not found")
			}
		}

		// Make pihole client request
		return client.FetchQueryLogs(nodeCtx, wireReq)
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
		return &domain.ClusterQueryLogResponse{
			Cursor:       *req.Cursor,
			Results:      responses,
			EndOfResults: true,
		}, nil
	}

	// All successful nodes returned cursor=0 → no more data (catches zero-result first pages too)
	allDone := true
	for _, cursor := range nextPiholeCursors {
		if cursor != 0 {
			allDone = false
			break
		}
	}

	newCursor := c.cursorManager.CreateCursor(req.Filters, nextPiholeCursors)
	return &domain.ClusterQueryLogResponse{
		Cursor:       newCursor,
		Results:      responses,
		EndOfResults: allDone,
	}, nil
}

func (c *Cluster) ListDomainRules(ctx context.Context, q domain.ListDomainRulesQuery) map[int64]*domain.NodeResult[*domain.DomainRuleSet] {
	c.logger.Debug().Msg("getting domain rules from all pihole nodes")
	out, _ := fanout(c, ctx, 0, 3*time.Second, func(nodeCtx context.Context, _ int64, client clientPort) (*domain.DomainRuleSet, error) {
		return client.ListDomainRules(nodeCtx, q)
	})
	return out
}

func (c *Cluster) AddDomainRule(ctx context.Context, cmd domain.AddDomainRulesCommand) map[int64]*domain.NodeResult[*domain.AddDomainRulesResult] {
	c.logger.Debug().Msg("adding domain rule to all pihole nodes")
	out, _ := fanout(c, ctx, 0, 3*time.Second, func(nodeCtx context.Context, _ int64, client clientPort) (*domain.AddDomainRulesResult, error) {
		return client.AddDomainRule(nodeCtx, cmd)
	})
	return out
}

func (c *Cluster) RemoveDomainRule(ctx context.Context, cmd domain.RemoveDomainRuleCommand) map[int64]*domain.NodeResult[struct{}] {
	c.logger.Debug().Msg("removing domain rule from all pihole nodes")
	out, _ := fanout(c, ctx, 0, 3*time.Second, func(nodeCtx context.Context, _ int64, client clientPort) (struct{}, error) {
		return struct{}{}, client.RemoveDomainRule(nodeCtx, cmd)
	})
	return out
}

func (c *Cluster) AddDomainRuleToNodes(ctx context.Context, cmd domain.AddDomainRulesCommand, nodeIDs []int64) map[int64]*domain.NodeResult[*domain.AddDomainRulesResult] {
	targets := c.filterClients(nodeIDs)
	out := make(map[int64]*domain.NodeResult[*domain.AddDomainRulesResult], len(targets))
	var mu sync.Mutex
	g, gctx := errgroup.WithContext(ctx)
	for id, cl := range targets {
		id, cl := id, cl
		g.Go(func() error {
			nodeCtx, cancel := context.WithTimeout(gctx, 3*time.Second)
			defer cancel()
			resp, err := cl.AddDomainRule(nodeCtx, cmd)
			if err != nil {
				err = mapClientErr(err)
			}
			mu.Lock()
			out[id] = &domain.NodeResult[*domain.AddDomainRulesResult]{
				PiholeNode: cl.GetNodeInfo(nodeCtx),
				Success:    err == nil,
				Error:      err,
				Response:   resp,
			}
			mu.Unlock()
			return nil
		})
	}
	_ = g.Wait()
	return out
}

func (c *Cluster) RemoveDomainRuleFromNodes(ctx context.Context, cmd domain.RemoveDomainRuleCommand, nodeIDs []int64) map[int64]*domain.NodeResult[struct{}] {
	targets := c.filterClients(nodeIDs)
	out := make(map[int64]*domain.NodeResult[struct{}], len(targets))
	var mu sync.Mutex
	g, gctx := errgroup.WithContext(ctx)
	for id, cl := range targets {
		id, cl := id, cl
		g.Go(func() error {
			nodeCtx, cancel := context.WithTimeout(gctx, 3*time.Second)
			defer cancel()
			err := cl.RemoveDomainRule(nodeCtx, cmd)
			if err != nil {
				err = mapClientErr(err)
			}
			mu.Lock()
			out[id] = &domain.NodeResult[struct{}]{
				PiholeNode: cl.GetNodeInfo(nodeCtx),
				Success:    err == nil,
				Error:      err,
			}
			mu.Unlock()
			return nil
		})
	}
	_ = g.Wait()
	return out
}

func (c *Cluster) filterClients(nodeIDs []int64) map[int64]clientPort {
	c.rw.RLock()
	defer c.rw.RUnlock()
	targets := make(map[int64]clientPort, len(nodeIDs))
	for _, id := range nodeIDs {
		if cl, ok := c.clients[id]; ok {
			targets[id] = cl
		}
	}
	return targets
}

// Stats

func (c *Cluster) GetStatsSummary(ctx context.Context) map[int64]*domain.NodeResult[*domain.StatsSummary] {
	out, _ := fanout(c, ctx, 0, 3*time.Second, func(nodeCtx context.Context, _ int64, client clientPort) (*domain.StatsSummary, error) {
		return client.GetStatsSummary(nodeCtx)
	})
	return out
}

func (c *Cluster) GetStatsHistory(ctx context.Context, from, until *int64) map[int64]*domain.NodeResult[*domain.StatsHistory] {
	out, _ := fanout(c, ctx, 0, 5*time.Second, func(nodeCtx context.Context, _ int64, client clientPort) (*domain.StatsHistory, error) {
		return client.GetStatsHistory(nodeCtx, from, until)
	})
	return out
}

func (c *Cluster) GetStatsTopDomains(ctx context.Context, from, until *int64, count *int) map[int64]*domain.NodeResult[*domain.StatsTopDomains] {
	out, _ := fanout(c, ctx, 0, 3*time.Second, func(nodeCtx context.Context, _ int64, client clientPort) (*domain.StatsTopDomains, error) {
		return client.GetStatsTopDomains(nodeCtx, from, until, count)
	})
	return out
}

func (c *Cluster) GetStatsTopClients(ctx context.Context, from, until *int64, count *int) map[int64]*domain.NodeResult[*domain.StatsTopClients] {
	out, _ := fanout(c, ctx, 0, 3*time.Second, func(nodeCtx context.Context, _ int64, client clientPort) (*domain.StatsTopClients, error) {
		return client.GetStatsTopClients(nodeCtx, from, until, count)
	})
	return out
}

func (c *Cluster) AuthStatus(ctx context.Context) map[int64]*domain.NodeResult[*domain.AuthStatus] {
	c.logger.Trace().Msg("getting auth status for cluster")
	out, _ := fanout(c, ctx, 0, 3*time.Second, func(nodeCtx context.Context, _ int64, client clientPort) (*domain.AuthStatus, error) {
		return client.AuthStatus(nodeCtx)
	})
	return out
}

func (c *Cluster) Logout(ctx context.Context) map[int64]*domain.NodeResult[struct{}] {
	c.logger.Debug().Msg("logging out all pihole nodes")
	out, _ := fanout(c, ctx, 0, 3*time.Second, func(nodeCtx context.Context, _ int64, client clientPort) (struct{}, error) {
		return struct{}{}, client.Logout(nodeCtx)
	})
	return out
}

func fanout[T any](
	c *Cluster,
	ctx context.Context,
	limit int,
	maxNodeTimeout time.Duration,
	op func(ctx context.Context, id int64, client clientPort) (T, error),
) (map[int64]*domain.NodeResult[T], error) {
	logs.Event(ctx, c.logger).Msg("fanout: starting operation on all pihole nodes")

	c.rw.RLock()
	clients := make(map[int64]clientPort, len(c.clients))
	for id, cl := range c.clients {
		clients[id] = cl
	}
	c.rw.RUnlock()

	results := make(map[int64]*domain.NodeResult[T], len(clients))
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

			nodeTimeout := maxNodeTimeout
			if dl, ok := gctx.Deadline(); ok {
				if remaining := time.Until(dl) / 2; remaining < nodeTimeout {
					nodeTimeout = remaining
				}
			}
			nodeCtx, cancel := context.WithTimeout(gctx, nodeTimeout)
			defer cancel()

			resp, err := op(nodeCtx, id, cl)

			if err != nil {
				err = mapClientErr(err)
			}

			node := cl.GetNodeInfo(nodeCtx)

			mu.Lock()
			results[id] = &domain.NodeResult[T]{
				PiholeNode: node,
				Success:    err == nil,
				Error:      err,
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
