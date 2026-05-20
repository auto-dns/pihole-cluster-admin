package configsvc

import (
	"context"
	"reflect"
	"sort"

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

	// Populate PerNode for all successful results.
	for id, nr := range results {
		if nr.Success && nr.Response != nil {
			out.PerNode[id] = nr.Response
		}
	}

	// Determine consensus using the lowest node ID so the result is stable
	// across calls (map iteration order is non-deterministic).
	ids := make([]int64, 0, len(out.PerNode))
	for id := range out.PerNode {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })

	for _, id := range ids {
		cfg := out.PerNode[id]
		if out.Consensus == nil {
			out.Consensus = cfg
		} else if !reflect.DeepEqual(*out.Consensus, *cfg) {
			out.Drifted = true
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
