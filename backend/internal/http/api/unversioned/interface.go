package unversioned

import (
	"context"

	"github.com/auto-dns/pihole-cluster-admin/internal/realtime"
)

type eventsService interface {
	Subscribe(ctx context.Context, topics []string) (<-chan realtime.Event, func())
}

type pinger interface {
	PingContext(ctx context.Context) error
}
