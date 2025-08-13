package health

import "context"

type ServiceInterface interface {
	Start(ctx context.Context)
	NodeHealth() map[int64]*NodeHealth
	Summary() Summary
	All() []*NodeHealth
}
