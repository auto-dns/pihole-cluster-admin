package user

import (
	"github.com/auto-dns/pihole-cluster-admin/internal/domain"
	user_s "github.com/auto-dns/pihole-cluster-admin/internal/service/user"
)

type service interface {
	Patch(id int64, params user_s.PatchUserParams) (*domain.User, error)
	UpdatePassword(id int64, params user_s.UpdatePasswordParams) (*domain.User, error)
}
