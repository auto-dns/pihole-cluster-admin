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
	SetBlockingState(ctx context.Context, blocking bool, timer *int) (*domain.BlockingState, error)
	FetchQueryLogs(ctx context.Context, req queriesWireRequest) (*domain.QueryLogPage, error)
	ListDomainRules(ctx context.Context, q domain.ListDomainRulesQuery) (*domain.DomainRuleSet, error)
	AddDomainRule(ctx context.Context, cmd domain.AddDomainRulesCommand) (*domain.AddDomainRulesResult, error)
	RemoveDomainRule(ctx context.Context, cmd domain.RemoveDomainRuleCommand) error
	ListAdlists(ctx context.Context) (*domain.AdlistSet, error)
	AddAdlist(ctx context.Context, cmd domain.AddAdlistCommand) (*domain.AdlistSet, error)
	UpdateAdlist(ctx context.Context, cmd domain.UpdateAdlistCommand) (*domain.AdlistSet, error)
	RemoveAdlist(ctx context.Context, cmd domain.RemoveAdlistCommand) error
	RebuildGravity(ctx context.Context) error
	FlushCache(ctx context.Context) error
	RestartDNS(ctx context.Context) error
	GetVersionInfo(ctx context.Context) (*domain.NodeVersionInfo, error)
	GetDatabaseInfo(ctx context.Context) (*domain.NodeDBInfo, error)
	AuthStatus(ctx context.Context) (*domain.AuthStatus, error)
	Logout(ctx context.Context) error
	GetStatsSummary(ctx context.Context) (*domain.StatsSummary, error)
	GetStatsHistory(ctx context.Context, from, until *int64) (*domain.StatsHistory, error)
	GetStatsTopDomains(ctx context.Context, from, until *int64, count *int) (*domain.StatsTopDomains, error)
	GetStatsTopClients(ctx context.Context, from, until *int64, count *int) (*domain.StatsTopClients, error)
	ListGroups(ctx context.Context) (*domain.GroupSet, error)
	AddGroup(ctx context.Context, cmd domain.AddGroupCommand) (*domain.GroupSet, error)
	UpdateGroup(ctx context.Context, cmd domain.UpdateGroupCommand) (*domain.GroupSet, error)
	RemoveGroup(ctx context.Context, cmd domain.RemoveGroupCommand) error
	ListPiholeClients(ctx context.Context) (*domain.PiholeClientSet, error)
	UpdatePiholeClient(ctx context.Context, cmd domain.UpdatePiholeClientCommand) (*domain.PiholeClientSet, error)
	RemovePiholeClient(ctx context.Context, cmd domain.RemovePiholeClientCommand) error
	SetPassword(ctx context.Context, newPassword string) error
	TestRegex(ctx context.Context, testDomain string) (*domain.RegexTestResult, error)
	GetConfig(ctx context.Context) (*domain.PiholeConfig, error)
	PatchConfig(ctx context.Context, patch domain.PiholeConfigPatch) error
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
