package server

import (
	"context"
	"net/http"
	"time"

	"github.com/auto-dns/pihole-cluster-admin/internal/config"
	"github.com/go-chi/chi"
	"github.com/rs/zerolog"
)

type Server struct {
	cfg    *config.ServerConfig
	logger zerolog.Logger
	router chi.Router
	http   *http.Server
}

func New(http *http.Server, router chi.Router, cfg *config.ServerConfig, logger zerolog.Logger) *Server {
	return &Server{
		cfg:    cfg,
		logger: logger,
		router: router,
		http:   http,
	}
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
