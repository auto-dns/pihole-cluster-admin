package health

import (
	"context"
	"encoding/json"
	"math"
	"sync"
	"time"

	"github.com/auto-dns/pihole-cluster-admin/internal/config"
	"github.com/rs/zerolog"
)

type Service struct {
	mu         sync.RWMutex
	nodeHealth map[int64]NodeHealth
	summary    Summary
	broker     broker
	cluster    piholeCluster
	cfg        config.HealthServiceConfig
	logger     zerolog.Logger
}

func NewService(cluster piholeCluster, broker broker, cfg config.HealthServiceConfig, logger zerolog.Logger) *Service {
	return &Service{
		nodeHealth: make(map[int64]NodeHealth),
		broker:     broker,
		cluster:    cluster,
		cfg:        cfg,
		logger:     logger,
	}
}

func (s *Service) Start(ctx context.Context) {
	s.logger.Info().Msg("Starting health service")

	s.sweepOnce(ctx)

	go s.loop(ctx)
}

func (s *Service) loop(ctx context.Context) {
	activeInterval := time.Duration(max(1, s.cfg.PollingIntervalSeconds)) * time.Second
	stopGracePeriod := time.Duration(s.cfg.GracePeriodSeconds) * time.Second

	newTicker := func(d time.Duration) *time.Ticker {
		return time.NewTicker(jitter(d))
	}

	for {
		// Idle state when no subscribers
		if s.broker.SubscriberCount() == 0 {
			select {
			case <-s.broker.SubscribersChanged():
				if s.broker.SubscriberCount() == 0 {
					continue
				}
			case <-ctx.Done():
				return
			}
		}

		// Active: at least one subscriber
		s.sweepOnce(ctx)
		ticker := newTicker(activeInterval)
		running := true

		for running {
			select {
			case <-ticker.C:
				s.sweepOnce(ctx)

			case <-s.broker.SubscribersChanged():
				if s.broker.SubscriberCount() == 0 {
					if stopGracePeriod > 0 {
						g := time.NewTimer(stopGracePeriod)
						defer g.Stop()
						select {
						case <-g.C:
							if s.broker.SubscriberCount() == 0 {
								running = false
							}
						case <-s.broker.SubscribersChanged():
							// Subscribers came back during grace period, keep runnign
						case <-ctx.Done():
							ticker.Stop()
							return
						}
					} else {
						running = false
					}
				}

			case <-ctx.Done():
				ticker.Stop()
				return
			}
		}

		ticker.Stop()
	}
}

func (s *Service) sweepOnce(ctx context.Context) {
	results := s.cluster.AuthStatus(ctx)

	now := time.Now()

	s.mu.Lock()
	defer s.mu.Unlock()

	for _, r := range results {
		var tookMs int
		var valid bool

		if r.Response != nil {
			tookMs = int(math.Round(r.Response.Took * 1000)) // server processing
			valid = r.Response.Session.Valid
		}

		nodeHealth := NodeHealth{
			Id:        r.PiholeNode.Id,
			Name:      r.PiholeNode.Name,
			Status:    pickStatus(valid, r.Error),
			LatencyMS: tookMs,
			UpdatedAt: now,
		}
		if r.Error != nil {
			nodeHealth.LastErr = r.Error.Error()
		}
		s.nodeHealth[r.PiholeNode.Id] = nodeHealth
	}
	s.recomputeLocked()
}

func pickStatus(valid bool, err error) Status {
	switch {
	case err != nil:
		return StatusOffline
	case valid:
		return StatusOnline
	default:
		return StatusDegraded
	}
}

func (s *Service) recomputeLocked() {
	s.logger.Trace().Msg("recomputing summary")
	online := 0

	for _, nodeHealth := range s.nodeHealth {
		if nodeHealth.Status == StatusOnline {
			online++
		}
	}
	s.summary = Summary{
		Online:    online,
		Total:     len(s.nodeHealth),
		UpdatedAt: time.Now(),
	}
	s.logger.Trace().Int("online", online).Int("total", len(s.nodeHealth)).Time("updated_at", s.summary.UpdatedAt).Msg("summary recomputed")

	if b, err := json.Marshal(s.summary); err == nil {
		s.broker.Publish("health_summary", b)
	} else {
		s.logger.Trace().Err(err).Msg("error serializing summary for broadcasting")
	}

	list := make([]NodeHealth, 0, len(s.nodeHealth))
	for _, nh := range s.nodeHealth {
		list = append(list, nh)
	}
	if b, err := json.Marshal(list); err == nil {
		s.broker.Publish("node_health", b)
	} else {
		s.logger.Trace().Err(err).Msg("error serializing node health for broadcasting")
	}
}

func (s *Service) NodeHealth() map[int64]NodeHealth {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.nodeHealth
}

func (s *Service) Summary() Summary {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.summary
}

func (s *Service) All() []NodeHealth {
	s.mu.RLock()
	defer s.mu.RUnlock()
	output := make([]NodeHealth, 0, len(s.nodeHealth))
	for _, nodeHealth := range s.nodeHealth {
		output = append(output, nodeHealth)
	}
	return output
}

func jitter(d time.Duration) time.Duration {
	// ~20% jitter
	j := d / 5
	return d - j + time.Duration(randInt63n(int64(2*j)))
}

func randInt63n(n int64) int64 {
	return time.Now().UnixNano() % n
}
