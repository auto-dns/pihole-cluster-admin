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
	Id          int64
	Name        string
	Status      ClusterHealthStatus
	Latency     time.Duration
	LastErr     string
	VersionInfo *NodeVersionInfo
	DBInfo      *NodeDBInfo
}

type NodeVersionInfo struct {
	PiholeVersion string
	FTLVersion    string
}

type NodeDBInfo struct {
	GravityCount     int64
	GravityUpdatedAt *time.Time
}

type ClusterHealthStatus string

const (
	ClusterHealthStatusOnline   ClusterHealthStatus = "online"
	ClusterHealthStatusDegraded ClusterHealthStatus = "degraded"
	ClusterHealthStatusOffline  ClusterHealthStatus = "offline"
)
