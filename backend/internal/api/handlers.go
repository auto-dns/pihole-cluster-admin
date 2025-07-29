package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/auto-dns/pihole-cluster-admin/internal/pihole"
	"github.com/rs/zerolog"
)

func ptrInt64(v int64) *int64 { return &v }

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
	// --- Build QueryOptions
	opts := pihole.FetchLogsQueryOptions{}

	// --- Parse optional timestamps (RFC3339)
	fromStr := r.URL.Query().Get("from")
	untilStr := r.URL.Query().Get("until")

	if fromStr == "" && untilStr == "" {
		until := time.Now().UTC()
		from := until.Add(-5 * time.Minute)
		opts.From = ptrInt64(from.Unix())
		opts.Until = ptrInt64(until.Unix())
	} else {
		if fromStr != "" {
			fromTime, err := time.Parse(time.RFC3339, fromStr)
			if err != nil {
				http.Error(w, "invalid 'from' time", http.StatusBadRequest)
				return
			}
			opts.From = ptrInt64(fromTime.Unix())
		}
		if untilStr != "" {
			untilTime, err := time.Parse(time.RFC3339, untilStr)
			if err != nil {
				http.Error(w, "invalid 'until' time", http.StatusBadRequest)
				return
			}
			opts.Until = ptrInt64(untilTime.Unix())
		}
	}

	// --- Parse other optional filters
	if v := r.URL.Query().Get("length"); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			opts.Length = &i
		}
	}
	if v := r.URL.Query().Get("start"); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			opts.Start = &i
		}
	}
	if v := r.URL.Query().Get("cursor"); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			opts.Cursor = &i
		}
	}
	if v := r.URL.Query().Get("domain"); v != "" {
		opts.Domain = &v
	}
	if v := r.URL.Query().Get("client_ip"); v != "" {
		opts.ClientIP = &v
	}
	if v := r.URL.Query().Get("client_name"); v != "" {
		opts.ClientName = &v
	}
	if v := r.URL.Query().Get("upstream"); v != "" {
		opts.Upstream = &v
	}
	if v := r.URL.Query().Get("type"); v != "" {
		opts.Type = &v
	}
	if v := r.URL.Query().Get("status"); v != "" {
		opts.Status = &v
	}
	if v := r.URL.Query().Get("reply"); v != "" {
		opts.Reply = &v
	}
	if v := r.URL.Query().Get("dnssec"); v != "" {
		opts.DNSSEC = &v
	}
	if v := r.URL.Query().Get("disk"); v != "" {
		b, err := strconv.ParseBool(v)
		if err == nil {
			opts.Disk = &b
		}
	}

	// --- Call cluster client
	queryLogResponses, errs := h.cluster.FetchLogs(opts)

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
