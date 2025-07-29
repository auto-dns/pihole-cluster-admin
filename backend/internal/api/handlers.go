package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/auto-dns/pihole-cluster-admin/internal/pihole"
	"github.com/rs/zerolog"
)

type Handler struct {
	cluster pihole.ClusterInterface
	logger  zerolog.Logger
}

func NewHandler(cluster pihole.ClusterInterface, logger zerolog.Logger) *Handler {
	return &Handler{
		cluster: cluster,
		logger:  logger,
	}
}

func (h *Handler) Healthcheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"status": "OK"}`))
}

func (h *Handler) FetchLogs(w http.ResponseWriter, r *http.Request) {
	fromStr := r.URL.Query().Get("from")
	untilStr := r.URL.Query().Get("until")

	var fromTime, untilTime time.Time
	var err error

	// --- Default range if both are empty (last 5 minutes)
	if fromStr == "" && untilStr == "" {
		untilTime = time.Now().UTC()
		fromTime = untilTime.Add(-5 * time.Minute)
	} else {
		if fromStr != "" {
			fromTime, err = time.Parse(time.RFC3339, fromStr)
			if err != nil {
				http.Error(w, "invalid 'from' time", http.StatusBadRequest)
				return
			}
		}

		if untilStr != "" {
			untilTime, err = time.Parse(time.RFC3339, untilStr)
			if err != nil {
				http.Error(w, "invalid 'until' time", http.StatusBadRequest)
				return
			}
		}

		if fromStr != "" && untilStr == "" {
			untilTime = time.Now().UTC()
		}
		if fromStr == "" && untilStr != "" {
			fromTime = untilTime.Add(-5 * time.Minute)
		}
	}

	fromUnix := fromTime.Unix()
	untilUnix := untilTime.Unix()

	queryLogResponses, errs := h.cluster.FetchLogs(fromUnix, untilUnix)

	// Log partial failures (but still return partial results)
	for _, e := range errs {
		if e != nil {
			h.logger.Warn().Err(e).Msg("partial failure fetching logs")
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(queryLogResponses); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}
