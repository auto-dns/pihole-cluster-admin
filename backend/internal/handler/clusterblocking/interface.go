package cluster

import (
	"context"

	clusterblocking_s "github.com/auto-dns/pihole-cluster-admin/internal/service/clusterblocking"
)

type service interface {
	GetState(ctx context.Context) (*clusterblocking_s.ClusterBlockingState, error)
}
