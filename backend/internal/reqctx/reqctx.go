package reqctx

import "context"

type ctxKey int

const requestIdKey ctxKey = iota

func WithRequestId(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, requestIdKey, id)
}

func RequestIdFrom(ctx context.Context) string {
	if v, ok := ctx.Value(requestIdKey).(string); ok {
		return v
	}
	return ""
}
