package auditlogsvc

import (
	"context"

	"github.com/auto-dns/pihole-cluster-admin/internal/domain"
)

type auditStore interface {
	Create(ctx context.Context, params domain.CreateAuditEntryParams) (*domain.AuditEntry, error)
	List(ctx context.Context, q domain.ListAuditEntriesQuery) ([]*domain.AuditEntry, int, error)
}
