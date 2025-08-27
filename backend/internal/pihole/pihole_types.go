package pihole

// Query log
// -- Request
type queriesWireRequest struct {
	Filters queriesWireFilters
	Cursor  *int
	Length  *int
	Start   *int
}

type queriesWireFilters struct {
	From, Until *int64
	Domain      *string
	ClientIP    *string
	ClientName  *string
	Upstream    *string
	Type        *string
	Status      *string
	Reply       *string
	DNSSEC      *string
	Disk        *bool
}

type queriesWireResponse struct {
	Queries         []queryWireEntry `json:"queries"`
	Cursor          int              `json:"cursor"`
	RecordsTotal    int64            `json:"recordsTotal"`
	RecordsFiltered int64            `json:"recordsFiltered"`
	Draw            int64            `json:"draw"`
	Took            float64          `json:"took"`
}

type queryWireEntry struct {
	ID       int64   `json:"id"`
	Time     float64 `json:"time"`
	Type     string  `json:"type"`
	Status   string  `json:"status"`
	DNSSEC   string  `json:"dnssec"`
	Domain   string  `json:"domain"`
	Upstream *string `json:"upstream"`
	Reply    struct {
		Type string  `json:"type"`
		Time float64 `json:"time"`
	} `json:"reply"`
	Client struct {
		IP   string  `json:"ip"`
		Name *string `json:"name"`
	} `json:"client"`
	ListID *int64 `json:"list_id"`
	EDE    struct {
		Code int64   `json:"code"`
		Text *string `json:"text"`
	} `json:"ede"`
	CNAME *string `json:"cname"`
}

// -- Response

type DomainInfo struct {
	Domain       string  `json:"domain"`
	Unicode      string  `json:"unicode"`
	Type         string  `json:"type"` // "allow" or "deny"
	Kind         string  `json:"kind"` // "exact" or "regex"
	Comment      *string `json:"comment,omitempty"`
	Groups       []int   `json:"groups"`
	Enabled      bool    `json:"enabled"`
	Id           int     `json:"id"`
	DateAdded    int64   `json:"date_added"`
	DateModified int64   `json:"date_modified"`
}

type ProcessedResult struct {
	Success []struct {
		Item string `json:"item"`
	} `json:"success"`
	Errors []struct {
		Item  string `json:"item"`
		Error string `json:"error"`
	} `json:"errors"`
}

type GetDomainRulesByTypeOptions struct {
	Type RuleType
}

type GetDomainRulesByKindOptions struct {
	Kind RuleKind
}

type GetDomainRulesByDomainOptions struct {
	Domain string
}

type GetDomainRulesByTypeKindOptions struct {
	Type RuleType
	Kind RuleKind
}

type GetDomainRulesByTypeKindDomainOptions struct {
	Type   RuleType
	Kind   RuleKind
	Domain string
}

type GetDomainRulesResponse struct {
	Domains []DomainInfo `json:"domains"`
	Took    float64      `json:"took"`
}

type AddDomainPayload struct {
	Domain  interface{} `json:"domain"`            // string OR []string
	Comment *string     `json:"comment,omitempty"` // optional
	Groups  []int       `json:"groups,omitempty"`  // optional, default empty
	Enabled *bool       `json:"enabled,omitempty"` // optional, default true
}

type AddDomainRuleOptions struct {
	Type    RuleType
	Kind    RuleKind
	Payload AddDomainPayload // request body
}

type AddDomainRuleResponse struct {
	Domains   []DomainInfo     `json:"domains"`
	Processed *ProcessedResult `json:"processed,omitempty"`
	Took      float64          `json:"took"`
}

type RemoveDomainRuleOptions struct {
	Type   RuleType
	Kind   RuleKind
	Domain string // a single domain to remove
}

// RemoveDomainRuleResponse is intentionally empty because Pi-hole returns no body.
// It exists only so we have a concrete T type for NodeResult.
type RemoveDomainRuleResponse struct{}
