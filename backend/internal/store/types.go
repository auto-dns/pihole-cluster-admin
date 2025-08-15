package store

import (
	"time"

	"github.com/auto-dns/pihole-cluster-admin/internal/domain"
)

// Initialization status store

type initStatusRow struct {
	UserCreated  bool
	PiholeStatus domain.PiholeStatus
}

// Pihole store

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

// Session store

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

// User store

type User struct {
	Id           int64
	Username     string
	PasswordHash *string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type CreateUserParams struct {
	Username string
	Password string
}
