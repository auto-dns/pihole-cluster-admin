package clusterblocking

import (
	"context"

	"github.com/auto-dns/pihole-cluster-admin/internal/domain"
)

type cluster interface {
	GetBlockingState(ctx context.Context) map[int64]*domain.NodeResult[*domain.BlockingState]
}
