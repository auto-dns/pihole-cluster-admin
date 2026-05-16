package syncsvc

import (
	"context"
	"fmt"

	"github.com/auto-dns/pihole-cluster-admin/internal/domain"
	"github.com/auto-dns/pihole-cluster-admin/internal/errs"
	"github.com/rs/zerolog"
)

type Service struct {
	cluster syncCluster
	audit   auditLogger
	logger  zerolog.Logger
}

func NewService(cluster syncCluster, audit auditLogger, logger zerolog.Logger) *Service {
	return &Service{cluster: cluster, audit: audit, logger: logger}
}

type ruleKey struct {
	Type   domain.RuleType
	Kind   domain.RuleKind
	Domain string
}

type typeKindKey struct {
	Type domain.RuleType
	Kind domain.RuleKind
}

func (s *Service) SyncFromNode(ctx context.Context, sourceNodeId int64) ([]domain.SyncNodeResult, error) {
	allRules := s.cluster.ListDomainRules(ctx, domain.ListDomainRulesQuery{})

	source, ok := allRules[sourceNodeId]
	if !ok {
		return nil, errs.NotFound("source node not found", fmt.Errorf("node %d not found", sourceNodeId))
	}
	if !source.Success || source.Response == nil {
		return nil, errs.Unavailable("could not fetch rules from source node", source.Error)
	}

	sourceRules := source.Response.Rules
	sourceName := source.PiholeNode.Name

	// Index source rules by key for fast lookup
	sourceSet := make(map[ruleKey]bool, len(sourceRules))
	for _, r := range sourceRules {
		sourceSet[ruleKey{r.Type, r.Kind, r.Domain}] = true
	}

	results := make([]domain.SyncNodeResult, 0, len(allRules)-1)

	for nodeId, nodeResult := range allRules {
		if nodeId == sourceNodeId {
			continue
		}

		res := domain.SyncNodeResult{
			NodeId:   nodeId,
			NodeName: nodeResult.PiholeNode.Name,
		}

		if !nodeResult.Success || nodeResult.Response == nil {
			res.Error = fmt.Sprintf("failed to fetch rules: %v", nodeResult.Error)
			results = append(results, res)
			continue
		}

		targetRules := nodeResult.Response.Rules

		// Index target rules
		targetSet := make(map[ruleKey]bool, len(targetRules))
		for _, r := range targetRules {
			targetSet[ruleKey{r.Type, r.Kind, r.Domain}] = true
		}

		// Compute diff
		var toAdd []domain.DomainRule
		for _, r := range sourceRules {
			if !targetSet[ruleKey{r.Type, r.Kind, r.Domain}] {
				toAdd = append(toAdd, r)
			}
		}
		var toRemove []domain.DomainRule
		for _, r := range targetRules {
			if !sourceSet[ruleKey{r.Type, r.Kind, r.Domain}] {
				toRemove = append(toRemove, r)
			}
		}

		// Add missing rules (grouped by type+kind for batching)
		addGroups := make(map[typeKindKey][]string)
		for _, r := range toAdd {
			k := typeKindKey{r.Type, r.Kind}
			addGroups[k] = append(addGroups[k], r.Domain)
		}
		for k, domains := range addGroups {
			addResults := s.cluster.AddDomainRuleToNodes(ctx, domain.AddDomainRulesCommand{
				Type:    k.Type,
				Kind:    k.Kind,
				Domains: domains,
			}, []int64{nodeId})
			if nr, exists := addResults[nodeId]; exists {
				if nr.Success {
					res.Added += len(domains)
				} else if nr.Error != nil {
					if res.Error != "" {
						res.Error += "; "
					}
					res.Error += nr.Error.Error()
				}
			}
		}

		// Remove extra rules (one at a time)
		for _, r := range toRemove {
			removeResults := s.cluster.RemoveDomainRuleFromNodes(ctx, domain.RemoveDomainRuleCommand{
				Type:   r.Type,
				Kind:   r.Kind,
				Domain: r.Domain,
			}, []int64{nodeId})
			if nr, exists := removeResults[nodeId]; exists {
				if nr.Success {
					res.Removed++
				} else if nr.Error != nil {
					if res.Error != "" {
						res.Error += "; "
					}
					res.Error += nr.Error.Error()
				}
			}
		}

		res.Success = res.Error == ""
		results = append(results, res)
	}

	// Record a single audit entry for the sync operation
	auditResults := make([]domain.AuditNodeResult, len(results))
	for i, r := range results {
		auditResults[i] = domain.AuditNodeResult{
			NodeId:   r.NodeId,
			NodeName: r.NodeName,
			Success:  r.Success,
			Error:    r.Error,
		}
	}
	s.audit.Record(ctx, domain.CreateAuditEntryParams{
		Action:         domain.AuditActionSyncFromNode,
		TargetNodeId:   &sourceNodeId,
		TargetNodeName: &sourceName,
		NodeResults:    auditResults,
	})

	return results, nil
}
