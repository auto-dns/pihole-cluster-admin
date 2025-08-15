package domain

import "time"

type User struct {
	Id        int64
	Username  string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type UserAuth struct {
	PasswordHash string
}
