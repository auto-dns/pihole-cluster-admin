package v1

import (
	"time"

	"github.com/auto-dns/pihole-cluster-admin/internal/domain"
)

type patchUserRequestDTO struct {
	Username *string `json:"username"`
}

type updatePasswordRequestDTO struct {
	CurrentPassword string `json:"currentPassword"`
	NewPassword     string `json:"newPassword"`
}

// Generic / used in various handlers

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
