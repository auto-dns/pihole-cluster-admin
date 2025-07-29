package pihole

type PiholeNode struct {
	ID   string `json:"id"`
	Host string `json:"host"`
}

type NodeResult[T any] struct {
	PiholeNode PiholeNode `json:"piholeNode"`
	Success    bool       `json:"success"`
	Error      string     `json:"error,omitempty"`
	Response   *T         `json:"response,omitempty"`
}

type DomainInfo struct {
	Domain       string  `json:"domain"`
	Unicode      string  `json:"unicode"`
	Type         string  `json:"type"` // "allow" or "deny"
	Kind         string  `json:"kind"` // "exact" or "regex"
	Comment      *string `json:"comment,omitempty"`
	Groups       []int   `json:"groups"`
	Enabled      bool    `json:"enabled"`
	ID           int     `json:"id"`
	DateAdded    int64   `json:"date_added"`
	DateModified int64   `json:"date_modified"`
}

// Arguments and responses

// FetchQueryLogs query options

// -- Request

type FetchQueryLogRequest struct {
	Filters  FetchQueryLogFilters
	CursorID *string
	Length   *int // number of results
	Start    *int // offset
}

type FetchQueryLogFilters struct {
	From       *int64  // Unix timestamp
	Until      *int64  // Unix timestamp
	Domain     *string // filter by domain
	ClientIP   *string // filter by client IP
	ClientName *string // filter by client hostname
	Upstream   *string // filter by upstream server
	Type       *string // query type (A, AAAA, etc.)
	Status     *string // query status (GRAVITY, FORWARDED, etc.)
	Reply      *string // reply type (NODATA, NXDOMAIN, etc.)
	DNSSEC     *string // DNSSEC status (SECURE, INSECURE, etc.)
	Disk       *bool   // load from on-disk database
}

// -- Response

type FetchQueryLogsClusterResponse struct {
	CursorID     string                               `json:"cursor"`
	Results      []*NodeResult[FetchQueryLogResponse] `json:"results"`
	EndOfResults bool                                 `json:"endOfResults"`
}

type FetchQueryLogResponse struct {
	Queries         []DNSLogEntry `json:"queries"`
	Cursor          int64         `json:"cursor"`
	RecordsTotal    int64         `json:"recordsTotal"`
	RecordsFiltered int64         `json:"recordsFiltered"`
	Draw            int64         `json:"draw"`
	Took            float64       `json:"took"`
}

type DNSLogEntry struct {
	ID       int64      `json:"id"`
	Time     float64    `json:"time"`
	Type     string     `json:"type"`
	Status   string     `json:"status"`
	DNSSEC   string     `json:"dnssec"`
	Domain   string     `json:"domain"`
	Upstream *string    `json:"upstream"`
	Reply    ReplyInfo  `json:"reply"`
	Client   ClientInfo `json:"client"`
	ListID   *int64     `json:"list_id"`
	EDE      EDEInfo    `json:"ede"`
	CNAME    *string    `json:"cname"`
}

type ReplyInfo struct {
	Type string  `json:"type"`
	Time float64 `json:"time"`
}

type ClientInfo struct {
	IP   string  `json:"ip"`
	Name *string `json:"name"`
}

type EDEInfo struct {
	Code int64   `json:"code"`
	Text *string `json:"text"`
}

// GetDomainRules

// -- Request

type GetDomainRulesOptions struct {
	Type   *string // "allow" or "deny"
	Kind   *string // "exact" or "regex"
	Domain *string //domain filter
}

// -- Response

type GetDomainRulesResponse struct {
	Domains []DomainInfo `json:"domains"`
	Took    float64      `json:"took"`
}

// AddDomainRule params

// -- Request
type AddDomainPayload struct {
	Domain  interface{} `json:"domain"`            // string OR []string
	Comment *string     `json:"comment,omitempty"` // optional
	Groups  []int       `json:"groups,omitempty"`  // optional, default empty
	Enabled *bool       `json:"enabled,omitempty"` // optional, default true
}

type AddDomainRuleOptions struct {
	Type    string           // "allow" or "deny"
	Kind    string           // "exact" or "regex"
	Payload AddDomainPayload // request body
}

// -- Response

type ProcessedResult struct {
	Success []struct {
		Item string `json:"item"`
	} `json:"success"`
	Errors []struct {
		Item  string `json:"item"`
		Error string `json:"error"`
	} `json:"errors"`
}

type AddDomainRuleResponse struct {
	Domains   []DomainInfo     `json:"domains"`
	Processed *ProcessedResult `json:"processed,omitempty"`
	Took      float64          `json:"took"`
}

// RemoveDomainRule params

// -- Request

type RemoveDomainRuleOptions struct {
	Type   string // "allow" or "deny"
	Kind   string // "exact" or "regex"
	Domain string // a single domain to remove
}

// -- Response

// RemoveDomainRuleResponse is intentionally empty because Pi-hole returns no body.
// It exists only so we have a concrete T type for NodeResult.
type RemoveDomainRuleResponse struct{}
