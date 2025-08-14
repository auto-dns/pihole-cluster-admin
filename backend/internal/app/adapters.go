package app

import "context"

type purgeAdapter struct {
	sm interface{ StartPurgeLoop(context.Context) }
}

func (a purgeAdapter) Start(ctx context.Context) { a.sm.StartPurgeLoop(ctx) }
