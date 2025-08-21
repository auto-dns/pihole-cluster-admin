package app

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/auto-dns/pihole-cluster-admin/internal/config"
	"github.com/auto-dns/pihole-cluster-admin/internal/database"
	"github.com/auto-dns/pihole-cluster-admin/internal/handler/authhandler"
	"github.com/auto-dns/pihole-cluster-admin/internal/handler/domainrulehandler"
	"github.com/auto-dns/pihole-cluster-admin/internal/handler/eventshandler"
	"github.com/auto-dns/pihole-cluster-admin/internal/handler/frontendhandler"
	"github.com/auto-dns/pihole-cluster-admin/internal/handler/healthcheckhandler"
	"github.com/auto-dns/pihole-cluster-admin/internal/handler/healthhandler"
	"github.com/auto-dns/pihole-cluster-admin/internal/handler/piholehandler"
	"github.com/auto-dns/pihole-cluster-admin/internal/handler/queryloghandler"
	"github.com/auto-dns/pihole-cluster-admin/internal/handler/setuphandler"
	"github.com/auto-dns/pihole-cluster-admin/internal/handler/userhandler"
	apimw "github.com/auto-dns/pihole-cluster-admin/internal/middleware"
	"github.com/auto-dns/pihole-cluster-admin/internal/pihole"
	"github.com/auto-dns/pihole-cluster-admin/internal/realtime"
	"github.com/auto-dns/pihole-cluster-admin/internal/server"
	"github.com/auto-dns/pihole-cluster-admin/internal/service/authservice"
	"github.com/auto-dns/pihole-cluster-admin/internal/service/domainruleservice"
	"github.com/auto-dns/pihole-cluster-admin/internal/service/eventsservice"
	"github.com/auto-dns/pihole-cluster-admin/internal/service/healthservice"
	"github.com/auto-dns/pihole-cluster-admin/internal/service/piholeservice"
	"github.com/auto-dns/pihole-cluster-admin/internal/service/querylogservice"
	"github.com/auto-dns/pihole-cluster-admin/internal/service/setupservice"
	"github.com/auto-dns/pihole-cluster-admin/internal/service/userservice"
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
	authService := authservice.NewService(userStore, sessionManager, logger)
	authHandler := authhandler.NewHandler(authService, sessionManager, logger)
	domainService := domainruleservice.NewService(cluster)
	domainRuleHandler := domainrulehandler.NewHandler(domainService, logger)
	eventsService := eventsservice.NewService(broker, logger)
	eventsHandler := eventshandler.NewHandler(cfg.Server.ServerSideEvents, eventsService, logger)
	frontendHandler := frontendhandler.NewHandler(logger)
	healthcheckHandler := healthcheckhandler.NewHandler(logger)
	healthService := healthservice.NewService(broker, cluster, cfg.HealthService, logger)
	healthHandler := healthhandler.NewHandler(healthService, logger)
	piholeService := piholeservice.NewService(cluster, piholeStore, logger)
	piholeHandler := piholehandler.NewHandler(piholeService, logger)
	queryLogService := querylogservice.NewService(cluster, logger)
	queryLogHandler := queryloghandler.NewHandler(queryLogService, logger)
	setupService := setupservice.NewService(initializationStatusStore, userStore, sessionManager, logger)
	setupHandler := setuphandler.NewHandler(setupService, sessionManager, logger)
	userService := userservice.NewService(userStore, logger)
	userHandler := userhandler.NewHandler(userService, logger)

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
