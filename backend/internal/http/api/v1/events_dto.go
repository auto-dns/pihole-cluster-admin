package v1

import (
	"time"

	"github.com/auto-dns/pihole-cluster-admin/internal/domain"
)

type healthSummaryDTO struct {
	Online    int    `json:"online"`
	Total     int    `json:"total"`
	UpdatedAt string `json:"updatedAt"`
}

func ToHealthSummaryDTO(s domain.ClusterHealthSummary) healthSummaryDTO {
	return healthSummaryDTO{
		Online:    s.Online,
		Total:     s.Total,
		UpdatedAt: s.UpdatedAt.UTC().Format(time.RFC3339),
	}
}

type nodeHealthDTO struct {
	Id        int64  `json:"id"`
	Name      string `json:"name"`
	Status    string `json:"status"`
	LatencyMs int64  `json:"latencyMs"`
	LastErr   string `json:"lastErr,omitempty"`
	UpdatedAt string `json:"updatedAt"`
}

func ToNodeHealthDTOs(in []domain.ClusterNodeHealth) []nodeHealthDTO {
	out := make([]nodeHealthDTO, 0, len(in))
	for _, n := range in {
		out = append(out, nodeHealthDTO{
			Id:        n.Id,
			Name:      n.Name,
			Status:    string(n.Status),
			LatencyMs: n.Latency.Milliseconds(),
			LastErr:   n.LastErr,
			UpdatedAt: n.UpdatedAt.UTC().Format(time.RFC3339),
		})
	}
	return out
}
