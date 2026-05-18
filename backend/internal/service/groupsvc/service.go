package groupsvc

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

func (s *Service) List(ctx context.Context) map[int64]*domain.NodeResult[*domain.GroupSet] {
	return s.cluster.ListGroups(ctx)
}

func (s *Service) Add(ctx context.Context, cmd domain.AddGroupCommand) map[int64]*domain.NodeResult[*domain.GroupSet] {
	results := s.cluster.AddGroup(ctx, cmd)
	s.audit.Record(ctx, domain.CreateAuditEntryParams{
		Action:      domain.AuditActionAddGroup,
		NodeResults: nodeResultsFromGroupSet(results),
	})
	return results
}

func (s *Service) Update(ctx context.Context, cmd domain.UpdateGroupCommand) map[int64]*domain.NodeResult[*domain.GroupSet] {
	results := s.cluster.UpdateGroup(ctx, cmd)
	s.audit.Record(ctx, domain.CreateAuditEntryParams{
		Action:      domain.AuditActionUpdateGroup,
		NodeResults: nodeResultsFromGroupSet(results),
	})
	return results
}

func (s *Service) Remove(ctx context.Context, cmd domain.RemoveGroupCommand) map[int64]*domain.NodeResult[struct{}] {
	results := s.cluster.RemoveGroup(ctx, cmd)
	s.audit.Record(ctx, domain.CreateAuditEntryParams{
		Action:      domain.AuditActionRemoveGroup,
		NodeResults: nodeResultsFromRemove(results),
	})
	return results
}

func nodeResultsFromGroupSet(results map[int64]*domain.NodeResult[*domain.GroupSet]) []domain.AuditNodeResult {
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
