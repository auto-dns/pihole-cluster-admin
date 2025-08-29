package v1

import (
	"time"

	"github.com/auto-dns/pihole-cluster-admin/internal/domain"
)

type piholeNodeAddRequestDTO struct {
	Scheme      string `json:"scheme"`
	Host        string `json:"host"`
	Port        int    `json:"port"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Password    string `json:"password"`
}

type piholeNodeUpdateRequestDTO struct {
	Scheme      *string `json:"scheme"`
	Host        *string `json:"host"`
	Port        *int    `json:"port"`
	Name        *string `json:"name"`
	Description *string `json:"description"`
	Password    *string `json:"password"`
}

type piholeNodeDTO struct {
	Id          int64     `json:"id"`
	Scheme      string    `json:"scheme"`
	Host        string    `json:"host"`
	Port        int       `json:"port"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

func FromDomainPiholeNode(d domain.PiholeNode) piholeNodeDTO {
	return piholeNodeDTO{
		Id:          d.Id,
		Scheme:      d.Scheme,
		Host:        d.Host,
		Port:        d.Port,
		Name:        d.Name,
		Description: d.Description,
		CreatedAt:   d.CreatedAt,
		UpdatedAt:   d.UpdatedAt,
	}
}

type piholeTestInstanceConnectionRequestDTO struct {
	Scheme   string `json:"scheme"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Password string `json:"password"`
}

type piholeTestExistingConnectionRequestDTO struct {
	Scheme   *string `json:"scheme"`
	Host     *string `json:"host"`
	Port     *int    `json:"port"`
	Password *string `json:"password"`
}

// Generics used in multiple handlers

type piholeNodeRefDTO struct {
	Id   int64  `json:"id"`
	Name string `json:"name"`
	Host string `json:"host"`
}
