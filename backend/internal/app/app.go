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

func NewPiholeStore(db *database.Database, encryptionKey string) store.PiholeStoreInterface {
	return store.NewPiholeStore(db, encryptionKey)
}

func NewUserStore(db *database.Database) store.UserStoreInterface {
	return store.NewUserStore(db)
}

func NewClient(cfg *pihole.ClientConfig, logger zerolog.Logger) pihole.ClientInterface {
	return pihole.NewClient(cfg, logger)
}

func NewCluster(clients []pihole.ClientInterface, logger zerolog.Logger) pihole.ClusterInterface {
	logger.Info().Int("node_count", len(clients)).Msg("cluster client created")
	return pihole.NewCluster(logger, clients...)
}

func NewSessionManager(userStore store.UserStoreInterface, cfg config.SessionConfig, logger zerolog.Logger) api.SessionInterface {
	logger.Info().Bool("secure_cookie", cfg.Secure).Int("ttl_hours", cfg.TTLHours).Msg("session manager initialized")
	return api.NewSessionManager(cfg, logger)
}

func NewHandler(cluster pihole.ClusterInterface, sessions api.SessionInterface, userStore store.UserStoreInterface, logger zerolog.Logger) api.HandlerInterface {
	return api.NewHandler(cluster, sessions, userStore, logger)
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
	// Initialize database and store
	db, err := NewDatabase(cfg.Database)
	if err != nil {
		logger.Error().Err(err).Msg("error initializing database")
		return nil, err
	}
	piholeStore := NewPiholeStore(db, cfg.EncryptionKey)
	userStore := NewUserStore(db)

	// Load piholes from database
	nodes, err := piholeStore.GetAllNodes()
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to load pihole nodes from database")
	}
	var clients []pihole.ClientInterface
	for _, node := range nodes {
		cfg := pihole.ClientConfig{
			ID:          fmt.Sprintf("node-%d", node.ID),
			Scheme:      node.Scheme,
			Host:        node.Host,
			Port:        node.Port,
			Password:    node.Password,
			Description: node.Description,
		}
		nodeLogger := logger.With().Int("db_id", node.ID).Str("host", node.Host).Int("port", node.Port).Logger()
		clients = append(clients, NewClient(&cfg, nodeLogger))
	}
	logger.Info().Int("node_count", len(nodes)).Msg("loaded pihole nodes")
	cluster := NewCluster(clients, logger)

	// Handler
	sessions := api.NewSessionManager(cfg.Server.Session, logger)
	handler := NewHandler(cluster, sessions, userStore, logger)

	// Server
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
