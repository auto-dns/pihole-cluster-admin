package v1

import (
	"net/http"

	"github.com/auto-dns/pihole-cluster-admin/internal/http/middleware"
	"github.com/go-chi/chi"
	"github.com/rs/zerolog"
)

type Deps struct {
	AdlistService          adlistService
	AuditLogService        auditLogService
	AuthService            authService
	ClusterBlockingService clusterBlockingService
	DomainRuleService      domainRuleService
	GroupService           groupService
	HealthService          healthService
	PiholeClientService    piholeClientService
	PiholeService          piholeService
	QueryLogService        queryLogService
	StatsService           statsService
	SetupService           setupService
	SyncService            syncService
	UserService            userService

	// Other dependencies
	HttpCookieFactory httpCookieFactory

	// Auth middleware for private routes
	AuthMW func(http.Handler) http.Handler

	// Other
	Logger zerolog.Logger
}

func RegisterAPIV1(r chi.Router, d Deps) {
	// Public
	r.Group(func(r chi.Router) {
		registerAuthPublic(r, d)
		registerSetupPublic(r, d)
	})

	// Private
	r.Group(func(r chi.Router) {
		r.Use(d.AuthMW)
		r.Use(middleware.CSRF)
		registerAdlists(r, d)
		registerAuditLog(r, d)
		registerAuthPrivate(r, d)
		registerClusterBlocking(r, d)
		registerPiholeClients(r, d)
		registerDiagnose(r, d)
		registerGroups(r, d)
		registerHealth(r, d)
		registerDomainRules(r, d)
		registerPihole(r, d)
		registerQueryLog(r, d)
		registerStats(r, d)
		registerSetupPrivate(r, d)
		registerSync(r, d)
		registerUser(r, d)
	})
}
