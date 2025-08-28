package v1

import (
	"time"

	"github.com/auto-dns/pihole-cluster-admin/internal/domain"
)

type clusterNodeHealthDTO struct {
	Id        int64     `json:"id"`
	Name      string    `json:"name"`
	Status    string    `json:"status"`
	LatencyMS int       `json:"latencyMs"`
	LastErr   string    `json:"lastErr,omitempty"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func FromDomainClusterNodeHealth(d domain.ClusterNodeHealth) clusterNodeHealthDTO {
	return clusterNodeHealthDTO{
		Id:        d.Id,
		Name:      d.Name,
		Status:    string(d.Status),
		LatencyMS: int(d.Latency.Milliseconds()),
		LastErr:   d.LastErr,
		UpdatedAt: d.UpdatedAt,
	}
}

type clusterHealthSummaryDTO struct {
	Online    int       `json:"online"`
	Total     int       `json:"total"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func FromDomainClusterHealthSummary(d domain.ClusterHealthSummary) clusterHealthSummaryDTO {
	return clusterHealthSummaryDTO{
		Online:    d.Online,
		Total:     d.Total,
		UpdatedAt: d.UpdatedAt,
	}
}
