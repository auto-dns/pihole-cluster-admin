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
