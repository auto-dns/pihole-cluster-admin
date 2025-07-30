package app

import (
	"context"
	"fmt"
	"net/http"

	"github.com/auto-dns/pihole-cluster-admin/internal/api"
	"github.com/auto-dns/pihole-cluster-admin/internal/config"
	"github.com/auto-dns/pihole-cluster-admin/internal/database"
	"github.com/auto-dns/pihole-cluster-admin/internal/pihole"
	"github.com/auto-dns/pihole-cluster-admin/internal/server"
	"github.com/auto-dns/pihole-cluster-admin/internal/store"
	"github.com/go-chi/chi"
	"github.com/rs/zerolog"
)

type App struct {
	Logger zerolog.Logger
	Server httpServer
}

func NewDatabase(cfg config.DatabaseConfig) (*database.Database, error) {
	db, err := database.NewDatabase(cfg)
	return db, err
}

func NewPiholeStore(db *database.Database, encryptionKey string) *store.PiholeStore {
	return store.NewPiholeStore(db, encryptionKey)
}

func NewUserStore(db *database.Database) *store.UserStore {
	return store.NewUserStore(db)
}

func NewPiholeClients(cfgs []config.PiholeConfig, logger zerolog.Logger) []pihole.ClientInterface {
	logger.Info().Int("count", len(cfgs)).Msg("creating Pi-hole clients")
	clients := make([]pihole.ClientInterface, 0, len(cfgs))
	for _, c := range cfgs {
		l := logger.With().Str("id", c.ID).Str("host", c.Host).Int("port", c.Port).Logger()
		l.Debug().Str("id", c.ID).Str("host", c.Host).Int("port", c.Port).Msg("adding Pi-hole client")
		client := pihole.NewClient(&c, l)
		clients = append(clients, client)
	}
	return clients
}

// NewClusterClient wires multiple Pi-hole node clients into one cluster client.
func NewClusterClient(clients []pihole.ClientInterface, logger zerolog.Logger) *pihole.Cluster {
	logger.Info().Int("node_count", len(clients)).Msg("cluster client created")
	return pihole.NewCluster(logger, clients...)
}

func NewSessionManager(cfg config.SessionConfig, logger zerolog.Logger) api.SessionInterface {
	logger.Info().Bool("secure_cookie", cfg.Secure).Int("ttl_hours", cfg.TTLHours).Msg("session manager initialized")
	return api.NewSessionManager(cfg, logger)
}

func NewHandler(cluster *pihole.Cluster, sessions api.SessionInterface, logger zerolog.Logger) api.HandlerInterface {
	return api.NewHandler(cluster, sessions, logger)
}

func NewServer(cfg *config.ServerConfig, handler api.HandlerInterface, sessions api.SessionInterface, logger zerolog.Logger) httpServer {
	router := chi.NewRouter()

	http := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Port),
		Handler: router,
	}

	logger.Info().Int("port", cfg.Port).Bool("tls", cfg.TLSEnabled).Msg("server created")

	return server.New(http, router, handler, sessions, cfg, logger)
}

// New creates a new App by wiring up all dependencies.
func New(cfg *config.Config, logger zerolog.Logger) (*App, error) {
	nodeClients := NewPiholeClients(cfg.Piholes, logger)

	cluster := NewClusterClient(nodeClients, logger)

	sessions := api.NewSessionManager(cfg.Server.Session, logger)

	handler := NewHandler(cluster, sessions, logger)

	srv := NewServer(&cfg.Server, handler, sessions, logger)

	logger.Info().Msg("application dependencies wired")

	return &App{
		Logger: logger,
		Server: srv,
	}, nil
}

// Run starts the application by running the sync engine.
func (a *App) Run(ctx context.Context) error {
	defer a.Logger.Info().Msg("Application stopped")
	a.Logger.Info().Msg("Application starting")
	return a.Server.Start(ctx)
}
