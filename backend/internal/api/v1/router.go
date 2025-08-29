package v1

import (
	"net/http"

	"github.com/auto-dns/pihole-cluster-admin/internal/config"
	"github.com/go-chi/chi"
	"github.com/rs/zerolog"
)

type Deps struct {
	AuthService            authService
	ClusterBlockingService clusterBlockingService
	DomainRuleService      domainRuleService
	EventsService          eventsService
	HealthService          healthService
	PiholeService          piholeService
	QueryLogService        queryLogService
	SetupService           setupService
	UserService            userService

	// Other dependencies
	HttpCookieFactory httpCookieFactory

	// Auth middleware for private routes
	AuthMW func(http.Handler) http.Handler

	// Other
	Cfg    config.ServerSideEventsConfig
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
		registerAuthPrivate(r, d)
		registerClusterBlocking(r, d)
		registerHealth(r, d)
		registerDomainRules(r, d)
		registerEvents(r, d)
		registerPihole(r, d)
		registerQueryLog(r, d)
		registerSetupPrivate(r, d)
		registerUser(r, d)
	})
}
