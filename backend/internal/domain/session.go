package domain

import "time"

type Session struct {
	Id        string
	UserId    int64
	CreatedAt time.Time
	ExpiresAt time.Time
}
