package healthsvc

import (
	"context"

	"github.com/auto-dns/pihole-cluster-admin/internal/domain"
)

type broker interface {
	SubscriberCount() int64
	SubscribersChanged() <-chan struct{}
	Publish(topic string, payload any)
}

type cluster interface {
	AuthStatus(ctx context.Context) map[int64]*domain.NodeResult[*domain.AuthStatus]
	GetVersionInfo(ctx context.Context) map[int64]*domain.NodeResult[*domain.NodeVersionInfo]
	GetDatabaseInfo(ctx context.Context) map[int64]*domain.NodeResult[*domain.NodeDBInfo]
}
