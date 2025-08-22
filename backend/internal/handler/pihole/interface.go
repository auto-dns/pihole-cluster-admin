package pihole

import (
	"context"

	"github.com/auto-dns/pihole-cluster-admin/internal/domain"
	pihole_s "github.com/auto-dns/pihole-cluster-admin/internal/service/pihole"
	"github.com/auto-dns/pihole-cluster-admin/internal/store"
)

type service interface {
	GetAll() ([]*domain.PiholeNode, error)
	Add(ctx context.Context, params store.AddPiholeParams) (*domain.PiholeNode, error)
	Update(ctx context.Context, id int64, params store.UpdatePiholeParams) (*domain.PiholeNode, error)
	Remove(ctx context.Context, id int64) (found bool, err error)
	TestExistingConnection(ctx context.Context, id int64, params pihole_s.TestExistingConnectionParams) error
	TestInstanceConnection(ctx context.Context, params pihole_s.TestInstanceConnectionParams) error
}
