package health

import health_s "github.com/auto-dns/pihole-cluster-admin/internal/service/health"

type service interface {
	GetSummary() health_s.Summary
	GetNodeHealth() map[int64]health_s.NodeHealth
}
