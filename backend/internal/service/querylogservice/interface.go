package querylogservice

import (
	"context"

	"github.com/auto-dns/pihole-cluster-admin/internal/pihole"
)

type cluster interface {
	FetchQueryLogs(ctx context.Context, params pihole.FetchQueryLogClusterRequest) (*pihole.FetchQueryLogsClusterResponse, error)
}
