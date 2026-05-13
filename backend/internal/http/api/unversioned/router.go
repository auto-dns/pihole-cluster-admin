package unversioned

import (
	"net/http"

	"github.com/auto-dns/pihole-cluster-admin/internal/config"
	"github.com/go-chi/chi"
	"github.com/rs/zerolog"
)

type Deps struct {
	EventsService eventsService
	// Auth middleware for private routes
	AuthMW func(http.Handler) http.Handler
	// Other
	Cfg    config.ServerSideEventsConfig
	Db     pinger
	Logger zerolog.Logger
}

func RegisterAPIUnversioned(r chi.Router, rootRouter chi.Router, d Deps) {
	registerFrontend(rootRouter, d)

	// Public
	r.Group(func(r chi.Router) {
		registerHealthcheck(r, d)
	})

	// Private
	r.Group(func(r chi.Router) {
		r.Use(d.AuthMW)
		registerEvents(r, d)
	})
}
