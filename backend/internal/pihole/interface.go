package pihole

import (
	"context"
	"time"
)

type ClientInterface interface {
	GetId(ctx context.Context) int64
	GetName(ctx context.Context) string
	GetScheme(ctx context.Context) string
	GetHost(ctx context.Context) string
	GetPort(ctx context.Context) int
	Update(ctx context.Context, cfg *ClientConfig)
	// Node management
	GetNodeInfo(ctx context.Context) PiholeNode
	// API calls
	// -- Query logs
	FetchQueryLogs(ctx context.Context, req FetchQueryLogClientRequest) (*FetchQueryLogResponse, error)
	// -- Domain management
	GetDomainRules(ctx context.Context, opts GetDomainRulesOptions) (*GetDomainRulesResponse, error)
	AddDomainRule(ctx context.Context, opts AddDomainRuleOptions) (*AddDomainRuleResponse, error)
	RemoveDomainRule(ctx context.Context, opts RemoveDomainRuleOptions) error
	// -- Authorization
	Login(ctx context.Context) (*AuthResponse, error)
	AuthStatus(ctx context.Context) (*AuthResponse, error)
	Logout(ctx context.Context) error
}

type ClusterInterface interface {
	// Node management
	AddClient(ctx context.Context, client ClientInterface) error
	RemoveClient(ctx context.Context, id int64) error
	UpdateClient(ctx context.Context, id int64, cfg *ClientConfig) error
	HasClient(ctx context.Context, id int64) bool
	// API calls
	// -- Query logs
	FetchQueryLogs(ctx context.Context, req FetchQueryLogClusterRequest) (*FetchQueryLogsClusterResponse, error)
	// -- Domain management
	GetDomainRules(ctx context.Context, opts GetDomainRulesOptions) map[int64]*NodeResult[GetDomainRulesResponse]
	AddDomainRule(ctx context.Context, opts AddDomainRuleOptions) map[int64]*NodeResult[AddDomainRuleResponse]
	RemoveDomainRule(ctx context.Context, opts RemoveDomainRuleOptions) map[int64]*NodeResult[RemoveDomainRuleResponse]
	// -- Authorization
	AuthStatus(ctx context.Context) map[int64]*NodeResult[AuthResponse]
	Logout(ctx context.Context) map[int64]error
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
