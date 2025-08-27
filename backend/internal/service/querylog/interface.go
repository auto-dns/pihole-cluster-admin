package querylog

import (
	"context"

	"github.com/auto-dns/pihole-cluster-admin/internal/domain"
)

type cluster interface {
	FetchQueryLogs(ctx context.Context, req domain.QueryLogRequest) (*domain.ClusterQueryLogResponse, error)
}
