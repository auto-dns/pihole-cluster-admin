package statssvc

import (
	"context"

	"github.com/auto-dns/pihole-cluster-admin/internal/domain"
)

type cluster interface {
	GetStatsSummary(ctx context.Context) map[int64]*domain.NodeResult[*domain.StatsSummary]
	GetStatsHistory(ctx context.Context, from, until *int64) map[int64]*domain.NodeResult[*domain.StatsHistory]
	GetStatsTopDomains(ctx context.Context, from, until *int64, count *int) map[int64]*domain.NodeResult[*domain.StatsTopDomains]
	GetStatsTopClients(ctx context.Context, from, until *int64, count *int) map[int64]*domain.NodeResult[*domain.StatsTopClients]
}
