package sessions

import "time"

type ContextKey string

const UserIdContextKey ContextKey = "userId"

type Session struct {
	Id        string
	UserId    int64
	ExpiresAt time.Time
}
