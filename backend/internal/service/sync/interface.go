package syncsvc

import (
	"context"

	"github.com/auto-dns/pihole-cluster-admin/internal/domain"
)

type syncCluster interface {
	ListDomainRules(ctx context.Context, q domain.ListDomainRulesQuery) map[int64]*domain.NodeResult[*domain.DomainRuleSet]
	AddDomainRuleToNodes(ctx context.Context, cmd domain.AddDomainRulesCommand, nodeIDs []int64) map[int64]*domain.NodeResult[*domain.AddDomainRulesResult]
	RemoveDomainRuleFromNodes(ctx context.Context, cmd domain.RemoveDomainRuleCommand, nodeIDs []int64) map[int64]*domain.NodeResult[struct{}]
}

type auditLogger interface {
	Record(ctx context.Context, params domain.CreateAuditEntryParams)
}
