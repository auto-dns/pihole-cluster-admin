package v1

type addDomainRuleResponseDTO struct {
	Domain  any     `json:"domain"` // string or []string (transport)
	Comment *string `json:"comment,omitempty"`
	Groups  []int   `json:"groups,omitempty"`
	Enabled *bool   `json:"enabled,omitempty"`
}
