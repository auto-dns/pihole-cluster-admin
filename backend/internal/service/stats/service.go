package statssvc

import (
	"context"
	"sort"
	"time"

	"github.com/auto-dns/pihole-cluster-admin/internal/domain"
	"github.com/rs/zerolog"
)

type Service struct {
	cluster cluster
	logger  zerolog.Logger
}

func NewService(cluster cluster, logger zerolog.Logger) *Service {
	return &Service{cluster: cluster, logger: logger}
}

func (s *Service) GetSummary(ctx context.Context) (*domain.ClusterStatsSummary, error) {
	nodes := s.cluster.GetStatsSummary(ctx)
	var agg domain.StatsSummary
	for _, nr := range nodes {
		if !nr.Success || nr.Response == nil {
			continue
		}
		r := nr.Response
		agg.QueriesTotal += r.QueriesTotal
		agg.QueriesBlocked += r.QueriesBlocked
		if r.GravitySize > agg.GravitySize {
			agg.GravitySize = r.GravitySize
		}
		agg.UniqueClients += r.UniqueClients
		agg.UniqueDomains += r.UniqueDomains
	}
	if agg.QueriesTotal > 0 {
		agg.BlockedPercent = float64(agg.QueriesBlocked) / float64(agg.QueriesTotal) * 100
	}
	return &domain.ClusterStatsSummary{Cluster: agg, Nodes: nodes}, nil
}

func (s *Service) GetHistory(ctx context.Context, from, until *int64) (*domain.ClusterStatsHistory, error) {
	nodes := s.cluster.GetStatsHistory(ctx, from, until)

	// Aggregate by aligning on timestamp buckets
	type bucket struct{ total, blocked int64 }
	buckets := make(map[int64]*bucket)
	for _, nr := range nodes {
		if !nr.Success || nr.Response == nil {
			continue
		}
		for _, p := range nr.Response.Points {
			ts := p.Timestamp.Unix()
			if buckets[ts] == nil {
				buckets[ts] = &bucket{}
			}
			buckets[ts].total += p.Total
			buckets[ts].blocked += p.Blocked
		}
	}

	points := make([]domain.StatsHistoryPoint, 0, len(buckets))
	for ts, b := range buckets {
		points = append(points, domain.StatsHistoryPoint{
			Timestamp: time.Unix(ts, 0).UTC(),
			Total:     b.total,
			Blocked:   b.blocked,
		})
	}
	sort.Slice(points, func(i, j int) bool { return points[i].Timestamp.Before(points[j].Timestamp) })
	return &domain.ClusterStatsHistory{ClusterPoints: points, Nodes: nodes}, nil
}

func (s *Service) GetTopDomains(ctx context.Context, from, until *int64, count *int) (*domain.ClusterStatsTopDomains, error) {
	nodes := s.cluster.GetStatsTopDomains(ctx, from, until, count)

	queriedMap := make(map[string]int64)
	blockedMap := make(map[string]int64)
	for _, nr := range nodes {
		if !nr.Success || nr.Response == nil {
			continue
		}
		for _, d := range nr.Response.TopQueried {
			queriedMap[d.Domain] += d.Count
		}
		for _, d := range nr.Response.TopBlocked {
			blockedMap[d.Domain] += d.Count
		}
	}

	limit := 10
	if count != nil && *count > 0 {
		limit = *count
	}

	// Note: each node returns its own top-N list, so a domain ranked outside the top N
	// on every individual node but top-N cluster-wide can be missed. This is an accepted
	// approximation; the alternative would require fetching all domains from every node.
	return &domain.ClusterStatsTopDomains{
		ClusterTopQueried: sortedDomains(queriedMap, limit),
		ClusterTopBlocked: sortedDomains(blockedMap, limit),
		Nodes:             nodes,
	}, nil
}

func (s *Service) GetTopClients(ctx context.Context, from, until *int64, count *int) (*domain.ClusterStatsTopClients, error) {
	nodes := s.cluster.GetStatsTopClients(ctx, from, until, count)

	type clientKey struct{ ip, name string }
	clientMap := make(map[clientKey]int64)
	for _, nr := range nodes {
		if !nr.Success || nr.Response == nil {
			continue
		}
		for _, c := range nr.Response.Clients {
			clientMap[clientKey{ip: c.IP, name: c.Name}] += c.Count
		}
	}

	limit := 10
	if count != nil && *count > 0 {
		limit = *count
	}

	entries := make([]domain.TopClientEntry, 0, len(clientMap))
	for k, cnt := range clientMap {
		entries = append(entries, domain.TopClientEntry{IP: k.ip, Name: k.name, Count: cnt})
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].Count > entries[j].Count })
	if len(entries) > limit {
		entries = entries[:limit]
	}

	return &domain.ClusterStatsTopClients{ClusterClients: entries, Nodes: nodes}, nil
}

func sortedDomains(m map[string]int64, limit int) []domain.TopDomainEntry {
	out := make([]domain.TopDomainEntry, 0, len(m))
	for d, c := range m {
		out = append(out, domain.TopDomainEntry{Domain: d, Count: c})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Count > out[j].Count })
	if len(out) > limit {
		out = out[:limit]
	}
	return out
}
