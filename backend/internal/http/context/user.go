package context

import "context"

type userIdKey struct{}

func WithUserID(ctx context.Context, id int64) context.Context {
	return context.WithValue(ctx, userIdKey{}, id)
}

func UserID(ctx context.Context) (int64, bool) {
	v := ctx.Value(userIdKey{})
	id, ok := v.(int64)
	return id, ok
}
