package app

import (
	"fmt"
	"net/http"
	"time"

	"github.com/auto-dns/pihole-cluster-admin/internal/config"
	"github.com/auto-dns/pihole-cluster-admin/internal/server"
	"github.com/go-chi/chi"
	"github.com/rs/zerolog"
)

func NewServer(cfg *config.ServerConfig, router chi.Router, logger zerolog.Logger) *server.Server {
	http := &http.Server{
		Addr:              fmt.Sprintf(":%d", cfg.Port),
		ReadHeaderTimeout: time.Duration(cfg.ReadHeaderTimeoutSeconds) * time.Second,
	}

	logger.Info().Int("port", cfg.Port).Bool("tls", cfg.TLSEnabled).Msg("server created")

	return server.New(http, router, cfg, logger)
}
