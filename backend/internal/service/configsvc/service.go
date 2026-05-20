package configsvc

import (
	"context"
	"reflect"

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

func (s *Service) GetClusterConfig(ctx context.Context) domain.ClusterConfig {
	results := s.cluster.GetConfig(ctx)

	out := domain.ClusterConfig{
		PerNode: make(map[int64]*domain.PiholeConfig, len(results)),
	}

	var first *domain.PiholeConfig
	for id, nr := range results {
		if nr.Success && nr.Response != nil {
			cfg := nr.Response
			out.PerNode[id] = cfg
			if first == nil {
				first = cfg
				out.Consensus = cfg
			} else if !reflect.DeepEqual(*first, *cfg) {
				out.Drifted = true
			}
		}
	}
	return out
}

func (s *Service) PatchConfig(ctx context.Context, patch domain.PiholeConfigPatch) map[int64]*domain.NodeResult[struct{}] {
	results := s.cluster.PatchConfig(ctx, patch)

	nodeResults := make([]domain.AuditNodeResult, 0, len(results))
	for _, nr := range results {
		nodeResults = append(nodeResults, domain.AuditNodeResult{
			NodeId:   nr.PiholeNode.Id,
			NodeName: nr.PiholeNode.Name,
			Success:  nr.Success,
			Error:    util.ErrorString(nr.Error),
		})
	}
	s.audit.Record(ctx, domain.CreateAuditEntryParams{
		Action:      domain.AuditActionUpdateConfig,
		NodeResults: nodeResults,
	})

	return results
}
