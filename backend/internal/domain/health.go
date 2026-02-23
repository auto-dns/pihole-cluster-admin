package domain

import "time"

type ClusterHealth struct {
	Summary   ClusterHealthSummary
	Nodes     map[int64]ClusterNodeHealth
	UpdatedAt time.Time
}

type ClusterHealthSummary struct {
	Online int
	Total  int
}

type ClusterNodeHealth struct {
	Id      int64
	Name    string
	Status  ClusterHealthStatus
	Latency time.Duration
	LastErr string
}

type ClusterHealthStatus string

const (
	ClusterHealthStatusOnline   ClusterHealthStatus = "online"
	ClusterHealthStatusDegraded ClusterHealthStatus = "degraded"
	ClusterHealthStatusOffline  ClusterHealthStatus = "offline"
)
