package adlistsvc

import (
	"context"

	"github.com/auto-dns/pihole-cluster-admin/internal/domain"
)

type cluster interface {
	ListAdlists(ctx context.Context) map[int64]*domain.NodeResult[*domain.AdlistSet]
	AddAdlist(ctx context.Context, cmd domain.AddAdlistCommand) map[int64]*domain.NodeResult[*domain.AdlistSet]
	UpdateAdlist(ctx context.Context, cmd domain.UpdateAdlistCommand) map[int64]*domain.NodeResult[*domain.AdlistSet]
	RemoveAdlist(ctx context.Context, cmd domain.RemoveAdlistCommand) map[int64]*domain.NodeResult[struct{}]
	RebuildGravity(ctx context.Context) map[int64]*domain.NodeResult[struct{}]
}

type auditLogger interface {
	Record(ctx context.Context, params domain.CreateAuditEntryParams)
}
