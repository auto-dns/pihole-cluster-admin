package domainrulesvc

import (
	"context"

	"github.com/auto-dns/pihole-cluster-admin/internal/domain"
)

type Service struct {
	cluster cluster
}

func NewService(cluster cluster) *Service {
	return &Service{
		cluster: cluster,
	}
}

func (s *Service) List(ctx context.Context, q domain.ListDomainRulesQuery) map[int64]*domain.NodeResult[*domain.DomainRuleSet] {
	return s.cluster.ListDomainRules(ctx, q)
}

func (s *Service) Add(ctx context.Context, cmd domain.AddDomainRulesCommand) map[int64]*domain.NodeResult[*domain.AddDomainRulesResult] {
	return s.cluster.AddDomainRule(ctx, cmd)
}

func (s *Service) Remove(ctx context.Context, cmd domain.RemoveDomainRuleCommand) map[int64]*domain.NodeResult[struct{}] {
	return s.cluster.RemoveDomainRule(ctx, cmd)
}
