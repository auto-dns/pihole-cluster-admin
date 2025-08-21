package userhandler

import (
	"github.com/auto-dns/pihole-cluster-admin/internal/domain"
	"github.com/auto-dns/pihole-cluster-admin/internal/service/userservice"
)

type service interface {
	Patch(id int64, params userservice.PatchUserParams) (*domain.User, error)
	UpdatePassword(id int64, params userservice.UpdatePasswordParams) (*domain.User, error)
}
