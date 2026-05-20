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
	broker  broker
	cluster cluster
	logger  zerolog.Logger
	cfg     config.PublisherConfig
}

func NewService(broker broker, cluster cluster, cfg config.PublisherConfig, logger zerolog.Logger) *Service {
	return &Service{
		broker:  broker,
		cluster: cluster,
		cfg:     cfg,
		logger:  logger,
	}
}

func (s *Service) GetClusterHealth(ctx context.Context) domain.ClusterHealth {
	var (
		authNodes    map[int64]*domain.NodeResult[*domain.AuthStatus]
		versionNodes map[int64]*domain.NodeResult[*domain.NodeVersionInfo]
		dbNodes      map[int64]*domain.NodeResult[*domain.NodeDBInfo]
	)
	var wg sync.WaitGroup
	wg.Add(3)
	go func() { defer wg.Done(); authNodes = s.cluster.AuthStatus(ctx) }()
	go func() { defer wg.Done(); versionNodes = s.cluster.GetVersionInfo(ctx) }()
	go func() { defer wg.Done(); dbNodes = s.cluster.GetDatabaseInfo(ctx) }()
	wg.Wait()

	res := domain.ClusterHealth{
		Summary:   domain.ClusterHealthSummary{},
		Nodes:     make(map[int64]domain.ClusterNodeHealth, len(authNodes)),
		UpdatedAt: time.Now(),
	}

	for id, n := range authNodes {
		var tookMs int
		var valid bool

		if n.Error == nil {
			tookMs = int(n.Response.Took.Milliseconds())
			valid = n.Response.Valid
		}

		nodeHealth := domain.ClusterNodeHealth{
			Id:      n.PiholeNode.Id,
			Name:    n.PiholeNode.Name,
			Status:  pickStatus(valid, n.Error),
			Latency: time.Duration(tookMs * int(time.Millisecond)),
		}
		if n.Error != nil {
			nodeHealth.LastErr = n.Error.Error()
		}
		if v, ok := versionNodes[id]; ok && v.Success {
			nodeHealth.VersionInfo = v.Response
		}
		if d, ok := dbNodes[id]; ok && d.Success {
			nodeHealth.DBInfo = d.Response
		}

		res.Summary.Total++
		if nodeHealth.Status == domain.ClusterHealthStatusOnline {
			res.Summary.Online++
		}

		res.Nodes[id] = nodeHealth
	}

	return res
}

func (s *Service) StartPublisher(ctx context.Context) {
	s.logger.Info().Msg("Starting health service")
	s.getAndPublish(ctx)
	p := poller.New(s.broker, poller.Config{
		Interval:    time.Duration(max(1, float64(s.cfg.PollingIntervalSeconds))) * time.Second,
		GracePeriod: time.Duration(s.cfg.GracePeriodSeconds) * time.Second,
		JitterRatio: 0.20,
	})
	go p.Run(ctx, s.getAndPublish)
}

func (s *Service) getAndPublish(ctx context.Context) {
	ctx = logger.WithMode(ctx, logger.ModeTrace)
	clusterHealth := s.GetClusterHealth(ctx)
	s.broker.Publish(realtime.TopicClusterHealthV1, clusterHealth)
}

func pickStatus(valid bool, err error) domain.ClusterHealthStatus {
	switch {
	case err != nil:
		return domain.ClusterHealthStatusOffline
	case valid:
		return domain.ClusterHealthStatusOnline
	default:
		return domain.ClusterHealthStatusDegraded
	}
}
