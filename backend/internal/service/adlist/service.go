package adlistsvc

import (
	"context"

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

func (s *Service) List(ctx context.Context) map[int64]*domain.NodeResult[*domain.AdlistSet] {
	return s.cluster.ListAdlists(ctx)
}

func (s *Service) Add(ctx context.Context, cmd domain.AddAdlistCommand) map[int64]*domain.NodeResult[*domain.AdlistSet] {
	results := s.cluster.AddAdlist(ctx, cmd)
	s.audit.Record(ctx, domain.CreateAuditEntryParams{
		Action:       domain.AuditActionAddAdlist,
		TargetDomain: &cmd.Address,
		NodeResults:  nodeResultsFromAdlistSet(results),
	})
	return results
}

func (s *Service) Update(ctx context.Context, cmd domain.UpdateAdlistCommand) map[int64]*domain.NodeResult[*domain.AdlistSet] {
	results := s.cluster.UpdateAdlist(ctx, cmd)
	s.audit.Record(ctx, domain.CreateAuditEntryParams{
		Action:      domain.AuditActionUpdateAdlist,
		NodeResults: nodeResultsFromAdlistSet(results),
	})
	return results
}

func (s *Service) Remove(ctx context.Context, cmd domain.RemoveAdlistCommand) map[int64]*domain.NodeResult[struct{}] {
	results := s.cluster.RemoveAdlist(ctx, cmd)
	s.audit.Record(ctx, domain.CreateAuditEntryParams{
		Action:      domain.AuditActionRemoveAdlist,
		NodeResults: nodeResultsFromRemove(results),
	})
	return results
}

func (s *Service) RebuildGravity(ctx context.Context) map[int64]*domain.NodeResult[struct{}] {
	results := s.cluster.RebuildGravity(ctx)
	s.audit.Record(ctx, domain.CreateAuditEntryParams{
		Action:      domain.AuditActionRebuildGravity,
		NodeResults: nodeResultsFromRemove(results),
	})
	return results
}

func nodeResultsFromAdlistSet(results map[int64]*domain.NodeResult[*domain.AdlistSet]) []domain.AuditNodeResult {
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
