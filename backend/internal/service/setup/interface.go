package setupsvc

import (
	"context"

	"github.com/auto-dns/pihole-cluster-admin/internal/domain"
	"github.com/auto-dns/pihole-cluster-admin/internal/store"
)

type initStatusStore interface {
	SetUserCreatedTx(ctx context.Context, q store.DBTX, created bool) error
	GetInitializationStatus() (*domain.InitStatus, error)
	SetPiholeStatus(piholeStatus domain.PiholeStatus) error
}

type userStore interface {
	IsInitialized() (bool, error)
	IsInitializedTx(ctx context.Context, q store.DBTX) (bool, error)
	CreateUserTx(ctx context.Context, q store.DBTX, params store.CreateUserParams) (*domain.User, error)
}

type sessionIssuer interface {
	CreateSession(userId int64) (string, error)
}

type txProvider interface {
	WithTx(ctx context.Context, fn func(ctx context.Context, q store.DBTX) error) error
}
