package clusterblockingsvc

import (
	"context"
	"sort"
	"time"

	"github.com/auto-dns/pihole-cluster-admin/internal/domain"
)

type Service struct {
	cluster cluster
}

func NewService(cluster cluster) *Service {
	return &Service{
		cluster: cluster,
	}
}

func (s *Service) GetState(ctx context.Context) (*domain.ClusterBlockingState, error) {
	nodes := s.cluster.GetBlockingSummary(ctx)
	return processClusterResponse(nodes), nil
}

func (s *Service) SetState(ctx context.Context, blocking bool, timer *int) (*domain.ClusterBlockingState, error) {
	nodes := s.cluster.SetBlockingSummary(ctx, blocking, timer)
	return processClusterResponse(nodes), nil
}

func processClusterResponse(nodes map[int64]*domain.NodeResult[*domain.BlockingState]) *domain.ClusterBlockingState {
	summary := domain.BlockingSummary{Total: len(nodes)}
	var timers []time.Duration
	var tookTotal time.Duration

	for _, r := range nodes {
		if !r.Success {
			summary.Errors++
			continue
		}
		switch r.Response.Status {
		case domain.BlockingEnabled:
			summary.Enabled++
		case domain.BlockingDisabled:
			summary.Disabled++
		case domain.BlockingFailed:
			summary.Failed++
		}
		if r.Response.TimerLeft != nil {
			timers = append(timers, *r.Response.TimerLeft)
		}
		if r.Response.Took > summary.MaxTook {
			summary.MaxTook = r.Response.Took
		}
		tookTotal += r.Response.Took
	}

	if summary.Total > 0 {
		summary.AvgTook = tookTotal / time.Duration(summary.Total)
	}
	if len(timers) > 0 {
		sort.Slice(timers, func(i, j int) bool { return timers[i] < timers[j] })
		min := timers[0]
		max := timers[len(timers)-1]
		summary.MinTimer, summary.MaxTimer = &min, &max
	}

	switch {
	case summary.Errors > 0 || summary.Failed > 0:
		summary.Mode = "degraded"
	case summary.Enabled == summary.Total:
		summary.Mode = "enabled"
		summary.Unanimous = true
	case summary.Disabled == summary.Total:
		summary.Mode = "disabled"
		summary.Unanimous = true
	default:
		summary.Mode = "mixed"
	}

	return &domain.ClusterBlockingState{
		Summary: summary,
		Nodes:   nodes,
	}
}
