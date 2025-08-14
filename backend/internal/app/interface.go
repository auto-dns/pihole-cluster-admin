package app

import "context"

type httpServer interface {
	Start(ctx context.Context) error
}
