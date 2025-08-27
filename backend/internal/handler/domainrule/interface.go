package domainrule

import (
	"context"

	"github.com/auto-dns/pihole-cluster-admin/internal/domain"
)

type service interface {
	List(ctx context.Context, q domain.ListDomainRulesQuery) map[int64]*domain.NodeResult[*domain.DomainRuleSet]
	Add(ctx context.Context, cmd domain.AddDomainRulesCommand) map[int64]*domain.NodeResult[*domain.AddDomainRulesResult]
	Remove(ctx context.Context, cmd domain.RemoveDomainRuleCommand) map[int64]*domain.NodeResult[struct{}]
}
