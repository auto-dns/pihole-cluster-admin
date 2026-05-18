package piholeClientsvc

import (
	"context"

	"github.com/auto-dns/pihole-cluster-admin/internal/domain"
)

type cluster interface {
	ListPiholeClients(ctx context.Context) map[int64]*domain.NodeResult[*domain.PiholeClientSet]
	UpdatePiholeClient(ctx context.Context, cmd domain.UpdatePiholeClientCommand) map[int64]*domain.NodeResult[*domain.PiholeClientSet]
	RemovePiholeClient(ctx context.Context, cmd domain.RemovePiholeClientCommand) map[int64]*domain.NodeResult[struct{}]
}

type auditLogger interface {
	Record(ctx context.Context, params domain.CreateAuditEntryParams)
}
