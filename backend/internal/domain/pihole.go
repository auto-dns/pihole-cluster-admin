package domain

import "time"

// Blocking

type BlockingStatus string

const (
	BlockingEnabled  BlockingStatus = "enabled"
	BlockingDisabled BlockingStatus = "disabled"
	BlockingFailed   BlockingStatus = "failed"
	BlockingUnknown  BlockingStatus = "unknown"
)

type BlockingState struct {
	Status    BlockingStatus
	TimerLeft *time.Duration
	Took      time.Duration
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
	Nodes   map[int64]*NodeResult[*BlockingState]
}
