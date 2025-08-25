package clusterblocking

import (
	"time"

	"github.com/auto-dns/pihole-cluster-admin/internal/domain"
)

type BlockingParams struct {
	Enable          bool   `json:"enable"`
	DurationSeconds *int64 `json:"durationSeconds"`
}

type BlockingSummary struct {
	Mode      string
	Unanimous bool
	Total     int
	Enabled   int
	Disabled  int
	Failed    int
	Errors    int
	MinTimer  *time.Duration
	MaxTimer  *time.Duration
	MaxTook   time.Duration
	AvgTook   time.Duration
}

type ClusterBlockingState struct {
	Summary BlockingSummary
	Nodes   map[int64]*domain.NodeResult[*domain.BlockingState]
}
