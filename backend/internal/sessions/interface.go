package sessions

import (
	"github.com/auto-dns/pihole-cluster-admin/internal/domain"
	"github.com/auto-dns/pihole-cluster-admin/internal/store"
)

type storage interface {
	Create(session Session) error
	GetAll() ([]Session, error)
	GetUserId(sessionId string) (int64, bool, error)
	Delete(sessionId string) error
}

type sqliteStore interface {
	CreateSession(params store.CreateSessionParams) (*domain.Session, error)
	GetAllSessions() ([]*domain.Session, error)
	GetSession(id string) (*domain.Session, error)
	DeleteSession(id string) (found bool, err error)
}
