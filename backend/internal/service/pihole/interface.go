package pihole

import (
	"context"

	"github.com/auto-dns/pihole-cluster-admin/internal/domain"
	p "github.com/auto-dns/pihole-cluster-admin/internal/pihole"
	"github.com/auto-dns/pihole-cluster-admin/internal/store"
)

type cluster interface {
	HasClient(ctx context.Context, id int64) bool
	AddClient(ctx context.Context, client *p.Client) error
	UpdateClient(ctx context.Context, id int64, cfg *p.ClientConfig) error
	RemoveClient(ctx context.Context, id int64) error
}

type piholeStore interface {
	GetAllPiholeNodes() ([]*domain.PiholeNode, error)
	GetPiholeNode(id int64) (*domain.PiholeNode, error)
	GetPiholeNodeSecret(id int64) (*domain.PiholeNodeSecret, error)
	AddPiholeNode(params store.AddPiholeParams) (*domain.PiholeNode, error)
	UpdatePiholeNode(id int64, params store.UpdatePiholeParams) (*domain.PiholeNode, error)
	RemovePiholeNode(id int64) (bool, error)
}
