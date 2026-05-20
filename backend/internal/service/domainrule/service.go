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

func (s *Service) SyncRule(ctx context.Context, cmd domain.SyncDomainRuleCommand) map[int64]*domain.NodeResult[*domain.SyncDomainRuleNodeResult] {
	ruleType := cmd.Type
	ruleKind := cmd.Kind
	listResults := s.cluster.ListDomainRules(ctx, domain.ListDomainRulesQuery{
		Type: &ruleType,
		Kind: &ruleKind,
	})

	type entry struct {
		nr      *domain.NodeResult[*domain.DomainRuleSet]
		present bool
	}
	entries := make(map[int64]entry, len(listResults))
	missingNodeIDs := make([]int64, 0)

	for nodeID, nr := range listResults {
		present := false
		if nr.Success && nr.Response != nil {
			for _, r := range nr.Response.Rules {
				if r.Domain == cmd.Domain {
					present = true
					break
				}
			}
		}
		entries[nodeID] = entry{nr: nr, present: present}
		if !present {
			missingNodeIDs = append(missingNodeIDs, nodeID)
		}
	}

	out := make(map[int64]*domain.NodeResult[*domain.SyncDomainRuleNodeResult], len(listResults))

	for nodeID, e := range entries {
		if e.present {
			out[nodeID] = &domain.NodeResult[*domain.SyncDomainRuleNodeResult]{
				PiholeNode: e.nr.PiholeNode,
				Success:    true,
				Response:   &domain.SyncDomainRuleNodeResult{AlreadyPresent: true},
			}
		}
	}

	if len(missingNodeIDs) == 0 {
		return out
	}

	addCmd := domain.AddDomainRulesCommand{
		Type:    cmd.Type,
		Kind:    cmd.Kind,
		Domains: []string{cmd.Domain},
		Comment: cmd.Comment,
	}
	addResults := s.cluster.AddDomainRuleToNodes(ctx, addCmd, missingNodeIDs)

	for nodeID, nr := range addResults {
		out[nodeID] = &domain.NodeResult[*domain.SyncDomainRuleNodeResult]{
			PiholeNode: nr.PiholeNode,
			Success:    nr.Success,
			Error:      nr.Error,
			Response:   &domain.SyncDomainRuleNodeResult{AlreadyPresent: false, Added: nr.Success},
		}
	}

	return out
}

func (s *Service) TestRegex(ctx context.Context, testDomain string) map[int64]*domain.NodeResult[*domain.RegexTestResult] {
	return s.cluster.TestRegex(ctx, testDomain)
}
