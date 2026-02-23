package v1

import (
	"github.com/auto-dns/pihole-cluster-admin/internal/domain"
)

type clusterHealthDTO struct {
	Summary clusterHealthSummaryDTO        `json:"summary"`
	Nodes   map[int64]clusterNodeHealthDTO `json:"nodes"`
}

type clusterHealthSummaryDTO struct {
	Online int `json:"online"`
	Total  int `json:"total"`
}

type clusterNodeHealthDTO struct {
	Id        int64  `json:"id"`
	Name      string `json:"name"`
	Status    string `json:"status"`
	LatencyMS int    `json:"latencyMs"`
	LastErr   string `json:"lastErr,omitempty"`
}

func fromDomainToClusterHealthDTO(d domain.ClusterHealth) clusterHealthDTO {
	dto := clusterHealthDTO{
		Summary: clusterHealthSummaryDTO{
			Online: d.Summary.Online,
			Total:  d.Summary.Total,
		},
		Nodes: make(map[int64]clusterNodeHealthDTO, len(d.Nodes)),
	}

	for id, n := range d.Nodes {
		dto.Nodes[id] = clusterNodeHealthDTO{
			Id:        n.Id,
			Name:      n.Name,
			Status:    string(n.Status),
			LatencyMS: int(n.Latency.Milliseconds()),
			LastErr:   n.LastErr,
		}
	}

	return dto
}
