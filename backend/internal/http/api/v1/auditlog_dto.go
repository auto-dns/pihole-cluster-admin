package v1

import (
	"time"

	"github.com/auto-dns/pihole-cluster-admin/internal/domain"
)

type auditEntryDTO struct {
	Id              int64             `json:"id"`
	Actor           string            `json:"actor"`
	Action          string            `json:"action"`
	TargetDomain    *string           `json:"targetDomain,omitempty"`
	RuleType        *string           `json:"ruleType,omitempty"`
	RuleKind        *string           `json:"ruleKind,omitempty"`
	BlockingEnabled *bool             `json:"blockingEnabled,omitempty"`
	BlockingTimer   *int              `json:"blockingTimer,omitempty"`
	TargetNodeId    *int64            `json:"targetNodeId,omitempty"`
	TargetNodeName  *string           `json:"targetNodeName,omitempty"`
	NodeResults     []auditNodeResult `json:"nodeResults"`
	CreatedAt       string            `json:"createdAt"`
}

type auditNodeResult struct {
	NodeId   int64  `json:"nodeId"`
	NodeName string `json:"nodeName"`
	Success  bool   `json:"success"`
	Error    string `json:"error,omitempty"`
}

type listAuditResponseDTO struct {
	Entries []auditEntryDTO `json:"entries"`
	Total   int             `json:"total"`
	Limit   int             `json:"limit"`
	Offset  int             `json:"offset"`
}

type rollbackResponseDTO struct {
	OriginalId int64                   `json:"originalId"`
	Nodes      []rollbackNodeResultDTO `json:"nodes"`
}

type rollbackNodeResultDTO struct {
	NodeId   int64  `json:"nodeId"`
	NodeName string `json:"nodeName"`
	Success  bool   `json:"success"`
	Error    string `json:"error,omitempty"`
}

func toAuditEntryDTO(e *domain.AuditEntry) auditEntryDTO {
	nodeResults := make([]auditNodeResult, 0, len(e.NodeResults))
	for _, nr := range e.NodeResults {
		nodeResults = append(nodeResults, auditNodeResult{
			NodeId:   nr.NodeId,
			NodeName: nr.NodeName,
			Success:  nr.Success,
			Error:    nr.Error,
		})
	}
	return auditEntryDTO{
		Id:              e.Id,
		Actor:           e.Actor,
		Action:          string(e.Action),
		TargetDomain:    e.TargetDomain,
		RuleType:        e.RuleType,
		RuleKind:        e.RuleKind,
		BlockingEnabled: e.BlockingEnabled,
		BlockingTimer:   e.BlockingTimer,
		TargetNodeId:    e.TargetNodeId,
		TargetNodeName:  e.TargetNodeName,
		NodeResults:     nodeResults,
		CreatedAt:       e.CreatedAt.UTC().Format(time.RFC3339),
	}
}
