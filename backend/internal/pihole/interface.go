package pihole

import "time"

type ClientInterface interface {
	GetId() int64
	GetName() string
	GetScheme() string
	GetHost() string
	GetPort() int
	Update(cfg *ClientConfig)
	// Node management
	GetNodeInfo() PiholeNode
	// API calls
	// -- Query logs
	FetchQueryLogs(req FetchQueryLogClientRequest) (*FetchQueryLogResponse, error)
	// -- Domain management
	GetDomainRules(opts GetDomainRulesOptions) (*GetDomainRulesResponse, error)
	AddDomainRule(opts AddDomainRuleOptions) (*AddDomainRuleResponse, error)
	RemoveDomainRule(opts RemoveDomainRuleOptions) error
	// -- Authorization
	Logout() error
}

type ClusterInterface interface {
	// Node management
	AddClient(client ClientInterface) error
	RemoveClient(id int64) error
	UpdateClient(id int64, cfg *ClientConfig) error
	HasClient(id int64) bool
	// API calls
	// -- Query logs
	FetchQueryLogs(req FetchQueryLogClusterRequest) (*FetchQueryLogsClusterResponse, error)
	// -- Domain management
	GetDomainRules(opts GetDomainRulesOptions) map[int64]*NodeResult[GetDomainRulesResponse]
	AddDomainRule(opts AddDomainRuleOptions) map[int64]*NodeResult[AddDomainRuleResponse]
	RemoveDomainRule(opts RemoveDomainRuleOptions) map[int64]*NodeResult[RemoveDomainRuleResponse]
	// -- Authorization
	Logout() map[int64]error
}

type CursorManagerInterface[T any] interface {
	CreateCursor(requestParams T, piholeCursors map[int64]int) string
	GetSearchState(id string) (searchState SearchStateInterface[T], exists bool)
	Clear()
}

type SearchStateInterface[T any] interface {
	Expiration() time.Time
	GetRequestParams() T
	GetPiholeCursor(id int64) (int, bool)
}
