package app

import (
	"context"

	"github.com/auto-dns/pihole-cluster-admin/internal/sessions"
	"github.com/auto-dns/pihole-cluster-admin/internal/store"
)

type Broker interface {
	SubscriberCount() int
	SubscribersChanged() <-chan struct{}
	Publish(topic string, payload []byte)
}

type HealthService interface {
	Start(ctx context.Context)
}
type HttpServer interface {
	StartAndServe(ctx context.Context) error
}

type SessionPurger interface {
	Start(ctx context.Context)
}

type SessionStorage interface {
	Create(session sessions.Session) error
	GetAll() ([]sessions.Session, error)
	GetUserId(sessionId string) (int64, bool, error)
	Delete(sessionId string) error
}

type PiholeGetter interface {
	GetAllPiholeNodesWithPasswords() ([]*store.PiholeNode, error)
}

type PiholeSqliteStore interface {
	AddPiholeNode(params store.AddPiholeParams) (*store.PiholeNode, error)
	UpdatePiholeNode(id int64, params store.UpdatePiholeParams) (*store.PiholeNode, error)
	RemovePiholeNode(id int64) (found bool, err error)
	GetAllPiholeNodes() ([]*store.PiholeNode, error)
	GetPiholeNodeWithPassword(id int64) (*store.PiholeNode, error)
}

type SessionSqliteStore interface {
	CreateSession(params store.CreateSessionParams) (*store.Session, error)
	GetAllSessions() ([]*store.Session, error)
	GetSession(id string) (*store.Session, error)
	DeleteSession(id string) (found bool, err error)
}
