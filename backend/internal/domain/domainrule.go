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
	Domain    string    `json:"domain"`
	Unicode   string    `json:"unicode"`
	Type      RuleType  `json:"type"`
	Kind      RuleKind  `json:"kind"`
	Comment   *string   `json:"comment,omitempty"`
	Groups    []int     `json:"groups"`
	Enabled   bool      `json:"enabled"`
	Id        int       `json:"id"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type DomainRuleSet struct {
	Rules []DomainRule  `json:"rules"`
	Took  time.Duration `json:"took"`
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

// Optional: result for “add” if you want echo-back
type AddDomainRulesResult struct {
	Rules []DomainRule  `json:"rules"`
	Took  time.Duration `json:"took"`
}
