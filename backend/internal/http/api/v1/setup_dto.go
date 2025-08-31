package v1

import (
	"github.com/auto-dns/pihole-cluster-admin/internal/domain"
	setupsvc "github.com/auto-dns/pihole-cluster-admin/internal/service/setup"
)

type isInitializedResponseDTO struct {
	Initialized bool `json:"initialized"`
}

type createUserRequestDTO struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type getInitializationStatusRequestDTO struct {
	UserCreated  bool   `json:"userCreated"`
	PiholeStatus string `json:"piholeStatus"`
}

func toInitStatusFromDomain(d *domain.InitStatus) getInitializationStatusRequestDTO {
	return getInitializationStatusRequestDTO{
		UserCreated:  d.UserCreated,
		PiholeStatus: string(d.PiholeStatus),
	}
}

type updatePiholeInitializationStatusRequestDTO struct {
	Status string `json:"status"`
}

func toUpdatePiholeInitStatusCommand(d updatePiholeInitializationStatusRequestDTO) setupsvc.UpdatePiholeInitializationStatusCommand {
	return setupsvc.UpdatePiholeInitializationStatusCommand{
		Status: domain.PiholeStatus(d.Status),
	}
}
