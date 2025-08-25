package clusterblocking

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

func (s *Service) GetState(ctx context.Context) (*ClusterBlockingState, error) {
	nodes := s.cluster.GetBlockingState(ctx)
	summary := BlockingSummary{Total: len(nodes)}
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
		summary.MinTimer, summary.MaxTimer = &timers[0], &timers[len(timers)-1]
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

	return &ClusterBlockingState{
		Summary: summary,
		Nodes:   nodes,
	}, nil
}
