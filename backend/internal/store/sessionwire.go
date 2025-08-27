package store

import "time"

type sessionRow struct {
	Id        string
	UserId    int64
	CreatedAt time.Time
	ExpiresAt time.Time
}

type CreateSessionParams struct {
	Id        string
	UserId    int64
	ExpiresAt time.Time
}
