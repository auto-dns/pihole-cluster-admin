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
	Server server.ServerInterface
}

func GetClients(piholeStore store.PiholeStoreInterface, logger zerolog.Logger) (map[int64]pihole.ClientInterface, error) {
	// Load piholes from database
	nodes, err := piholeStore.GetAllPiholeNodesWithPasswords()
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to load pihole nodes from database")
		return nil, err
	}

	clients := make(map[int64]pihole.ClientInterface, len(nodes))
	for _, node := range nodes {
		node := node
		cfg := &pihole.ClientConfig{
			Id:       node.Id,
			Scheme:   node.Scheme,
			Host:     node.Host,
			Port:     node.Port,
			Password: *node.Password,
			Name:     node.Name,
		}
		nodeLogger := logger.With().Int64("db_id", node.Id).Str("host", node.Host).Int("port", node.Port).Logger()
		clients[node.Id] = pihole.NewClient(cfg, nodeLogger)
	}
	logger.Info().Int("node_count", len(nodes)).Msg("loaded pihole nodes")

	return clients, nil
}

func NewServer(cfg *config.ServerConfig, handler api.HandlerInterface, sessions api.SessionManagerInterface, logger zerolog.Logger) server.ServerInterface {
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
	db, err := database.NewDatabase(cfg.Database)
	if err != nil {
		logger.Error().Err(err).Msg("error initializing database")
		return nil, err
	}
	initializationStatusStore := store.NewInitializationStore(db, logger)
	piholeStore := store.NewPiholeStore(db, cfg.EncryptionKey, logger)
	userStore := store.NewUserStore(db, logger)

	clients, err := GetClients(piholeStore, logger)
	if err != nil {
		logger.Error().Err(err).Msg("error loading clients from database")
	}
	cursorManager := pihole.NewCursorManager[pihole.FetchQueryLogFilters](cfg.Server.Session.TTLHours)
	cluster := pihole.NewCluster(clients, cursorManager, logger)

	// Handler
	sessions := api.NewSessionManager(cfg.Server.Session, logger)
	handler := api.NewHandler(cluster, sessions, initializationStatusStore, piholeStore, userStore, cfg.Server.Session, logger)

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
