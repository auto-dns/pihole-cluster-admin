package app

import "context"

type HealthService interface {
	Start(ctx context.Context)
}
type HttpServer interface {
	StartAndServe(ctx context.Context) error
}

type SessionPurger interface {
	Start(ctx context.Context)
}
