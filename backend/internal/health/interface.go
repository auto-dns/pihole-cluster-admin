package health

import "context"

type ServiceInterface interface {
	Start(ctx context.Context)
	Summary() Summary
	All() []*NodeHealth
}
