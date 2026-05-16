package v1

import (
	"time"

	"github.com/auto-dns/pihole-cluster-admin/internal/domain"
)

// Generic

type domainRuleDTO struct {
	Id        int     `json:"id"`
	Domain    string  `json:"domain"`
	Unicode   string  `json:"unicode"`
	Type      string  `json:"type"`
	Kind      string  `json:"kind"`
	Comment   *string `json:"comment"`
	Groups    []int   `json:"groups"`
	Enabled   bool    `json:"enabled"`
	CreatedAt string  `json:"createdAt"`
	UpdatedAt string  `json:"updatedAt"`
}

// Get Domains

type listDomainRulesResponseDTO struct {
	Summary listSummaryDTO        `json:"summary"`
	Nodes   map[int64]listNodeDTO `json:"nodes"`
}

type listSummaryDTO struct {
	TotalNodes int `json:"totalNodes"`
	OkNodes    int `json:"okNodes"`
	ErrorNodes int `json:"errorNodes"`
	TotalRules int `json:"totalRules"`
}

type listNodeDTO struct {
	Node   piholeNodeRefDTO `json:"node"`
	Rules  []domainRuleDTO  `json:"rules"`
	TookMS int64            `json:"tookMs"`
	Error  string           `json:"error,omitempty"`
}

// Add Domain
type addDomainRuleRequestDTO struct {
	Domain  any     `json:"domain"` // string or []string (transport)
	Comment *string `json:"comment,omitempty"`
	Groups  []int   `json:"groups,omitempty"`
	Enabled *bool   `json:"enabled,omitempty"`
}

type addDomainRuleResponseDTO struct {
	Nodes map[int64]addDomainRuleNodeDTO `json:"nodes"`
}

type addDomainRuleNodeDTO struct {
	Node   piholeNodeRefDTO       `json:"node"`
	Result addDomainRuleResultDTO `json:"result"`
	Error  string                 `json:"error,omitempty"`
}

type addDomainRuleResultDTO struct {
	Domains   []domainRuleDTO    `json:"domains"`
	Processed domainProcessedDTO `json:"processed"`
	TookMS    int64              `json:"tookMs"`
}

type domainProcessedDTO struct {
	Success []domainProcessedSuccessDTO `json:"success"`
	Errors  []domainProcessedErrorDTO   `json:"errors"`
}

type domainProcessedSuccessDTO struct {
	Item string `json:"item"`
}

type domainProcessedErrorDTO struct {
	Item  string `json:"item"`
	Error string `json:"error"`
}

func toDomainRuleDTO(d domain.DomainRule) domainRuleDTO {
	return domainRuleDTO{
		Id:        d.Id,
		Domain:    d.Domain,
		Unicode:   d.Unicode,
		Type:      string(d.Type),
		Kind:      string(d.Kind),
		Comment:   d.Comment,
		Groups:    d.Groups,
		Enabled:   d.Enabled,
		CreatedAt: d.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt: d.UpdatedAt.UTC().Format(time.RFC3339),
	}
}

func toProcessedDTO(p domain.DomainProcessed) domainProcessedDTO {
	out := domainProcessedDTO{}
	if len(p.Success) > 0 {
		out.Success = make([]domainProcessedSuccessDTO, 0, len(p.Success))
		for _, s := range p.Success {
			out.Success = append(out.Success, domainProcessedSuccessDTO{Item: s.Item})
		}
	}
	if len(p.Errors) > 0 {
		out.Errors = make([]domainProcessedErrorDTO, 0, len(p.Errors))
		for _, e := range p.Errors {
			out.Errors = append(out.Errors, domainProcessedErrorDTO{
				Item:  e.Item,
				Error: e.Error,
			})
		}
	}
	return out
}

// Sync parity

type syncDomainRuleRequestDTO struct {
	Type    string  `json:"type"`
	Kind    string  `json:"kind"`
	Domain  string  `json:"domain"`
	Comment *string `json:"comment,omitempty"`
}

type syncDomainRuleResponseDTO struct {
	Summary syncSummaryDTO                  `json:"summary"`
	Nodes   map[int64]syncDomainRuleNodeDTO `json:"nodes"`
}

type syncSummaryDTO struct {
	TotalNodes          int `json:"totalNodes"`
	SyncedNodes         int `json:"syncedNodes"`
	AlreadyPresentNodes int `json:"alreadyPresentNodes"`
	FailedNodes         int `json:"failedNodes"`
}

type syncDomainRuleNodeDTO struct {
	Node           piholeNodeRefDTO `json:"node"`
	AlreadyPresent bool             `json:"alreadyPresent"`
	Added          bool             `json:"added"`
	Error          string           `json:"error,omitempty"`
}

// Remove domain

type removeDomainRuleResponseDTO struct {
	Summary removeSummaryDTO                  `json:"summary"`
	Nodes   map[int64]removeDomainRuleNodeDTO `json:"nodes"`
}

type removeSummaryDTO struct {
	Total   int `json:"total"`
	Removed int `json:"removed"`
	Failed  int `json:"failed"`
	Errors  int `json:"errors"`
}

type removeDomainRuleNodeDTO struct {
	Node    piholeNodeRefDTO `json:"node"`
	Removed bool             `json:"removed"`
	Error   string           `json:"error,omitempty"`
}
