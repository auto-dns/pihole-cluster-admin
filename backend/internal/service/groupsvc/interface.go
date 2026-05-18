package groupsvc

import (
	"context"

	"github.com/auto-dns/pihole-cluster-admin/internal/domain"
)

type cluster interface {
	ListGroups(ctx context.Context) map[int64]*domain.NodeResult[*domain.GroupSet]
	AddGroup(ctx context.Context, cmd domain.AddGroupCommand) map[int64]*domain.NodeResult[*domain.GroupSet]
	UpdateGroup(ctx context.Context, cmd domain.UpdateGroupCommand) map[int64]*domain.NodeResult[*domain.GroupSet]
	RemoveGroup(ctx context.Context, cmd domain.RemoveGroupCommand) map[int64]*domain.NodeResult[struct{}]
}

type auditLogger interface {
	Record(ctx context.Context, params domain.CreateAuditEntryParams)
}
