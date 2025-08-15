package api

import (
	"context"
	"net/http"

	"github.com/auto-dns/pihole-cluster-admin/internal/health"
	"github.com/auto-dns/pihole-cluster-admin/internal/pihole"
	"github.com/auto-dns/pihole-cluster-admin/internal/realtime"
	"github.com/auto-dns/pihole-cluster-admin/internal/store"
)

type eventSubscriber interface {
	Subscribe(topics []string) (<-chan realtime.Event, func())
}

type sessionAuth interface {
	AuthMiddleware(next http.Handler) http.Handler
}

type sessionIssuer interface {
	CreateSession(userId int64) (string, error)
	GetUserId(sessionID string) (int64, bool, error)
	DestroySession(sessionID string) error
	Cookie(value string) *http.Cookie
}

type sessionDeps interface {
	sessionAuth
	sessionIssuer
}

type healthService interface {
	NodeHealth() map[int64]health.NodeHealth
	Summary() health.Summary
}

type initStatusStore interface {
	GetInitializationStatus() (*store.InitializationStatus, error)
	SetUserCreated(userCreated bool) error
	SetPiholeStatus(piholeStatus store.PiholeStatus) error
}

type piholeStore interface {
	AddPiholeNode(params store.AddPiholeParams) (*store.PiholeNode, error)
	UpdatePiholeNode(id int64, params store.UpdatePiholeParams) (*store.PiholeNode, error)
	RemovePiholeNode(id int64) (found bool, err error)
	GetAllPiholeNodes() ([]*store.PiholeNode, error)
	GetPiholeNodeWithPassword(id int64) (*store.PiholeNode, error)
}

type userStore interface {
	CreateUser(params store.CreateUserParams) (*store.User, error)
	GetUser(id int64) (*store.User, error)
	ValidateUser(username, password string) (*store.User, error)
	IsInitialized() (bool, error)
}

type piholeCluster interface {
	AddClient(ctx context.Context, client *pihole.Client) error
	UpdateClient(ctx context.Context, id int64, cfg *pihole.ClientConfig) error
	HasClient(ctx context.Context, id int64) bool
	RemoveClient(ctx context.Context, id int64) error
	FetchQueryLogs(ctx context.Context, req pihole.FetchQueryLogClusterRequest) (*pihole.FetchQueryLogsClusterResponse, error)
	GetDomainRules(ctx context.Context, opts pihole.GetDomainRulesOptions) map[int64]*pihole.NodeResult[pihole.GetDomainRulesResponse]
	AddDomainRule(ctx context.Context, opts pihole.AddDomainRuleOptions) map[int64]*pihole.NodeResult[pihole.AddDomainRuleResponse]
	RemoveDomainRule(ctx context.Context, opts pihole.RemoveDomainRuleOptions) map[int64]*pihole.NodeResult[pihole.RemoveDomainRuleResponse]
}
