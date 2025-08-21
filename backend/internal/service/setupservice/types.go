package setupservice

import "github.com/auto-dns/pihole-cluster-admin/internal/domain"

type CreateUserParams struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type UpdatePiholeInitializationStatusParams struct {
	Status domain.PiholeStatus `json:"status"`
}
