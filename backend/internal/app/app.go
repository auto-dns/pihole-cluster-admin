package app

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/auto-dns/pihole-cluster-admin/internal/api"
	"github.com/auto-dns/pihole-cluster-admin/internal/config"
	"github.com/auto-dns/pihole-cluster-admin/internal/database"
	"github.com/auto-dns/pihole-cluster-admin/internal/health"
	"github.com/auto-dns/pihole-cluster-admin/internal/pihole"
	"github.com/auto-dns/pihole-cluster-admin/internal/realtime"
	"github.com/auto-dns/pihole-cluster-admin/internal/server"
	"github.com/auto-dns/pihole-cluster-admin/internal/sessions"
	"github.com/auto-dns/pihole-cluster-admin/internal/store"
	"github.com/go-chi/chi"
	"github.com/rs/zerolog"
)

type App struct {
	Logger        zerolog.Logger
	Server        HttpServer
	Sessions      SessionPurger
	HealthService HealthService
}

func GetClients(piholeGetter PiholeGetter, logger zerolog.Logger) (map[int64]*pihole.Client, error) {
	// Load piholes from database
	nodes, err := piholeGetter.GetAllPiholeNodesWithPasswords()
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to load pihole nodes from database")
		return nil, err
	}

	clients := make(map[int64]*pihole.Client, len(nodes))
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

func NewServer(cfg *config.ServerConfig, handler *api.Handler, logger zerolog.Logger) *server.Server {
	router := chi.NewRouter()

	http := &http.Server{
		Addr:              fmt.Sprintf(":%d", cfg.Port),
		Handler:           router,
		ReadHeaderTimeout: time.Duration(cfg.ReadHeaderTimeoutSeconds) * time.Second,
	}

	logger.Info().Int("port", cfg.Port).Bool("tls", cfg.TLSEnabled).Msg("server created")

	return server.New(http, router, handler, cfg, logger)
}

func newSessionStorage(cfg config.SessionConfig, sessionSqliteStore SessionSqliteStore, logger zerolog.Logger) SessionStorage {
	switch strings.ToLower(cfg.Backend) {
	case "memory":
		logger.Info().Msg("using in-memory session store")
		return sessions.NewMemorySessionStore()
	case "sqlite", "":
		logger.Info().Msg("using sqlite session store")
		return sessions.NewSqliteSessionStore(sessionSqliteStore)
	default:
		logger.Warn().Str("backend", cfg.Backend).Msg("unknown session backend; falling back to sqlite")
		return sessions.NewSqliteSessionStore(sessionSqliteStore)
	}
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
	sessionStore := store.NewSessionStore(db, logger)
	userStore := store.NewUserStore(db, logger)

	clients, err := GetClients(piholeStore, logger)
	if err != nil {
		logger.Error().Err(err).Msg("error loading clients from database")
	}
	cursorManager := pihole.NewCursorManager[pihole.FetchQueryLogFilters](cfg.Server.Session.TTLHours)
	cluster := pihole.NewCluster(clients, cursorManager, logger)

	// Broker
	broker := realtime.NewBroker()

	// Health Service
	healthService := health.NewService(cluster, broker, cfg.HealthService, logger)

	// Handler
	sessionStorage := newSessionStorage(cfg.Server.Session, sessionStore, logger)
	sessionsManager := sessions.NewSessionManager(sessionStorage, cfg.Server.Session, logger)
	handler := api.NewHandler(cluster, sessionsManager, initializationStatusStore, piholeStore, userStore, healthService, broker, cfg.Server, logger)

	// Server
	srv := NewServer(&cfg.Server, handler, logger)
	logger.Info().Msg("application dependencies wired")

	return &App{
		Logger:        logger,
		Server:        srv,
		Sessions:      purgeAdapter{sessionsManager},
		HealthService: healthService,
	}, nil
}

// Run starts the application by running the sync engine.
func (a *App) Run(ctx context.Context) error {
	defer a.Logger.Info().Msg("Application stopped")
	a.Logger.Info().Msg("Application starting")

	// Start health service
	go a.HealthService.Start(ctx)

	// Start session purge loop
	go a.Sessions.Start(ctx)

	// Start http server
	return a.Server.StartAndServe(ctx)
}
