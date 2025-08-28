package v1

import (
	"context"
	"net/http"

	"github.com/auto-dns/pihole-cluster-admin/internal/domain"
	"github.com/auto-dns/pihole-cluster-admin/internal/realtime"
	auth_s "github.com/auto-dns/pihole-cluster-admin/internal/service/auth"
)

type AuthService interface {
	Login(params auth_s.LoginParams) (*domain.User, string, error)
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

type httpCookieFactory interface {
	Cookie(value string) *http.Cookie
	CookieName() string
}
