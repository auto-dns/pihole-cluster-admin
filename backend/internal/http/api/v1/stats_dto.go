package v1

import (
	"errors"
	"fmt"
	"net/url"
	"sort"
	"time"

	"github.com/auto-dns/pihole-cluster-admin/internal/domain"
	"github.com/auto-dns/pihole-cluster-admin/internal/errs"
	"github.com/auto-dns/pihole-cluster-admin/internal/util"
)

// -- Summary

type statsSummaryResponseDTO struct {
	Cluster statsSummaryDTO  `json:"cluster"`
	Nodes   []statsNodeDTO[statsSummaryDTO] `json:"nodes"`
}

type statsSummaryDTO struct {
	QueriesTotal   int64   `json:"queriesTotal"`
	QueriesBlocked int64   `json:"queriesBlocked"`
	BlockedPercent float64 `json:"blockedPercent"`
	GravitySize    int64   `json:"gravitySize"`
	UniqueClients  int64   `json:"uniqueClients"`
	UniqueDomains  int64   `json:"uniqueDomains"`
}

func statsSummaryFromDomain(s domain.StatsSummary) statsSummaryDTO {
	return statsSummaryDTO{
		QueriesTotal:   s.QueriesTotal,
		QueriesBlocked: s.QueriesBlocked,
		BlockedPercent: s.BlockedPercent,
		GravitySize:    s.GravitySize,
		UniqueClients:  s.UniqueClients,
		UniqueDomains:  s.UniqueDomains,
	}
}

func statsSummaryResponseFromDomain(res *domain.ClusterStatsSummary) statsSummaryResponseDTO {
	return statsSummaryResponseDTO{
		Cluster: statsSummaryFromDomain(res.Cluster),
		Nodes:   statsNodesFromDomain(res.Nodes, func(s *domain.StatsSummary) statsSummaryDTO {
			if s == nil {
				return statsSummaryDTO{}
			}
			return statsSummaryFromDomain(*s)
		}),
	}
}

// -- History

type statsHistoryResponseDTO struct {
	Cluster []statsHistoryPointDTO  `json:"cluster"`
	Nodes   []statsNodeDTO[[]statsHistoryPointDTO] `json:"nodes"`
}

type statsHistoryPointDTO struct {
	Timestamp string `json:"timestamp"`
	Total     int64  `json:"total"`
	Blocked   int64  `json:"blocked"`
}

func statsHistoryPointFromDomain(p domain.StatsHistoryPoint) statsHistoryPointDTO {
	return statsHistoryPointDTO{
		Timestamp: rfc3339(p.Timestamp),
		Total:     p.Total,
		Blocked:   p.Blocked,
	}
}

func statsHistoryPointsFromDomain(points []domain.StatsHistoryPoint) []statsHistoryPointDTO {
	out := make([]statsHistoryPointDTO, 0, len(points))
	for _, p := range points {
		out = append(out, statsHistoryPointFromDomain(p))
	}
	return out
}

func statsHistoryResponseFromDomain(res *domain.ClusterStatsHistory) statsHistoryResponseDTO {
	return statsHistoryResponseDTO{
		Cluster: statsHistoryPointsFromDomain(res.ClusterPoints),
		Nodes: statsNodesFromDomain(res.Nodes, func(h *domain.StatsHistory) []statsHistoryPointDTO {
			if h == nil {
				return nil
			}
			return statsHistoryPointsFromDomain(h.Points)
		}),
	}
}

// -- Top domains

type statsTopDomainsResponseDTO struct {
	ClusterTopQueried []topDomainDTO  `json:"clusterTopQueried"`
	ClusterTopBlocked []topDomainDTO  `json:"clusterTopBlocked"`
	Nodes             []statsNodeDTO[statsTopDomainsNodePayload] `json:"nodes"`
}

type statsTopDomainsNodePayload struct {
	TopQueried []topDomainDTO `json:"topQueried"`
	TopBlocked []topDomainDTO `json:"topBlocked"`
}

type topDomainDTO struct {
	Domain string `json:"domain"`
	Count  int64  `json:"count"`
}

func topDomainsFromDomain(entries []domain.TopDomainEntry) []topDomainDTO {
	out := make([]topDomainDTO, 0, len(entries))
	for _, e := range entries {
		out = append(out, topDomainDTO{Domain: e.Domain, Count: e.Count})
	}
	return out
}

