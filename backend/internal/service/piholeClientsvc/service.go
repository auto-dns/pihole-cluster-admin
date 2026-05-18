package piholeClientsvc

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

func (s *Service) List(ctx context.Context) map[int64]*domain.NodeResult[*domain.PiholeClientSet] {
	return s.cluster.ListPiholeClients(ctx)
}

func (s *Service) Update(ctx context.Context, cmd domain.UpdatePiholeClientCommand) map[int64]*domain.NodeResult[*domain.PiholeClientSet] {
	results := s.cluster.UpdatePiholeClient(ctx, cmd)
	s.audit.Record(ctx, domain.CreateAuditEntryParams{
		Action:      domain.AuditActionUpdateClient,
		NodeResults: nodeResultsFromClientSet(results),
	})
	return results
}

func (s *Service) Remove(ctx context.Context, cmd domain.RemovePiholeClientCommand) map[int64]*domain.NodeResult[struct{}] {
	results := s.cluster.RemovePiholeClient(ctx, cmd)
	s.audit.Record(ctx, domain.CreateAuditEntryParams{
		Action:      domain.AuditActionRemoveClient,
		NodeResults: nodeResultsFromRemove(results),
	})
	return results
}

func nodeResultsFromClientSet(results map[int64]*domain.NodeResult[*domain.PiholeClientSet]) []domain.AuditNodeResult {
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
