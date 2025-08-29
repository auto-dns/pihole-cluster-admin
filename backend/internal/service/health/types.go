package health

import (
	"time"

	"github.com/auto-dns/pihole-cluster-admin/internal/domain"
)

type status string

const (
	statusOnline   status = "online"
	statusDegraded status = "degraded"
	statusOffline  status = "offline"
)

type nodeHealth struct {
	Id        int64
	Name      string
	Status    status
	Latency   time.Duration
	LastErr   string
	UpdatedAt time.Time
}

func ToDomainNodeHealth(n nodeHealth) domain.ClusterNodeHealth {
	return domain.ClusterNodeHealth{
		Id:        n.Id,
		Name:      n.Name,
		Status:    domain.ClusterHealthStatus(n.Status),
		Latency:   n.Latency,
		LastErr:   n.LastErr,
		UpdatedAt: n.UpdatedAt,
	}
}

type summary struct {
	Online    int
	Total     int
	UpdatedAt time.Time
}

func ToDomainHealthSummary(s summary) domain.ClusterHealthSummary {
	return domain.ClusterHealthSummary{
		Online:    s.Online,
		Total:     s.Total,
		UpdatedAt: s.UpdatedAt,
	}
}
