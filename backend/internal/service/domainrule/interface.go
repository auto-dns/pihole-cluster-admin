package domainrulesvc

import (
	"context"

	"github.com/auto-dns/pihole-cluster-admin/internal/domain"
)

type cluster interface {
	ListDomainRules(ctx context.Context, q domain.ListDomainRulesQuery) map[int64]*domain.NodeResult[*domain.DomainRuleSet]
	AddDomainRule(ctx context.Context, cmd domain.AddDomainRulesCommand) map[int64]*domain.NodeResult[*domain.AddDomainRulesResult]
	AddDomainRuleToNodes(ctx context.Context, cmd domain.AddDomainRulesCommand, nodeIDs []int64) map[int64]*domain.NodeResult[*domain.AddDomainRulesResult]
	RemoveDomainRule(ctx context.Context, cmd domain.RemoveDomainRuleCommand) map[int64]*domain.NodeResult[struct{}]
}
