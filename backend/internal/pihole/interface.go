package pihole

import (
	"context"
	"time"

	"github.com/auto-dns/pihole-cluster-admin/internal/domain"
)

type clientPort interface {
	GetId(ctx context.Context) int64
	GetName(ctx context.Context) string
	GetScheme(ctx context.Context) string
	GetHost(ctx context.Context) string
	GetPort(ctx context.Context) int
	Update(ctx context.Context, cfg *ClientConfig)
	GetNodeInfo(ctx context.Context) domain.PiholeNodeRef
	GetBlockingState(ctx context.Context) (*domain.BlockingState, error)
	FetchQueryLogs(ctx context.Context, req fetchQueryLogClientRequest) (*FetchQueryLogResponse, error)
	GetAllDomainRules(ctx context.Context) (*GetDomainRulesResponse, error)
	GetDomainRulesByType(ctx context.Context, opts GetDomainRulesByTypeOptions) (*GetDomainRulesResponse, error)
	GetDomainRulesByKind(ctx context.Context, opts GetDomainRulesByKindOptions) (*GetDomainRulesResponse, error)
	GetDomainRulesByDomain(ctx context.Context, opts GetDomainRulesByDomainOptions) (*GetDomainRulesResponse, error)
	GetDomainRulesByTypeKind(ctx context.Context, opts GetDomainRulesByTypeKindOptions) (*GetDomainRulesResponse, error)
	GetDomainRulesByTypeKindDomain(ctx context.Context, opts GetDomainRulesByTypeKindDomainOptions) (*GetDomainRulesResponse, error)
	AddDomainRule(ctx context.Context, opts AddDomainRuleOptions) (*AddDomainRuleResponse, error)
	RemoveDomainRule(ctx context.Context, opts RemoveDomainRuleOptions) error
	AuthStatus(ctx context.Context) (*domain.AuthStatus, error)
	Logout(ctx context.Context) error
}

type cursorManagerPort[T any] interface {
	CreateCursor(requestParams T, piholeCursors map[int64]int) string
	GetSearchState(id string) (searchState searchStatePort[T], exists bool)
	Clear()
}

type searchStatePort[T any] interface {
	Expiration() time.Time
	GetRequestParams() T
	GetPiholeCursor(id int64) (int, bool)
}
