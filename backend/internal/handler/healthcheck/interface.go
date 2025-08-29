package healthcheck

import "context"

type pinger interface {
	PingContext(ctx context.Context) error
}
