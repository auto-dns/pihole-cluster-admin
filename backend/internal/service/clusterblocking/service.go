package clusterblockingsvc

import (
	"context"
	"sort"
	"time"

	"github.com/auto-dns/pihole-cluster-admin/internal/config"
	"github.com/auto-dns/pihole-cluster-admin/internal/domain"
	"github.com/auto-dns/pihole-cluster-admin/internal/logger"
	"github.com/auto-dns/pihole-cluster-admin/internal/poller"
	"github.com/auto-dns/pihole-cluster-admin/internal/realtime"
	"github.com/rs/zerolog"
)

type Service struct {
	cluster cluster
	broker  broker
	cfg     config.PublisherConfig
	logger  zerolog.Logger
}

func NewService(cluster cluster, broker broker, cfg config.PublisherConfig, logger zerolog.Logger) *Service {
	return &Service{
		cluster: cluster,
		broker:  broker,
		cfg:     cfg,
		logger:  logger,
	}
}

func (s *Service) GetState(ctx context.Context) (*domain.ClusterBlockingState, error) {
	nodes := s.cluster.GetBlockingState(ctx)
	return processClusterResponse(nodes), nil
}

func (s *Service) SetState(ctx context.Context, blocking bool, timer *int) (*domain.ClusterBlockingState, error) {
	nodes := s.cluster.SetBlockingState(ctx, blocking, timer)
	state := processClusterResponse(nodes)
	s.broker.Publish(realtime.TopicClusterBlockingV1, state)
	return state, nil
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

func (s *Service) StartPublisher(ctx context.Context) {
	s.logger.Info().Msg("Starting cluster blocking service")
	s.getAndPublish(ctx)
	p := poller.New(s.broker, poller.Config{
		Interval:    time.Duration(max(1, s.cfg.PollingIntervalSeconds)) * time.Second,
		GracePeriod: time.Duration(s.cfg.GracePeriodSeconds) * time.Second,
		JitterRatio: 0.20,
	})
	go p.Run(ctx, s.getAndPublish)
}

func (s *Service) getAndPublish(ctx context.Context) {
	ctx = logger.WithMode(ctx, logger.ModeTrace)
	summary := s.cluster.GetBlockingState(ctx)
	s.broker.Publish(realtime.TopicClusterBlockingV1, summary)
}
