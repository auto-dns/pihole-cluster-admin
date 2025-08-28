package v1

import (
	"time"

	"github.com/auto-dns/pihole-cluster-admin/internal/domain"
)

type userResponseDTO struct {
	Id        int64     `json:"id"`
	Username  string    `json:"username"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func FromDomainUser(u *domain.User) userResponseDTO {
	return userResponseDTO{
		Id:        u.Id,
		Username:  u.Username,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}
