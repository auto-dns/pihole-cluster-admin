package store

import "time"

type auditLogRow struct {
	Id              int64
	Actor           string
	Action          string
	TargetDomain    *string
	RuleType        *string
	RuleKind        *string
	BlockingEnabled *int64
	BlockingTimer   *int
	TargetNodeId    *int64
	TargetNodeName  *string
	NodeResults     string
	CreatedAt       time.Time
}
