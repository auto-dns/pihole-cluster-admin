package domainrulehandler

import (
	"context"

	"github.com/auto-dns/pihole-cluster-admin/internal/domain"
	"github.com/auto-dns/pihole-cluster-admin/internal/pihole"
)

type service interface {
	GetAll(ctx context.Context) map[int64]*domain.NodeResult[pihole.GetDomainRulesResponse]
	GetByType(ctx context.Context, opts pihole.GetDomainRulesByTypeOptions) map[int64]*domain.NodeResult[pihole.GetDomainRulesResponse]
	GetByKind(ctx context.Context, opts pihole.GetDomainRulesByKindOptions) map[int64]*domain.NodeResult[pihole.GetDomainRulesResponse]
	GetByDomain(ctx context.Context, opts pihole.GetDomainRulesByDomainOptions) map[int64]*domain.NodeResult[pihole.GetDomainRulesResponse]
	GetByTypeKind(ctx context.Context, opts pihole.GetDomainRulesByTypeKindOptions) map[int64]*domain.NodeResult[pihole.GetDomainRulesResponse]
	GetByTypeKindDomain(ctx context.Context, opts pihole.GetDomainRulesByTypeKindDomainOptions) map[int64]*domain.NodeResult[pihole.GetDomainRulesResponse]
	Add(ctx context.Context, opts pihole.AddDomainRuleOptions) map[int64]*domain.NodeResult[pihole.AddDomainRuleResponse]
	Remove(ctx context.Context, opts pihole.RemoveDomainRuleOptions) map[int64]*domain.NodeResult[pihole.RemoveDomainRuleResponse]
}
