package eventshandler

import (
	"context"

	"github.com/auto-dns/pihole-cluster-admin/internal/realtime"
)

type service interface {
	Subscribe(ctx context.Context, topics []string) (<-chan realtime.Event, func())
}
