package app

import (
	"context"
	"strings"

	"github.com/auto-dns/pihole-cluster-admin/internal/config"
	"github.com/auto-dns/pihole-cluster-admin/internal/database"
	"github.com/auto-dns/pihole-cluster-admin/internal/handler/authhandler"
	"github.com/auto-dns/pihole-cluster-admin/internal/handler/domainrulehandler"
	"github.com/auto-dns/pihole-cluster-admin/internal/handler/eventshandler"
	"github.com/auto-dns/pihole-cluster-admin/internal/handler/piholehandler"
	"github.com/auto-dns/pihole-cluster-admin/internal/handler/setuphandler"
	"github.com/auto-dns/pihole-cluster-admin/internal/handler/userhandler"
	"github.com/auto-dns/pihole-cluster-admin/internal/health"
	"github.com/auto-dns/pihole-cluster-admin/internal/pihole"
	"github.com/auto-dns/pihole-cluster-admin/internal/realtime"
	"github.com/auto-dns/pihole-cluster-admin/internal/service/authservice"
	"github.com/auto-dns/pihole-cluster-admin/internal/service/domainruleservice"
	"github.com/auto-dns/pihole-cluster-admin/internal/service/eventsservice"
	"github.com/auto-dns/pihole-cluster-admin/internal/service/piholeservice"
	"github.com/auto-dns/pihole-cluster-admin/internal/service/setupservice"
	"github.com/auto-dns/pihole-cluster-admin/internal/service/userservice"
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
	sessionManager := sessions.NewSessionManager(sessionStorage, cfg.Server.Session, logger)

	// Router
	authService := authservice.NewService(userStore, sessionManager, logger)
	authHandler := authhandler.NewHandler(authService, sessionManager, logger)
	domainService := domainruleservice.NewService(cluster)
	domainRuleHandler := domainrulehandler.NewHandler(domainService, logger)
	eventsService := eventsservice.NewService(broker, logger)
	eventsHandler := eventshandler.NewHandler(cfg.Server.ServerSideEvents, eventsService, logger)
	piholeService := piholeservice.NewService(cluster, piholeStore, logger)
	piholeHandler := piholehandler.NewHandler(piholeService, logger)
	setupService := setupservice.NewService(initializationStatusStore, userStore, sessionManager, logger)
	setupHandler := setuphandler.NewHandler(setupService, sessionManager, logger)
	userService := userservice.NewService(userStore, logger)
	userHandler := userhandler.NewHandler(userService, logger)

	rootRouter := chi.NewRouter()
	// API router
	apiRouter := chi.NewRouter()
	rootRouter.Mount("/api", apiRouter)
	// Public
	publicRouter := chi.NewRouter()
	apiRouter.Mount("/", publicRouter)
	publicRouter.Mount("/", authHandler.PublicRoutes())
	publicRouter.Mount("/setup", setupHandler.PublicRoutes())
	// Private
	privateRouter := chi.NewRouter()
	apiRouter.Mount("/", privateRouter)
	privateRouter.Mount("/", privateRouter)
	privateRouter.Mount("/", authHandler.PrivateRoutes())
	privateRouter.Mount("/domain", domainRuleHandler.Routes())
	privateRouter.Mount("/events", eventsHandler.Routes())
	privateRouter.Mount("/pihole", piholeHandler.Routes())
	privateRouter.Mount("/setup", setupHandler.PrivateRoutes())
	privateRouter.Mount("/user", userHandler.Routes())

	// Server
	srv := NewServer(&cfg.Server, rootRouter, logger)
	logger.Info().Msg("application dependencies wired")

	return &App{
		Logger:        logger,
		Server:        srv,
		Sessions:      purgeAdapter{sessionManager},
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
