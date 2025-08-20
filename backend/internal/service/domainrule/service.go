package domainrule

import (
	"context"

	"github.com/auto-dns/pihole-cluster-admin/internal/domain"
	"github.com/auto-dns/pihole-cluster-admin/internal/pihole"
)

type Service struct {
	cluster cluster
}

func NewService(cluster cluster) *Service {
	return &Service{
		cluster: cluster,
	}
}

func (s *Service) GetAll(ctx context.Context) map[int64]*domain.NodeResult[pihole.GetDomainRulesResponse] {
	return s.cluster.GetAllDomainRules(ctx)
}

func (s *Service) GetByType(ctx context.Context, opts pihole.GetDomainRulesByTypeOptions) map[int64]*domain.NodeResult[pihole.GetDomainRulesResponse] {
	return s.cluster.GetDomainRulesByType(ctx, opts)
}

func (s *Service) GetByKind(ctx context.Context, opts pihole.GetDomainRulesByKindOptions) map[int64]*domain.NodeResult[pihole.GetDomainRulesResponse] {
	return s.cluster.GetDomainRulesByKind(ctx, opts)
}

func (s *Service) GetByDomain(ctx context.Context, opts pihole.GetDomainRulesByDomainOptions) map[int64]*domain.NodeResult[pihole.GetDomainRulesResponse] {
	return s.cluster.GetDomainRulesByDomain(ctx, opts)
}

func (s *Service) GetByTypeKind(ctx context.Context, opts pihole.GetDomainRulesByTypeKindOptions) map[int64]*domain.NodeResult[pihole.GetDomainRulesResponse] {
	return s.cluster.GetDomainRulesByTypeKind(ctx, opts)
}

func (s *Service) GetByTypeKindDomain(ctx context.Context, opts pihole.GetDomainRulesByTypeKindDomainOptions) map[int64]*domain.NodeResult[pihole.GetDomainRulesResponse] {
	return s.cluster.GetDomainRulesByTypeKindDomain(ctx, opts)
}

func (s *Service) Add(ctx context.Context, opts pihole.AddDomainRuleOptions) map[int64]*domain.NodeResult[pihole.AddDomainRuleResponse] {
	return s.cluster.AddDomainRule(ctx, opts)
}

func (s *Service) Remove(ctx context.Context, opts pihole.RemoveDomainRuleOptions) map[int64]*domain.NodeResult[pihole.RemoveDomainRuleResponse] {
	return s.cluster.RemoveDomainRule(ctx, opts)
}
