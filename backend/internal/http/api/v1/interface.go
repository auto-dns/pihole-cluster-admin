package v1

import (
	"context"
	"net/http"

	"github.com/auto-dns/pihole-cluster-admin/internal/domain"
	authsvc "github.com/auto-dns/pihole-cluster-admin/internal/service/auth"
	piholesvc "github.com/auto-dns/pihole-cluster-admin/internal/service/pihole"
	setupsvc "github.com/auto-dns/pihole-cluster-admin/internal/service/setup"
	usersvc "github.com/auto-dns/pihole-cluster-admin/internal/service/user"
)

type authService interface {
	Login(params authsvc.LoginCommand) (*domain.User, string, error)
	Logout(sessionId string) error
	GetUser(id int64) (*domain.User, error)
}

type clusterBlockingService interface {
	GetState(ctx context.Context) (*domain.ClusterBlockingState, error)
	SetState(ctx context.Context, blocking bool, timer *int) (*domain.ClusterBlockingState, error)
}

type domainRuleService interface {
	List(ctx context.Context, q domain.ListDomainRulesQuery) map[int64]*domain.NodeResult[*domain.DomainRuleSet]
	Add(ctx context.Context, cmd domain.AddDomainRulesCommand) map[int64]*domain.NodeResult[*domain.AddDomainRulesResult]
	Remove(ctx context.Context, cmd domain.RemoveDomainRuleCommand) map[int64]*domain.NodeResult[struct{}]
}

type healthService interface {
	GetSummary() domain.ClusterHealthSummary
	GetNodeHealth() map[int64]domain.ClusterNodeHealth
}

type piholeService interface {
	GetAll() ([]*domain.PiholeNode, error)
	Add(ctx context.Context, params piholesvc.AddNodeCommand) (*domain.PiholeNode, error)
	Update(ctx context.Context, id int64, params piholesvc.UpdateNodeCommand) (*domain.PiholeNode, error)
	Remove(ctx context.Context, id int64) (found bool, err error)
	TestExistingConnection(ctx context.Context, id int64, params piholesvc.TestExistingConnectionCommand) error
	TestInstanceConnection(ctx context.Context, params piholesvc.TestInstanceConnectionCommand) error
}

type queryLogService interface {
	Fetch(ctx context.Context, req domain.QueryLogQuery) (*domain.ClusterQueryLogResponse, error)
}

type setupService interface {
	IsInitialized() (bool, error)
	CreateUser(ctx context.Context, params setupsvc.CreateUserCommand) (*domain.User, string, error)
	GetInitializationStatus() (*domain.InitStatus, error)
	UpdatePiholeInitializationStatus(params setupsvc.UpdatePiholeInitializationStatusCommand) error
}

type userService interface {
	Patch(id int64, params usersvc.PatchUserCommand) (*domain.User, error)
	UpdatePassword(id int64, params usersvc.UpdatePasswordCommand) (*domain.User, error)
}

type httpCookieFactory interface {
	Make(value string) *http.Cookie
	Name() string
}
