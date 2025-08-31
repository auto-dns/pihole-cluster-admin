package clusterblockingsvc

import (
	"context"

	"github.com/auto-dns/pihole-cluster-admin/internal/domain"
)

type cluster interface {
	GetBlockingState(ctx context.Context) map[int64]*domain.NodeResult[*domain.BlockingState]
	SetBlockingState(ctx context.Context, blocking bool, timer *int) map[int64]*domain.NodeResult[*domain.BlockingState]
}

type broker interface {
	SubscriberCount() int64
	SubscribersChanged() <-chan struct{}
	Publish(topic string, payload any)
}
