package querylog

import (
	"context"

	"github.com/auto-dns/pihole-cluster-admin/internal/domain"
)

type service interface {
	Fetch(ctx context.Context, req domain.QueryLogRequest) (*domain.ClusterQueryLogResponse, error)
}
