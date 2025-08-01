package server

import "context"

type ServerInterface interface {
	Start(ctx context.Context) error
}
