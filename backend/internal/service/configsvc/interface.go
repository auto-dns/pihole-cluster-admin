package configsvc

import (
	"context"

	"github.com/auto-dns/pihole-cluster-admin/internal/domain"
)

type cluster interface {
	GetConfig(ctx context.Context) map[int64]*domain.NodeResult[*domain.PiholeConfig]
	PatchConfig(ctx context.Context, patch domain.PiholeConfigPatch) map[int64]*domain.NodeResult[struct{}]
}

type auditLogger interface {
	Record(ctx context.Context, params domain.CreateAuditEntryParams)
}
