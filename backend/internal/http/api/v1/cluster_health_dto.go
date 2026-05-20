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
	Id               int64  `json:"id"`
	Name             string `json:"name"`
	Status           string `json:"status"`
	LatencyMS        int    `json:"latencyMs"`
	LastErr          string `json:"lastErr,omitempty"`
	PiholeVersion    string `json:"piholeVersion,omitempty"`
	FTLVersion       string `json:"ftlVersion,omitempty"`
	GravityCount     *int64 `json:"gravityCount,omitempty"`
	GravityUpdatedAt *int64 `json:"gravityUpdatedAt,omitempty"`
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
		node := clusterNodeHealthDTO{
			Id:        n.Id,
			Name:      n.Name,
			Status:    string(n.Status),
			LatencyMS: int(n.Latency.Milliseconds()),
			LastErr:   n.LastErr,
		}
		if n.VersionInfo != nil {
			node.PiholeVersion = n.VersionInfo.PiholeVersion
			node.FTLVersion = n.VersionInfo.FTLVersion
		}
		if n.DBInfo != nil {
			node.GravityCount = &n.DBInfo.GravityCount
			if n.DBInfo.GravityUpdatedAt != nil {
				v := n.DBInfo.GravityUpdatedAt.Unix()
				node.GravityUpdatedAt = &v
			}
		}
		dto.Nodes[id] = node
	}

	return dto
}
