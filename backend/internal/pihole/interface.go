package pihole

type ClientInterface interface {
	GetNodeInfo() PiholeNode
	// Query logs
	FetchQueryLogs(opts FetchQueryLogOptions) (*FetchQueryLogResponse, error)
	// Domain management
	AddDomainRule(opts AddDomainRuleOptions) (*AddDomainRuleResponse, error)
	// GetDomainRules()
	RemoveDomainRule(opts RemoveDomainRuleOptions) error
	// Authorization
	Logout() error
}

type ClusterInterface interface {
	// Query logs
	FetchQueryLogs(opts FetchQueryLogOptions) []*NodeResult[FetchQueryLogResponse]
	// Domain management
	AddDomainRule(opts AddDomainRuleOptions) []*NodeResult[AddDomainRuleResponse]
	// GetDomainRules()
	RemoveDomainRule(opts RemoveDomainRuleOptions) []*NodeResult[RemoveDomainRuleResponse]
	// Authorization
	Logout() []error
}
