package mcphandler

import (
	"context"
	"net/http"

	"github.com/auto-dns/pihole-cluster-admin/internal/domain"
	"github.com/mark3labs/mcp-go/server"
	"github.com/rs/zerolog"
)

type clusterBlockingService interface {
	GetState(ctx context.Context) (*domain.ClusterBlockingState, error)
	SetState(ctx context.Context, blocking bool, timer *int) (*domain.ClusterBlockingState, error)
	SetStateForNode(ctx context.Context, nodeID int64, blocking bool, timer *int) (*domain.ClusterBlockingState, error)
}

type queryLogService interface {
	Fetch(ctx context.Context, params domain.QueryLogQuery) (*domain.ClusterQueryLogResponse, error)
}

type domainRuleService interface {
	List(ctx context.Context, q domain.ListDomainRulesQuery) map[int64]*domain.NodeResult[*domain.DomainRuleSet]
	Add(ctx context.Context, cmd domain.AddDomainRulesCommand) map[int64]*domain.NodeResult[*domain.AddDomainRulesResult]
	Remove(ctx context.Context, cmd domain.RemoveDomainRuleCommand) map[int64]*domain.NodeResult[struct{}]
}

type healthService interface {
	GetClusterHealth(ctx context.Context) domain.ClusterHealth
}

type Deps struct {
	ClusterBlockingService clusterBlockingService
	DomainRuleService      domainRuleService
	QueryLogService        queryLogService
	HealthService          healthService
	Logger                 zerolog.Logger
}

func NewHandler(d Deps) http.Handler {
	s := server.NewMCPServer("pihole-cluster-admin", "1.0.0")
	registerTools(s, d)
	return server.NewStreamableHTTPServer(s, server.WithStateLess(true))
}
