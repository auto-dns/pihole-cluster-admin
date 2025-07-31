package pihole

type ClientInterface interface {
	GetNodeInfo() PiholeNode
	// Query logs
	FetchQueryLogs(req FetchQueryLogClientRequest) (*FetchQueryLogResponse, error)
	// Domain management
	GetDomainRules(opts GetDomainRulesOptions) (*GetDomainRulesResponse, error)
	AddDomainRule(opts AddDomainRuleOptions) (*AddDomainRuleResponse, error)
	RemoveDomainRule(opts RemoveDomainRuleOptions) error
	// Authorization
	Logout() error
}

type ClusterInterface interface {
	// Query logs
	FetchQueryLogs(req FetchQueryLogClusterRequest) (*FetchQueryLogsClusterResponse, error)
	// Domain management
	GetDomainRules(opts GetDomainRulesOptions) []*NodeResult[GetDomainRulesResponse]
	AddDomainRule(opts AddDomainRuleOptions) []*NodeResult[AddDomainRuleResponse]
	RemoveDomainRule(opts RemoveDomainRuleOptions) []*NodeResult[RemoveDomainRuleResponse]
	// Authorization
	Logout() []error
}
