package querylog

import (
	"context"

	"github.com/auto-dns/pihole-cluster-admin/internal/pihole"
	"github.com/rs/zerolog"
)

type Service struct {
	cluster cluster
	logger  zerolog.Logger
}

func NewService(cluster cluster, logger zerolog.Logger) *Service {
	return &Service{
		cluster: cluster,
		logger:  logger,
	}
}

func (s *Service) Fetch(ctx context.Context, params pihole.FetchQueryLogClusterRequest) (*pihole.FetchQueryLogsClusterResponse, error) {
	return s.cluster.FetchQueryLogs(ctx, params)
}
