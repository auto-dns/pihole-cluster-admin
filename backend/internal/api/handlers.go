package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/auto-dns/pihole-cluster-admin/internal/config"
	"github.com/auto-dns/pihole-cluster-admin/internal/health"
	"github.com/auto-dns/pihole-cluster-admin/internal/pihole"
	"github.com/rs/zerolog"
)

func ptrInt64(v int64) *int64 { return &v }

type Handler struct {
	cluster         piholeCluster
	sessions        sessionDeps
	initStatusStore initStatusStore
	piholeStore     piholeStore
	userStore       userStore
	healthService   healthService
	eventSubscriber eventSubscriber
	logger          zerolog.Logger
	cfg             config.ServerConfig
}

func NewHandler(cluster piholeCluster, sessions sessionDeps, initStatusStore initStatusStore, piholeStore piholeStore, userStore userStore, healthService healthService, eventSubscriber eventSubscriber, cfg config.ServerConfig, logger zerolog.Logger) *Handler {
	return &Handler{
		cluster:         cluster,
		sessions:        sessions,
		initStatusStore: initStatusStore,
		piholeStore:     piholeStore,
		userStore:       userStore,
		healthService:   healthService,
		eventSubscriber: eventSubscriber,
		logger:          logger,
		cfg:             cfg,
	}
}

// Convenience function

// Handlers

func (h *Handler) AuthMiddleware(next http.Handler) http.Handler {
	return h.sessions.AuthMiddleware(next)
}

// Unauthenticated routes

func (h *Handler) Healthcheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status": "OK"}`))
}

// Authenticated routes

// -- Health Status

func (h *Handler) GetHealthSummary(w http.ResponseWriter, r *http.Request) {
	healthSummary := h.healthService.Summary()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(healthSummary)
}

func (h *Handler) GetNodeHealth(w http.ResponseWriter, r *http.Request) {
	nodeHealth := h.healthService.NodeHealth()
	nodeHealthSlice := make([]health.NodeHealth, 0, len(nodeHealth))
	for _, value := range nodeHealth {
		nodeHealthSlice = append(nodeHealthSlice, value)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(nodeHealthSlice)
}

// Application business logic routes

func (h *Handler) FetchQueryLogs(w http.ResponseWriter, r *http.Request) {
	var req pihole.FetchQueryLogClusterRequest
	ctxLogger := h.logger.With()

	cursor := r.URL.Query().Get("cursor")
	if cursor != "" {
		// Cursor request: only cursor and optional length override
		req.Cursor = &cursor
		if v := r.URL.Query().Get("length"); v != "" {
			if i, err := strconv.Atoi(v); err == nil {
				req.Length = &i
			}
		}
		ctxLogger.Str("cursor", cursor)
		if req.Length != nil {
			ctxLogger.Int("length", *req.Length)
		}
	} else {
		// --- Parse optional timestamps (RFC3339)
		fromStr := r.URL.Query().Get("from")
		untilStr := r.URL.Query().Get("until")

		if fromStr == "" && untilStr == "" {
			until := time.Now().UTC()
			from := until.Add(-5 * time.Minute)
			req.Filters.From = ptrInt64(from.Unix())
			req.Filters.Until = ptrInt64(until.Unix())
		} else {
			if fromStr != "" {
				fromTime, err := time.Parse(time.RFC3339, fromStr)
				if err != nil {
					writeJSONError(w, "invalid 'from' time", http.StatusBadRequest)
					return
				}
				req.Filters.From = ptrInt64(fromTime.Unix())
			}
			if untilStr != "" {
				untilTime, err := time.Parse(time.RFC3339, untilStr)
				if err != nil {
					writeJSONError(w, "invalid 'until' time", http.StatusBadRequest)
					return
				}
				req.Filters.Until = ptrInt64(untilTime.Unix())
			}
		}
		ctxLogger.Int64("from", *req.Filters.From).Int64("until", *req.Filters.Until)

		// --- Parse filters only when not using cursor
		if v := r.URL.Query().Get("length"); v != "" {
			if i, err := strconv.Atoi(v); err == nil {
				req.Length = &i
				ctxLogger.Int("length", i)
			}
		}
		if v := r.URL.Query().Get("start"); v != "" {
			if i, err := strconv.Atoi(v); err == nil {
				req.Start = &i
				ctxLogger.Int("start", i)
			}
		}
		if v := r.URL.Query().Get("domain"); v != "" {
			ctxLogger.Str("domain", v)
			req.Filters.Domain = &v
		}
		if v := r.URL.Query().Get("client_ip"); v != "" {
			ctxLogger.Str("client_ip", v)
			req.Filters.ClientIP = &v
		}
		if v := r.URL.Query().Get("client_name"); v != "" {
			ctxLogger.Str("client_name", v)
			req.Filters.ClientName = &v
		}
		if v := r.URL.Query().Get("upstream"); v != "" {
			ctxLogger.Str("upstream", v)
			req.Filters.Upstream = &v
		}
		if v := r.URL.Query().Get("type"); v != "" {
			ctxLogger.Str("type", v)
			req.Filters.Type = &v
		}
		if v := r.URL.Query().Get("status"); v != "" {
			ctxLogger.Str("status", v)
			req.Filters.Status = &v
		}
		if v := r.URL.Query().Get("reply"); v != "" {
			ctxLogger.Str("reply", v)
			req.Filters.Reply = &v
		}
		if v := r.URL.Query().Get("dnssec"); v != "" {
			ctxLogger.Str("dnssec", v)
			req.Filters.DNSSEC = &v
		}
		if v := r.URL.Query().Get("disk"); v != "" {
			b, err := strconv.ParseBool(v)
			if err == nil {
				ctxLogger.Bool("disk", b)
				req.Filters.Disk = &b
			}
		}
	}

	logger := ctxLogger.Logger()
	logger.Debug().Msg("fetching query logs")

	// --- Call cluster client
	res, err := h.cluster.FetchQueryLogs(r.Context(), req)
	if err != nil {
		writeJSONError(w, err.Error(), http.StatusBadRequest)
		return
	}

	for _, nr := range res.Results {
		if nr.Error != nil {
			h.logger.Warn().Err(nr.Error).Msg("partial failure fetching logs")
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(res); err != nil {
		h.logger.Error().Err(err).Msg("failed to encode response")
		writeJSONError(w, "failed to encode response", http.StatusInternalServerError)
	}
}
