package store

import "time"

type userRow struct {
	Id           int64
	Username     string
	PasswordHash string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type CreateUserParams struct {
	Username string
	Password string
}

type UpdateUserParams struct {
	Username *string
	Password *string
}
