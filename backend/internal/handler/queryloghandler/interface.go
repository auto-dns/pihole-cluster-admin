package queryloghandler

import (
	"context"

	"github.com/auto-dns/pihole-cluster-admin/internal/pihole"
)

type service interface {
	Fetch(ctx context.Context, req pihole.FetchQueryLogClusterRequest) (*pihole.FetchQueryLogsClusterResponse, error)
}
