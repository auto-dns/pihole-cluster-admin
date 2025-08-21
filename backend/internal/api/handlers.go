package api

import (
	"net/http"

	"github.com/auto-dns/pihole-cluster-admin/internal/config"
	"github.com/rs/zerolog"
)

type Handler struct {
	cluster         piholeCluster
	sessions        sessionDeps
	initStatusStore initStatusStore
	piholeStore     piholeStore
	userStore       userStore
	healthService   healthService
	eventSubscriber eventSubscriber
	logger          zerolog.Logger
	cfg             config.ServerConfig
}

func NewHandler(cluster piholeCluster, sessions sessionDeps, initStatusStore initStatusStore, piholeStore piholeStore, userStore userStore, healthService healthService, eventSubscriber eventSubscriber, cfg config.ServerConfig, logger zerolog.Logger) *Handler {
	return &Handler{
		cluster:         cluster,
		sessions:        sessions,
		initStatusStore: initStatusStore,
		piholeStore:     piholeStore,
		userStore:       userStore,
		healthService:   healthService,
		eventSubscriber: eventSubscriber,
		logger:          logger,
		cfg:             cfg,
	}
}

// Convenience function

// Handlers

func (h *Handler) AuthMiddleware(next http.Handler) http.Handler {
	return h.sessions.AuthMiddleware(next)
}

// Unauthenticated routes

func (h *Handler) Healthcheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status": "OK"}`))
}

// Authenticated routes

// Application business logic routes
