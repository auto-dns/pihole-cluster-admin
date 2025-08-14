package server

import (
	"context"
	"io/fs"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/auto-dns/pihole-cluster-admin/internal/config"
	"github.com/auto-dns/pihole-cluster-admin/internal/frontend"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/rs/zerolog"
)

type Server struct {
	cfg     *config.ServerConfig
	logger  zerolog.Logger
	router  chi.Router
	http    *http.Server
	handler handler
}

func New(http *http.Server, router chi.Router, handler handler, cfg *config.ServerConfig, logger zerolog.Logger) *Server {
	s := &Server{
		cfg:     cfg,
		logger:  logger,
		router:  router,
		http:    http,
		handler: handler,
	}
	s.registerRoutes()
	return s
}

func (s *Server) registerRoutes() {
	// API routes
	// -- Global middlewares
	s.router.Use(RequestLogger(s.logger))
	s.router.Use(middleware.Recoverer)

	api := chi.NewRouter()

	// -- Public routes
	api.Get("/healthcheck", s.handler.Healthcheck)
	api.Post("/login", s.handler.Login)
	api.Post("/logout", s.handler.Logout)
	api.Get("/setup/initialized", s.handler.GetIsInitialized)
	api.Post("/setup/user", s.handler.CreateUser)
	// -- Protected routes
	protected := chi.NewRouter()
	protected.Use(s.handler.AuthMiddleware)

	// ---- Setup Status
	protected.Get("/setup/status", s.handler.GetInitializationStatus)
	protected.Patch("/setup/status/pihole", s.handler.UpdatePiholeInitializationStatus)
	// ---- Event Streaming
	protected.Get("/events", s.handler.HandleEvents)
	// ---- Health Status
	protected.Get("/cluster/health/summary", s.handler.GetHealthSummary)
	protected.Get("/cluster/health/node", s.handler.GetNodeHealth)
	// ---- User
	protected.Get("/session/user", s.handler.GetSessionUser)
	// ---- Piholes
	protected.Get("/piholes", s.handler.GetAllPiholeNodes)
	protected.Post("/piholes", s.handler.AddPiholeNode)
	protected.Patch("/piholes/{id}", s.handler.UpdatePiholeNode)
	protected.Delete("/piholes/{id}", s.handler.RemovePiholeNode)
	protected.Post("/piholes/{id}/test", s.handler.TestExistingPiholeConnection)
	protected.Post("/piholes/test", s.handler.TestPiholeInstanceConnection)
	// ---- Query logs
	protected.Get("/logs/queries", s.handler.FetchQueryLogs)
	// ---- Domain management
	protected.Get("/domains", s.handler.GetDomainRules)
	protected.Get("/domains/*", s.handler.GetDomainRules)
	protected.Post("/domains/{type}/{kind}", s.handler.AddDomainRule)
	protected.Delete("/domains/{type}/{kind}/{domain}", s.handler.RemoveDomainRule)

	api.Mount("/", protected)
	s.router.Mount("/api", api)

	s.registerFrontEnd()
}

func (s *Server) registerFrontEnd() {
	sub, err := fs.Sub(frontend.Files, "internal/frontend/dist")
	if err != nil {
		s.logger.Warn().Msg("No embedded frontend found; assuming development mode (Vite)")
		return
	}

	fileServer := http.FileServer(http.FS(sub))

	// Handle all non-API routes
	s.router.HandleFunc("/*", func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/")
		if path == "" {
			path = "index.html"
		}
		_, err := sub.Open(path)
		if err != nil {
			// SPA fallback → serve index.html
			index, err := sub.Open("index.html")
			if err != nil {
				http.Error(w, "index.html missing in embedded frontend", http.StatusInternalServerError)
				return
			}
			defer index.Close()
			http.ServeFile(w, r, filepath.Join("internal/frontend/dist", "index.html"))
			return
		}
		fileServer.ServeHTTP(w, r)
	})

	s.logger.Info().Msg("Serving embedded frontend build")
}

func (s *Server) StartAndServe(ctx context.Context) error {
	s.logger.Info().Str("addr", s.http.Addr).Msg("Starting HTTP server")
	go func() {
		var err error
		if s.cfg.TLSEnabled {
			s.logger.Info().Str("cert", s.cfg.TLSCertFile).Str("key", s.cfg.TLSKeyFile).Msg("TLS enabled")
			err = s.http.ListenAndServeTLS(s.cfg.TLSCertFile, s.cfg.TLSKeyFile)
		} else {
			err = s.http.ListenAndServe()
		}
		if err != nil && err != http.ErrServerClosed {
			s.logger.Error().Err(err).Msg("HTTP server failed")
		}
	}()

	<-ctx.Done()

	s.logger.Info().Msg("Shutting down HTTP server")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := s.http.Shutdown(shutdownCtx); err != nil {
		s.logger.Error().Err(err).Msg("Graceful shutdown failed")
		return err
	}
	s.logger.Info().Msg("Server shut down cleanly")
	return nil
}

func RequestLogger(logger zerolog.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
			next.ServeHTTP(ww, r)
			logger.Info().Str("method", r.Method).Str("path", r.URL.Path).Str("remote", r.RemoteAddr).Str("agent", r.UserAgent()).Int("status", ww.Status()).Dur("duration", time.Since(start)).Msg("request completed")
		})
	}
}
