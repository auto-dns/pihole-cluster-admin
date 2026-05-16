package domain

import "time"

type AuditAction string

const (
	AuditActionAddDomainRule      AuditAction = "add_domain_rule"
	AuditActionRemoveDomainRule   AuditAction = "remove_domain_rule"
	AuditActionSetClusterBlocking AuditAction = "set_cluster_blocking"
	AuditActionSetNodeBlocking    AuditAction = "set_node_blocking"
)

type AuditNodeResult struct {
	NodeId   int64
	NodeName string
	Success  bool
	Error    string
}

type AuditEntry struct {
	Id              int64
	Actor           string
	Action          AuditAction
	TargetDomain    *string
	RuleType        *string
	RuleKind        *string
	BlockingEnabled *bool
	BlockingTimer   *int
	TargetNodeId    *int64
	TargetNodeName  *string
	NodeResults     []AuditNodeResult
	CreatedAt       time.Time
}

type CreateAuditEntryParams struct {
	Actor           string
	Action          AuditAction
	TargetDomain    *string
	RuleType        *string
	RuleKind        *string
	BlockingEnabled *bool
	BlockingTimer   *int
	TargetNodeId    *int64
	TargetNodeName  *string
	NodeResults     []AuditNodeResult
}

type ListAuditEntriesQuery struct {
	Limit  int
	Offset int
}
