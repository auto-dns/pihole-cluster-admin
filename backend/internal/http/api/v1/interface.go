package v1

import (
	"context"
	"net/http"

	"github.com/auto-dns/pihole-cluster-admin/internal/domain"
	auth_s "github.com/auto-dns/pihole-cluster-admin/internal/service/auth"
	pihole_s "github.com/auto-dns/pihole-cluster-admin/internal/service/pihole"
	setup_s "github.com/auto-dns/pihole-cluster-admin/internal/service/setup"
	user_s "github.com/auto-dns/pihole-cluster-admin/internal/service/user"
)

type authService interface {
	Login(params auth_s.LoginCommand) (*domain.User, string, error)
	Logout(sessionId string) error
	GetUser(id int64) (*domain.User, error)
}

type clusterBlockingService interface {
	GetState(ctx context.Context) (*domain.ClusterBlockingState, error)
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
	Add(ctx context.Context, params pihole_s.AddNodeCommand) (*domain.PiholeNode, error)
	Update(ctx context.Context, id int64, params pihole_s.UpdateNodeCommand) (*domain.PiholeNode, error)
	Remove(ctx context.Context, id int64) (found bool, err error)
	TestExistingConnection(ctx context.Context, id int64, params pihole_s.TestExistingConnectionCommand) error
	TestInstanceConnection(ctx context.Context, params pihole_s.TestInstanceConnectionCommand) error
}

type queryLogService interface {
	Fetch(ctx context.Context, req domain.QueryLogQuery) (*domain.ClusterQueryLogResponse, error)
}

type setupService interface {
	IsInitialized() (bool, error)
	CreateUser(ctx context.Context, params setup_s.CreateUserCommand) (*domain.User, string, error)
	GetInitializationStatus() (*domain.InitStatus, error)
	UpdatePiholeInitializationStatus(params setup_s.UpdatePiholeInitializationStatusCommand) error
}

type userService interface {
	Patch(id int64, params user_s.PatchUserCommand) (*domain.User, error)
	UpdatePassword(id int64, params user_s.UpdatePasswordCommand) (*domain.User, error)
}

type httpCookieFactory interface {
	Make(value string) *http.Cookie
	Name() string
}
