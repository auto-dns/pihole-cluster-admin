package domainruleservice

import (
	"context"

	"github.com/auto-dns/pihole-cluster-admin/internal/domain"
	"github.com/auto-dns/pihole-cluster-admin/internal/pihole"
)

type cluster interface {
	GetAllDomainRules(ctx context.Context) map[int64]*domain.NodeResult[pihole.GetDomainRulesResponse]
	GetDomainRulesByType(ctx context.Context, opts pihole.GetDomainRulesByTypeOptions) map[int64]*domain.NodeResult[pihole.GetDomainRulesResponse]
	GetDomainRulesByKind(ctx context.Context, opts pihole.GetDomainRulesByKindOptions) map[int64]*domain.NodeResult[pihole.GetDomainRulesResponse]
	GetDomainRulesByDomain(ctx context.Context, opts pihole.GetDomainRulesByDomainOptions) map[int64]*domain.NodeResult[pihole.GetDomainRulesResponse]
	GetDomainRulesByTypeKind(ctx context.Context, opts pihole.GetDomainRulesByTypeKindOptions) map[int64]*domain.NodeResult[pihole.GetDomainRulesResponse]
	GetDomainRulesByTypeKindDomain(ctx context.Context, opts pihole.GetDomainRulesByTypeKindDomainOptions) map[int64]*domain.NodeResult[pihole.GetDomainRulesResponse]
	AddDomainRule(ctx context.Context, opts pihole.AddDomainRuleOptions) map[int64]*domain.NodeResult[pihole.AddDomainRuleResponse]
	RemoveDomainRule(ctx context.Context, opts pihole.RemoveDomainRuleOptions) map[int64]*domain.NodeResult[pihole.RemoveDomainRuleResponse]
}
