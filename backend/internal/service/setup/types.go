package setup

import "github.com/auto-dns/pihole-cluster-admin/internal/domain"

type CreateUserCommand struct {
	Username string
	Password string
}

type UpdatePiholeInitializationStatusCommand struct {
	Status domain.PiholeStatus
}
