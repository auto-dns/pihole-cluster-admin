package app

import (
	"context"
	"fmt"
	"net/http"

	"github.com/auto-dns/pihole-cluster-admin/internal/api"
	"github.com/auto-dns/pihole-cluster-admin/internal/config"
	"github.com/auto-dns/pihole-cluster-admin/internal/pihole"
	"github.com/auto-dns/pihole-cluster-admin/internal/server"
	"github.com/rs/zerolog"
)

type App struct {
	Logger zerolog.Logger
	Server httpServer
}

func NewPiholeClients(cfgs []config.PiholeConfig) []pihole.ClientInterface {
	clients := make([]pihole.ClientInterface, 0, len(cfgs))
	for _, c := range cfgs {
		client := pihole.NewClient(&c)
		clients = append(clients, client)
	}
	return clients
}

// NewClusterClient wires multiple Pi-hole node clients into one cluster client.
func NewClusterClient(clients []pihole.ClientInterface) *pihole.Cluster {
	return pihole.NewCluster(clients...)
}

func NewHandler(cluster *pihole.Cluster, logger zerolog.Logger) api.HandlerInterface {
	return api.NewHandler(cluster, logger)
}

func NewServer(cfg *config.ServerConfig, handler api.HandlerInterface, logger zerolog.Logger) httpServer {
	mux := http.NewServeMux()

	http := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Port),
		Handler: mux,
	}

	return server.New(http, mux, handler, cfg, logger)
}

// New creates a new App by wiring up all dependencies.
func New(cfg *config.Config, logger zerolog.Logger) (*App, error) {
	nodeClients := NewPiholeClients(cfg.Piholes)

	cluster := NewClusterClient(nodeClients)

	handler := NewHandler(cluster, logger)

	srv := NewServer(&cfg.Server, handler, logger)

	return &App{
		Logger: logger,
		Server: srv,
	}, nil
}

// Run starts the application by running the sync engine.
func (a *App) Run(ctx context.Context) error {
	a.Logger.Info().Msg("Application starting")
	return a.Server.Start(ctx)
}
