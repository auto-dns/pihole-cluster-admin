package api

import "time"

type ContextKey string

const userIdContextKey ContextKey = "userId"

type session struct {
	Id        string
	UserId    int64
	ExpiresAt time.Time
}
