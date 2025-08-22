package auth

import (
	"github.com/auto-dns/pihole-cluster-admin/internal/domain"
)

type userStore interface {
	ValidateUser(username, password string) (*domain.User, error)
	GetUser(id int64) (*domain.User, error)
}

type sessionIssuer interface {
	CreateSession(userId int64) (string, error)
	DestroySession(userId string) error
	GetUserId(sessionId string) (int64, bool, error)
}
