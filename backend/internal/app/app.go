package app

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/auto-dns/pihole-cluster-admin/internal/config"
	"github.com/auto-dns/pihole-cluster-admin/internal/database"
	auth_h "github.com/auto-dns/pihole-cluster-admin/internal/handler/auth"
	domainrule_h "github.com/auto-dns/pihole-cluster-admin/internal/handler/domainrule"
	events_h "github.com/auto-dns/pihole-cluster-admin/internal/handler/events"
	frontend_h "github.com/auto-dns/pihole-cluster-admin/internal/handler/frontend"
	health_h "github.com/auto-dns/pihole-cluster-admin/internal/handler/health"
	healthcheck_h "github.com/auto-dns/pihole-cluster-admin/internal/handler/healthcheck"
	pihole_h "github.com/auto-dns/pihole-cluster-admin/internal/handler/pihole"
	querylog_h "github.com/auto-dns/pihole-cluster-admin/internal/handler/querylog"
	setup_h "github.com/auto-dns/pihole-cluster-admin/internal/handler/setup"
	user_h "github.com/auto-dns/pihole-cluster-admin/internal/handler/user"
	apimw "github.com/auto-dns/pihole-cluster-admin/internal/middleware"
	"github.com/auto-dns/pihole-cluster-admin/internal/pihole"
	"github.com/auto-dns/pihole-cluster-admin/internal/realtime"
	"github.com/auto-dns/pihole-cluster-admin/internal/server"
	auth_s "github.com/auto-dns/pihole-cluster-admin/internal/service/auth"
	domainrule_s "github.com/auto-dns/pihole-cluster-admin/internal/service/domainrule"
	events_s "github.com/auto-dns/pihole-cluster-admin/internal/service/events"
	health_s "github.com/auto-dns/pihole-cluster-admin/internal/service/health"
	pihole_s "github.com/auto-dns/pihole-cluster-admin/internal/service/pihole"
	querylog_s "github.com/auto-dns/pihole-cluster-admin/internal/service/querylog"
	setup_s "github.com/auto-dns/pihole-cluster-admin/internal/service/setup"
	user_s "github.com/auto-dns/pihole-cluster-admin/internal/service/user"
	"github.com/auto-dns/pihole-cluster-admin/internal/sessions"
	"github.com/auto-dns/pihole-cluster-admin/internal/store"
	"github.com/go-chi/chi"
	chimw "github.com/go-chi/chi/middleware"
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

	// Handler
	sessionStorage := newSessionStorage(cfg.Server.Session, sessionStore, logger)
	sessionManager := sessions.NewSessionManager(sessionStorage, cfg.Server.Session, logger)

	// Router
	authService := auth_s.NewService(userStore, sessionManager, logger)
	authHandler := auth_h.NewHandler(authService, sessionManager, logger)
	domainService := domainrule_s.NewService(cluster)
	domainRuleHandler := domainrule_h.NewHandler(domainService, logger)
	eventsService := events_s.NewService(broker, logger)
	eventsHandler := events_h.NewHandler(cfg.Server.ServerSideEvents, eventsService, logger)
	frontendHandler := frontend_h.NewHandler(logger)
	healthcheckHandler := healthcheck_h.NewHandler(logger)
	healthService := health_s.NewService(broker, cluster, cfg.HealthService, logger)
	healthHandler := health_h.NewHandler(healthService, logger)
	piholeService := pihole_s.NewService(cluster, piholeStore, logger)
	piholeHandler := pihole_h.NewHandler(piholeService, logger)
	queryLogService := querylog_s.NewService(cluster, logger)
	queryLogHandler := querylog_h.NewHandler(queryLogService, logger)
	setupService := setup_s.NewService(initializationStatusStore, userStore, sessionManager, logger)
	setupHandler := setup_h.NewHandler(setupService, sessionManager, logger)
	userService := user_s.NewService(userStore, logger)
	userHandler := user_h.NewHandler(userService, logger)

	// Root router
	rootRouter := chi.NewRouter()
	rootRouter.Use(apimw.RequestLogger(logger))
	rootRouter.Use(chimw.RequestID, chimw.RealIP, chimw.Recoverer, chimw.CleanPath, chimw.RedirectSlashes)
	// API router
	apiRouter := chi.NewRouter()
	rootRouter.Mount("/api", apiRouter)
	apiRouter.Use(chimw.AllowContentType("application/json"), chimw.Compress(-1), chimw.Timeout(30*time.Second))

	// Public
	apiRouter.Group(func(r chi.Router) {
		authHandler.RegisterPublic(r)
		r.Route("/healthcheck", func(r chi.Router) { healthcheckHandler.Register(r) })
	})

	// Private
	apiRouter.Group(func(r chi.Router) {
		// Middleware
		r.Use(sessionManager.AuthMiddleware)
		// Routes
		authHandler.RegisterPrivate(r)
		r.Route("/cluster/health", func(r chi.Router) { healthHandler.Register(r) })
		r.Route("/domain", func(r chi.Router) { domainRuleHandler.Register(r) })
		r.Route("/events", func(r chi.Router) { eventsHandler.Register(r) })
		r.Route("/pihole", func(r chi.Router) { piholeHandler.Register(r) })
		r.Route("/querylog", func(r chi.Router) { queryLogHandler.Register(r) })
		r.Route("/user", func(r chi.Router) { userHandler.Register(r) })
	})

	// Mixed
	apiRouter.Route("/setup", func(r chi.Router) {
		// Public
		setupHandler.RegisterPublic(r)

		// Private
		r.Group(func(r chi.Router) {
			r.Use(sessionManager.AuthMiddleware)
			setupHandler.RegisterPrivate(r)
		})
	})

	// Front end
	frontendHandler.Register(rootRouter)

	// Server
	httpServer := &http.Server{
		Addr:              fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:           rootRouter,
		ReadHeaderTimeout: time.Duration(cfg.Server.ReadHeaderTimeoutSeconds) * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
	srv := server.New(httpServer, rootRouter, &cfg.Server, logger)
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
