package userservice

import (
	"github.com/auto-dns/pihole-cluster-admin/internal/domain"
	"github.com/auto-dns/pihole-cluster-admin/internal/store"
)

type userStore interface {
	GetUser(id int64) (*domain.User, error)
	GetUserAuth(id int64) (*domain.UserAuth, error)
	UpdateUser(id int64, params store.UpdateUserParams) (*domain.User, error)
}
