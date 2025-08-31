package app

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/auto-dns/pihole-cluster-admin/internal/config"
	"github.com/auto-dns/pihole-cluster-admin/internal/database"
	"github.com/auto-dns/pihole-cluster-admin/internal/domain"
	"github.com/auto-dns/pihole-cluster-admin/internal/http/api/unversioned"
	v1 "github.com/auto-dns/pihole-cluster-admin/internal/http/api/v1"
	"github.com/auto-dns/pihole-cluster-admin/internal/http/cookies"
	"github.com/auto-dns/pihole-cluster-admin/internal/http/middleware"
	"github.com/auto-dns/pihole-cluster-admin/internal/http/server"
	"github.com/auto-dns/pihole-cluster-admin/internal/pihole"
	"github.com/auto-dns/pihole-cluster-admin/internal/realtime"
	authsvc "github.com/auto-dns/pihole-cluster-admin/internal/service/auth"
	clusterblockingsvc "github.com/auto-dns/pihole-cluster-admin/internal/service/clusterblocking"
	domainrulesvc "github.com/auto-dns/pihole-cluster-admin/internal/service/domainrule"
	eventssvc "github.com/auto-dns/pihole-cluster-admin/internal/service/events"
	healthsvc "github.com/auto-dns/pihole-cluster-admin/internal/service/health"
	piholesvc "github.com/auto-dns/pihole-cluster-admin/internal/service/pihole"
	querylogsvc "github.com/auto-dns/pihole-cluster-admin/internal/service/querylog"
	setupsvc "github.com/auto-dns/pihole-cluster-admin/internal/service/setup"
	usersvc "github.com/auto-dns/pihole-cluster-admin/internal/service/user"
	"github.com/auto-dns/pihole-cluster-admin/internal/sessions"
	"github.com/auto-dns/pihole-cluster-admin/internal/store"
	"github.com/go-chi/chi"
	chimw "github.com/go-chi/chi/middleware"
	"github.com/rs/zerolog"
)

type App struct {
	Logger        zerolog.Logger
	Server        HttpServer
	SSEPublishers []SSEPublisher
	Sessions      SessionPurger
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
	txProvider := store.NewTransactor(db)

	clients, err := GetClients(piholeStore, logger)
	if err != nil {
		logger.Error().Err(err).Msg("error loading clients from database")
	}
	cursorManager := pihole.NewCursorManager[domain.QueryLogFilters](cfg.Server.Session.TTLHours)
	cluster := pihole.NewCluster(clients, cursorManager, logger)

	// Broker
	broker := realtime.NewBroker()

	// Handler
	sessionStorage := newSessionStorage(cfg.Server.Session, sessionStore, logger)
	sessionManager := sessions.NewSessionManager(sessionStorage, cfg.Server.Session, logger)

	// Cookie factory
	cookieFactory := cookies.NewSessionCookieFactory(cfg.Server.Session)

	// Router
	authService := authsvc.NewService(userStore, sessionManager, logger)
	clusterBlockingService := clusterblockingsvc.NewService(cluster, broker, logger)
	domainRuleService := domainrulesvc.NewService(cluster)
	eventsService := eventssvc.NewService(broker, logger)
	healthService := healthsvc.NewService(broker, cluster, cfg.HealthService, logger)
	piholeService := piholesvc.NewService(cluster, piholeStore, logger)
	queryLogService := querylogsvc.NewService(cluster, logger)
	setupService := setupsvc.NewService(initializationStatusStore, userStore, sessionManager, txProvider, logger)
	userService := usersvc.NewService(userStore, logger)
	// Middleware
	requireAuthMiddleware := middleware.RequireAuth(middleware.AuthDeps{
		Sessions: sessionManager,
		Cfg:      cfg.Server.Session,
		Logger:   logger,
	})

	// Root router
	rootRouter := chi.NewRouter()
	rootRouter.Use(middleware.RequestLogger(logger))
	rootRouter.Use(
		chimw.RequestID,
		chimw.RealIP,
		chimw.Recoverer,
		chimw.CleanPath,
		chimw.RedirectSlashes,
	)
	// API router
	apiRouter := chi.NewRouter()
	rootRouter.Mount("/api", apiRouter)

	apiRouter.Use(
		chimw.AllowContentType("application/json"),
		chimw.Compress(-1),
		chimw.Timeout(30*time.Second),
	)

	// Register unversioned routes
	unversioned.RegisterAPIUnversioned(apiRouter, unversioned.Deps{
		EventsService: eventsService,
		AuthMW:        requireAuthMiddleware,
		Cfg:           cfg.Server.ServerSideEvents,
		Db:            db,
		Logger:        logger,
	})

	// Register v1 routes
	apiV1 := chi.NewRouter()
	apiRouter.Mount("/v1", apiV1)
	v1.RegisterAPIV1(apiV1, v1.Deps{
		AuthService:            authService,
		ClusterBlockingService: clusterBlockingService,
		DomainRuleService:      domainRuleService,
		HealthService:          healthService,
		PiholeService:          piholeService,
		QueryLogService:        queryLogService,
		SetupService:           setupService,
		UserService:            userService,
		HttpCookieFactory:      cookieFactory,
		AuthMW:                 requireAuthMiddleware,
		Logger:                 logger,
	})

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
		SSEPublishers: []SSEPublisher{clusterBlockingService, healthService},
		Sessions:      purgeAdapter{sessionManager},
	}, nil
}

// Run starts the application by running the sync engine.
func (a *App) Run(ctx context.Context) error {
	defer a.Logger.Info().Msg("Application stopped")
	a.Logger.Info().Msg("Application starting")

	// Start SSE publishers
	for _, p := range a.SSEPublishers {
		go p.StartPublisher(ctx)
	}

	// Start session purge loop
	go a.Sessions.Start(ctx)

	// Start http server
	return a.Server.StartAndServe(ctx)
}
