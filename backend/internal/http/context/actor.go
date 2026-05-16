package context

import "context"

type actorKey struct{}

func WithActor(ctx context.Context, username string) context.Context {
	return context.WithValue(ctx, actorKey{}, username)
}

func Actor(ctx context.Context) string {
	if v, ok := ctx.Value(actorKey{}).(string); ok && v != "" {
		return v
	}
	return "unknown"
}
