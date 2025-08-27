package store

import "time"

type piholeRow struct {
	Id          int64
	Scheme      string
	Host        string
	Port        int
	Name        string
	Description string
	PasswordEnc string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type AddPiholeParams struct {
	Scheme      string
	Host        string
	Port        int
	Name        string
	Description string
	Password    string
}

type UpdatePiholeParams struct {
	Scheme      *string
	Host        *string
	Port        *int
	Name        *string
	Description *string
	Password    *string
}
