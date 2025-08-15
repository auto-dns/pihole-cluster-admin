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
	FetchQueryLogs(ctx context.Context, req FetchQueryLogClientRequest) (*FetchQueryLogResponse, error)
	GetDomainRules(ctx context.Context, opts GetDomainRulesOptions) (*GetDomainRulesResponse, error)
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
