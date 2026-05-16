package domain

import "time"

type StatsSummary struct {
	QueriesTotal   int64
	QueriesBlocked int64
	BlockedPercent float64
	GravitySize    int64
	UniqueClients  int64
	UniqueDomains  int64
	Took           time.Duration
}

type StatsHistoryPoint struct {
	Timestamp time.Time
	Total     int64
	Blocked   int64
}

type StatsHistory struct {
	Points []StatsHistoryPoint
	Took   time.Duration
}

type TopDomainEntry struct {
	Domain string
	Count  int64
}

type StatsTopDomains struct {
	TopQueried []TopDomainEntry
	TopBlocked []TopDomainEntry
	Took       time.Duration
}

type TopClientEntry struct {
	IP    string
	Name  string
	Count int64
}

type StatsTopClients struct {
	Clients []TopClientEntry
	Took    time.Duration
}

type ClusterStatsSummary struct {
	Cluster StatsSummary
	Nodes   map[int64]*NodeResult[*StatsSummary]
}

type ClusterStatsHistory struct {
	ClusterPoints []StatsHistoryPoint
	Nodes         map[int64]*NodeResult[*StatsHistory]
}

type ClusterStatsTopDomains struct {
	ClusterTopQueried []TopDomainEntry
	ClusterTopBlocked []TopDomainEntry
	Nodes             map[int64]*NodeResult[*StatsTopDomains]
}

type ClusterStatsTopClients struct {
	ClusterClients []TopClientEntry
	Nodes          map[int64]*NodeResult[*StatsTopClients]
}
