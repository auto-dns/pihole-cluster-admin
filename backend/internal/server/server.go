package server

import (
	"context"
	"net/http"
	"time"

	"github.com/auto-dns/pihole-cluster-admin/internal/api"
	"github.com/auto-dns/pihole-cluster-admin/internal/config"
	"github.com/auto-dns/pihole-cluster-admin/internal/frontend"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/rs/zerolog"
)

type Server struct {
	cfg      *config.ServerConfig
	logger   zerolog.Logger
	router   chi.Router
	http     *http.Server
	handler  api.HandlerInterface
	sessions api.SessionInterface
}

func New(http *http.Server, router chi.Router, handler api.HandlerInterface, sessions api.SessionInterface, cfg *config.ServerConfig, logger zerolog.Logger) *Server {
	s := &Server{
		cfg:      cfg,
		logger:   logger,
		router:   router,
		http:     http,
		handler:  handler,
		sessions: sessions,
	}
	s.registerRoutes()
	return s
}

func (s *Server) registerRoutes() {
	// API routes
	// -- Global middlewares
	s.router.Use(RequestLogger(s.logger))
	s.router.Use(middleware.Recoverer)

	// -- Public routes
	s.router.Get("/api/healthcheck", s.handler.Healthcheck)
	s.router.Post("/api/login", s.handler.Login)
	s.router.Post("/api/logout", s.handler.Logout)

	// -- Protected routes
	protected := chi.NewRouter()
	protected.Use(s.handler.AuthMiddleware)

	protected.Get("/api/logs/queries", s.handler.FetchQueryLogs)
	protected.Get("/api/domains", s.handler.GetDomainRules)
	protected.Get("/api/domains/*", s.handler.GetDomainRules)
	protected.Post("/api/domains/{type}/{kind}", s.handler.AddDomainRule)
	protected.Delete("/api/domains/{type}/{kind}/{domain}", s.handler.RemoveDomainRule)

	s.router.Mount("/", protected)

	// Frontend
	if s.cfg.Proxy.Enable {
		s.logger.Debug().Msg("Proxying frontend to Vite")
		s.router.Handle("/", frontend.ProxyToVite())
	} else {
		s.logger.Debug().Msg("Serving embedded frontend")
		s.router.Handle("/", frontend.ServeStatic())
	}
}

func (s *Server) Start(ctx context.Context) error {
	s.logger.Info().Str("addr", s.http.Addr).Msg("Starting HTTP server")
	go func() {
		var err error
		if s.cfg.TLSEnabled {
			s.logger.Info().
				Str("cert", s.cfg.TLSCertFile).
				Str("key", s.cfg.TLSKeyFile).
				Msg("TLS enabled")
			err = s.http.ListenAndServeTLS(s.cfg.TLSCertFile, s.cfg.TLSKeyFile)
		} else {
			err = s.http.ListenAndServe()
		}
		if err != nil && err != http.ErrServerClosed {
			s.logger.Error().Err(err).Msg("HTTP server failed")
		}
	}()

	go func() {
		ticker := time.NewTicker(10 * time.Minute)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				s.sessions.PurgeExpired()
			case <-ctx.Done():
				s.logger.Info().Msg("Stopping session purge loop")
				return
			}
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
			logger.Info().
				Str("method", r.Method).
				Str("path", r.URL.Path).
				Int("status", ww.Status()).
				Dur("duration", time.Since(start)).
				Msg("request completed")
		})
	}
}
