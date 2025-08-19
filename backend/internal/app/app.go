package app

import (
	"context"
	"strings"

	"github.com/auto-dns/pihole-cluster-admin/internal/config"
	"github.com/auto-dns/pihole-cluster-admin/internal/database"
	domainsH "github.com/auto-dns/pihole-cluster-admin/internal/handler/domains"
	"github.com/auto-dns/pihole-cluster-admin/internal/health"
	"github.com/auto-dns/pihole-cluster-admin/internal/pihole"
	"github.com/auto-dns/pihole-cluster-admin/internal/realtime"
	domainsS "github.com/auto-dns/pihole-cluster-admin/internal/service/domains"
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

	// Router
	domainsService := domainsS.NewService(cluster)
	domainsHandler := domainsH.NewHandler(domainsService, logger)

	rootRouter := chi.NewRouter()
	apiRouter := chi.NewRouter()
	apiRouter.Mount("/domains", domainsHandler.Routes())
	rootRouter.Mount("/api", apiRouter)

	// Server
	srv := NewServer(&cfg.Server, rootRouter, logger)
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
