package v1

import (
	"context"
	"net/http"

	"github.com/auto-dns/pihole-cluster-admin/internal/domain"
	"github.com/auto-dns/pihole-cluster-admin/internal/realtime"
	auth_s "github.com/auto-dns/pihole-cluster-admin/internal/service/auth"
	pihole_s "github.com/auto-dns/pihole-cluster-admin/internal/service/pihole"
)

type AuthService interface {
	Login(params auth_s.LoginCommand) (*domain.User, string, error)
	Logout(sessionId string) error
	GetUser(id int64) (*domain.User, error)
}

type ClusterBlockingService interface {
	GetState(ctx context.Context) (*domain.ClusterBlockingState, error)
}

type DomainRuleService interface {
	List(ctx context.Context, q domain.ListDomainRulesQuery) map[int64]*domain.NodeResult[*domain.DomainRuleSet]
	Add(ctx context.Context, cmd domain.AddDomainRulesCommand) map[int64]*domain.NodeResult[*domain.AddDomainRulesResult]
	Remove(ctx context.Context, cmd domain.RemoveDomainRuleCommand) map[int64]*domain.NodeResult[struct{}]
}

type EventsService interface {
	Subscribe(ctx context.Context, topics []string) (<-chan realtime.Event, func())
}

type HealthService interface {
	GetSummary() domain.ClusterHealthSummary
	GetNodeHealth() map[int64]domain.ClusterNodeHealth
}

type PiholeService interface {
	GetAll() ([]*domain.PiholeNode, error)
	Add(ctx context.Context, params pihole_s.AddNodeCommand) (*domain.PiholeNode, error)
	Update(ctx context.Context, id int64, params pihole_s.UpdateNodeCommand) (*domain.PiholeNode, error)
	Remove(ctx context.Context, id int64) (found bool, err error)
	TestExistingConnection(ctx context.Context, id int64, params pihole_s.TestExistingConnectionCommand) error
	TestInstanceConnection(ctx context.Context, params pihole_s.TestInstanceConnectionCommand) error
}

type QueryLogService interface {
	Fetch(ctx context.Context, req domain.QueryLogQuery) (*domain.ClusterQueryLogResponse, error)
}

type httpCookieFactory interface {
	Cookie(value string) *http.Cookie
	CookieName() string
}
