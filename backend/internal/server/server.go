package server

import (
	"context"
	"net/http"
	"time"

	"github.com/auto-dns/pihole-cluster-admin/internal/api"
	"github.com/auto-dns/pihole-cluster-admin/internal/config"
	"github.com/auto-dns/pihole-cluster-admin/internal/frontend"
	"github.com/go-chi/chi"
	"github.com/rs/zerolog"
)

type Server struct {
	cfg     *config.ServerConfig
	logger  zerolog.Logger
	router  chi.Router
	http    *http.Server
	handler api.HandlerInterface
}

func New(http *http.Server, router chi.Router, handler api.HandlerInterface, cfg *config.ServerConfig, logger zerolog.Logger) *Server {
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
	s.router.Get("/api/healthcheck", s.handler.Healthcheck)
	s.router.Get("/api/logs/queries", s.handler.FetchQueryLogs)
	s.router.Get("/api/domains", s.handler.HandleGetDomainRules)
	s.router.Get("/api/domains/*", s.handler.HandleGetDomainRules)
	s.router.Post("/api/domains/{type}/{kind}", s.handler.HandleAddDomainRule)
	s.router.Delete("/api/domains/{type}/{kind}/{domain}", s.handler.HandleRemoveDomainRule)

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
		if err := s.http.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.logger.Error().Err(err).Msg("HTTP server failed")
		}
	}()

	<-ctx.Done() // Wait for context cancellation

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
