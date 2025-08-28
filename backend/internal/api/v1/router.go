package v1

import (
	"net/http"

	"github.com/auto-dns/pihole-cluster-admin/internal/config"
	"github.com/go-chi/chi"
	"github.com/rs/zerolog"
)

type Deps struct {
	AuthService            AuthService
	UserService            UserService
	SetupService           SetupService
	PiholeService          PiholeService
	DomainRuleService      DomainRuleService
	QueryLogService        QueryLogService
	ClusterBlockingService ClusterBlockingService
	HealthService          HealthService
	EventsService          EventsService

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
		registerAuthPublic(r, d)  // e.g. /auth/login
		registerHealthcheck(r, d) // if you keep it versioned
		registerSetupPublic(r, d) // /setup/...
	})

	// Private
	r.Group(func(r chi.Router) {
		r.Use(d.AuthMW)
		registerAuthPrivate(r, d)     // /auth/session/user, /auth/logout
		registerClusterBlocking(r, d) // /cluster/blocking
		registerHealth(r, d)          // /cluster/health
		registerDomainRules(r, d)     // /domain/...
		registerEvents(r, d)          // /events/...
		registerPihole(r, d)          // /pihole/...
		registerQueryLog(r, d)        // /querylog/...
		registerUser(r, d)            // /user/...
	})

	// Mixed (if you need /setup split like before)
	registerSetupMixed(r, d)
}
