package domain

import "time"

type ClusterHealthStatus string

const (
	ClusterHealthStatusOnline   ClusterHealthStatus = "online"
	ClusterHealthStatusDegraded ClusterHealthStatus = "degraded"
	ClusterHealthStatusOffline  ClusterHealthStatus = "offline"
)

type ClusterNodeHealth struct {
	Id        int64
	Name      string
	Status    ClusterHealthStatus
	Latency   time.Duration
	LastErr   string
	UpdatedAt time.Time
}

type ClusterHealthSummary struct {
	Online    int
	Total     int
	UpdatedAt time.Time
}
