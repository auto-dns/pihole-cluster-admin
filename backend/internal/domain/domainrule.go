package domain

import "time"

type RuleType string

const (
	RuleTypeAllow RuleType = "allow"
	RuleTypeDeny  RuleType = "deny"
)

type RuleKind string

const (
	RuleKindExact RuleKind = "exact"
	RuleKindRegex RuleKind = "regex"
)

type DomainRule struct {
	Domain    string
	Unicode   string
	Type      RuleType
	Kind      RuleKind
	Comment   *string
	Groups    []int
	Enabled   bool
	Id        int
	CreatedAt time.Time
	UpdatedAt time.Time
}

type DomainRuleSet struct {
	Rules []DomainRule
	Took  time.Duration
}

// Queries/Commands your service & cluster expose
type ListDomainRulesQuery struct {
	Type   *RuleType
	Kind   *RuleKind
	Domain *string
}

type AddDomainRulesCommand struct {
	Type    RuleType
	Kind    RuleKind
	Domains []string
	Comment *string
	Groups  []int
	Enabled *bool
}

type RemoveDomainRuleCommand struct {
	Type   RuleType
	Kind   RuleKind
	Domain string
}

// Add Domain
type DomainProcessedItem struct {
	Item string
}

type DomainProcessedError struct {
	Item  string
	Error string
}

type DomainProcessed struct {
	Success []DomainProcessedItem
	Errors  []DomainProcessedError
}

type AddDomainRulesResult struct {
	Rules     []DomainRule
	Processed DomainProcessed
	Took      time.Duration
}

type SyncDomainRuleCommand struct {
	Type    RuleType
	Kind    RuleKind
	Domain  string
	Comment *string
}

type SyncDomainRuleNodeResult struct {
	AlreadyPresent bool
	Added          bool
}

type RegexMatch struct {
	ID      int
	Pattern string
	Kind    string // "deny" | "allow"
	Enabled bool
}

type RegexTestResult struct {
	Domain  string
	Matches []RegexMatch
}
