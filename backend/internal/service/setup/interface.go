package setup

import (
	"github.com/auto-dns/pihole-cluster-admin/internal/domain"
	"github.com/auto-dns/pihole-cluster-admin/internal/store"
)

type initStatusStore interface {
	SetUserCreated(bool) error
	GetInitializationStatus() (*domain.InitStatus, error)
	SetPiholeStatus(piholeStatus domain.PiholeStatus) error
}

type userStore interface {
	IsInitialized() (bool, error)
	CreateUser(store.CreateUserParams) (*domain.User, error)
}

type sessionIssuer interface {
	CreateSession(userId int64) (string, error)
}
