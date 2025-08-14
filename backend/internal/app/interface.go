package app

import (
	"context"

	"github.com/auto-dns/pihole-cluster-admin/internal/sessions"
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
