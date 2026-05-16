package domainrulesvc

import (
	"context"
	"strings"

	"github.com/auto-dns/pihole-cluster-admin/internal/domain"
	"github.com/auto-dns/pihole-cluster-admin/internal/util"
)

type Service struct {
	cluster cluster
	audit   auditLogger
}

func NewService(cluster cluster, audit auditLogger) *Service {
	return &Service{cluster: cluster, audit: audit}
}

func (s *Service) List(ctx context.Context, q domain.ListDomainRulesQuery) map[int64]*domain.NodeResult[*domain.DomainRuleSet] {
	return s.cluster.ListDomainRules(ctx, q)
}

func (s *Service) Add(ctx context.Context, cmd domain.AddDomainRulesCommand) map[int64]*domain.NodeResult[*domain.AddDomainRulesResult] {
	results := s.cluster.AddDomainRule(ctx, cmd)

	targetDomain := strings.Join(cmd.Domains, ", ")
	ruleType := string(cmd.Type)
	ruleKind := string(cmd.Kind)
	s.audit.Record(ctx, domain.CreateAuditEntryParams{
		Action:       domain.AuditActionAddDomainRule,
		TargetDomain: &targetDomain,
		RuleType:     &ruleType,
		RuleKind:     &ruleKind,
		NodeResults:  nodeResultsFromAdd(results),
	})

	return results
}

func (s *Service) Remove(ctx context.Context, cmd domain.RemoveDomainRuleCommand) map[int64]*domain.NodeResult[struct{}] {
	results := s.cluster.RemoveDomainRule(ctx, cmd)

	ruleType := string(cmd.Type)
	ruleKind := string(cmd.Kind)
	s.audit.Record(ctx, domain.CreateAuditEntryParams{
		Action:       domain.AuditActionRemoveDomainRule,
		TargetDomain: &cmd.Domain,
		RuleType:     &ruleType,
		RuleKind:     &ruleKind,
		NodeResults:  nodeResultsFromRemove(results),
	})

	return results
}

func nodeResultsFromAdd(results map[int64]*domain.NodeResult[*domain.AddDomainRulesResult]) []domain.AuditNodeResult {
	out := make([]domain.AuditNodeResult, 0, len(results))
	for _, nr := range results {
		out = append(out, domain.AuditNodeResult{
			NodeId:   nr.PiholeNode.Id,
			NodeName: nr.PiholeNode.Name,
			Success:  nr.Success,
			Error:    util.ErrorString(nr.Error),
		})
	}
	return out
}

func nodeResultsFromRemove(results map[int64]*domain.NodeResult[struct{}]) []domain.AuditNodeResult {
	out := make([]domain.AuditNodeResult, 0, len(results))
	for _, nr := range results {
		out = append(out, domain.AuditNodeResult{
			NodeId:   nr.PiholeNode.Id,
			NodeName: nr.PiholeNode.Name,
			Success:  nr.Success,
			Error:    util.ErrorString(nr.Error),
		})
	}
	return out
}
