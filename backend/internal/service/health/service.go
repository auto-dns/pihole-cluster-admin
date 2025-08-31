package healthsvc

import (
	"context"
	"sync"
	"time"

	"github.com/auto-dns/pihole-cluster-admin/internal/config"
	"github.com/auto-dns/pihole-cluster-admin/internal/domain"
	"github.com/auto-dns/pihole-cluster-admin/internal/logger"
	"github.com/auto-dns/pihole-cluster-admin/internal/poller"
	"github.com/auto-dns/pihole-cluster-admin/internal/realtime"
	"github.com/rs/zerolog"
)

type Service struct {
	broker     broker
	cluster    cluster
	logger     zerolog.Logger
	cfg        config.PublisherConfig
	mu         sync.RWMutex
	nodeHealth map[int64]nodeHealth
	summary    summary
}

func NewService(broker broker, cluster cluster, cfg config.PublisherConfig, logger zerolog.Logger) *Service {
	return &Service{
		broker:     broker,
		cluster:    cluster,
		cfg:        cfg,
		logger:     logger,
		nodeHealth: make(map[int64]nodeHealth),
	}
}

func (s *Service) StartPublisher(ctx context.Context) {
	s.logger.Info().Msg("Starting health service")
	s.sweepOnce(ctx)
	p := poller.New(s.broker, poller.Config{
		Interval:    time.Duration(max(1, s.cfg.PollingIntervalSeconds)) * time.Second,
		GracePeriod: time.Duration(s.cfg.GracePeriodSeconds) * time.Second,
		JitterRatio: 0.20,
	})
	go p.Run(ctx, s.sweepOnce)
}

func (s *Service) GetSummary() domain.ClusterHealthSummary {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// TODO: change this to call cluster directly
	return ToDomainHealthSummary(s.summary)
}

func (s *Service) GetNodeHealth() map[int64]domain.ClusterNodeHealth {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// TODO: change this to call cluster directly
	res := make(map[int64]domain.ClusterNodeHealth, len(s.nodeHealth))
	for id, nh := range s.nodeHealth {
		res[id] = ToDomainNodeHealth(nh)
	}

	return res
}

func (s *Service) sweepOnce(ctx context.Context) {
	pollLog := s.logger.With().Str("component", "health").Logger()
	ctx = logger.WithContext(ctx, pollLog)
	ctx = logger.WithMode(ctx, logger.ModeTrace)

	results := s.cluster.AuthStatus(ctx)

	now := time.Now()

	s.mu.Lock()
	defer s.mu.Unlock()

	for _, r := range results {
		var tookMs int
		var valid bool

		if r.Error == nil {
			tookMs = int(r.Response.Took.Milliseconds())
			valid = r.Response.Valid
		}

		nodeHealth := nodeHealth{
			Id:        r.PiholeNode.Id,
			Name:      r.PiholeNode.Name,
			Status:    pickStatus(valid, r.Error),
			Latency:   time.Duration(tookMs * int(time.Millisecond)),
			UpdatedAt: now,
		}
		if r.Error != nil {
			nodeHealth.LastErr = r.Error.Error()
		}
		s.nodeHealth[r.PiholeNode.Id] = nodeHealth
	}
	s.recomputeLocked()
}

func pickStatus(valid bool, err error) status {
	switch {
	case err != nil:
		return statusOffline
	case valid:
		return statusOnline
	default:
		return statusDegraded
	}
}

func (s *Service) recomputeLocked() {
	s.logger.Trace().Msg("recomputing summary")
	online := 0

	for _, nodeHealth := range s.nodeHealth {
		if nodeHealth.Status == statusOnline {
			online++
		}
	}
	s.summary = summary{
		Online:    online,
		Total:     len(s.nodeHealth),
		UpdatedAt: time.Now(),
	}
	s.logger.Trace().Int("online", online).Int("total", len(s.nodeHealth)).Time("updated_at", s.summary.UpdatedAt).Msg("summary recomputed")
	s.broker.Publish(realtime.TopicHealthSummaryV1, ToDomainHealthSummary(s.summary))

	list := make([]domain.ClusterNodeHealth, 0, len(s.nodeHealth))
	for _, nh := range s.nodeHealth {
		list = append(list, ToDomainNodeHealth(nh))
	}
	s.broker.Publish(realtime.TopicNodeHealthV1, list)
}

func randInt63n(n int64) int64 {
	return time.Now().UnixNano() % n
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
