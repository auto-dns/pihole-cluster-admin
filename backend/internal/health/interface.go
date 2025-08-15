package health

import (
	"context"

	"github.com/auto-dns/pihole-cluster-admin/internal/domain"
	"github.com/auto-dns/pihole-cluster-admin/internal/pihole"
)

type broker interface {
	SubscriberCount() int64
	SubscribersChanged() <-chan struct{}
	Publish(topic string, payload []byte)
}

type piholeCluster interface {
	AuthStatus(ctx context.Context) map[int64]*domain.NodeResult[pihole.AuthResponse]
}