func statsTopDomainsResponseFromDomain(res *domain.ClusterStatsTopDomains) statsTopDomainsResponseDTO {
	return statsTopDomainsResponseDTO{
		ClusterTopQueried: topDomainsFromDomain(res.ClusterTopQueried),
		ClusterTopBlocked: topDomainsFromDomain(res.ClusterTopBlocked),
		Nodes: statsNodesFromDomain(res.Nodes, func(d *domain.StatsTopDomains) statsTopDomainsNodePayload {
			if d == nil {
				return statsTopDomainsNodePayload{}
			}
			return statsTopDomainsNodePayload{
				TopQueried: topDomainsFromDomain(d.TopQueried),
				TopBlocked: topDomainsFromDomain(d.TopBlocked),
			}
		}),
	}
}

// -- Top clients

type statsTopClientsResponseDTO struct {
	ClusterClients []topClientDTO  `json:"clusterClients"`
	Nodes          []statsNodeDTO[[]topClientDTO] `json:"nodes"`
}

type topClientDTO struct {
	IP    string `json:"ip"`
	Name  string `json:"name"`
	Count int64  `json:"count"`
}

func topClientsFromDomain(entries []domain.TopClientEntry) []topClientDTO {
	out := make([]topClientDTO, 0, len(entries))
	for _, e := range entries {
		out = append(out, topClientDTO{IP: e.IP, Name: e.Name, Count: e.Count})
	}
	return out
}

func statsTopClientsResponseFromDomain(res *domain.ClusterStatsTopClients) statsTopClientsResponseDTO {
	return statsTopClientsResponseDTO{
		ClusterClients: topClientsFromDomain(res.ClusterClients),
		Nodes: statsNodesFromDomain(res.Nodes, func(c *domain.StatsTopClients) []topClientDTO {
			if c == nil {
				return nil
			}
			return topClientsFromDomain(c.Clients)
		}),
	}
}

// -- Generic node result wrapper

type statsNodeDTO[T any] struct {
	Node    piholeNodeRefDTO `json:"node"`
	Success bool             `json:"success"`
	Error   string           `json:"error,omitempty"`
	Data    T                `json:"data"`
}

func statsNodesFromDomain[R any, T any](
	nodes map[int64]*domain.NodeResult[R],
	extract func(R) T,
) []statsNodeDTO[T] {
	ids := make([]int64, 0, len(nodes))
	for id := range nodes {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })

	out := make([]statsNodeDTO[T], 0, len(nodes))
	for _, id := range ids {
		nr := nodes[id]
		out = append(out, statsNodeDTO[T]{
			Node: piholeNodeRefDTO{
				Id:   nr.PiholeNode.Id,
				Name: nr.PiholeNode.Name,
				Host: nr.PiholeNode.Host,
			},
			Success: nr.Success,
			Error:   util.ErrorString(nr.Error),
			Data:    extract(nr.Response),
		})
	}
	return out
}

// -- Query params

type statsTimeRangeParams struct {
	From  *int64
	Until *int64
	Count *int
}

func parseStatsTimeRangeParams(q url.Values) (statsTimeRangeParams, error) {
	var p statsTimeRangeParams

	now := time.Now().UTC()
	defaultFrom := now.Add(-24 * time.Hour).Unix()
	defaultUntil := now.Unix()
	p.From = &defaultFrom
	p.Until = &defaultUntil

	if r := q.Get("range"); r != "" {
		var dur time.Duration
		switch r {
		case "1h":
			dur = time.Hour
		case "6h":
			dur = 6 * time.Hour
		case "24h":
			dur = 24 * time.Hour
		default:
			return p, errs.Invalid(
				fmt.Sprintf("invalid range %q", r),
				errors.New("range must be one of: 1h, 6h, 24h"),
			)
		}
		from := now.Add(-dur).Unix()
		until := now.Unix()
		p.From = &from
		p.Until = &until
	}

	if v := q.Get("count"); v != "" {
		var n int
		if _, err := fmt.Sscanf(v, "%d", &n); err != nil || n <= 0 {
			return p, errs.Invalid("invalid count", errors.New("count must be a positive integer"))
		}
		p.Count = &n
	}

	return p, nil
}
