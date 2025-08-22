package querylog

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/auto-dns/pihole-cluster-admin/internal/pihole"
	"github.com/auto-dns/pihole-cluster-admin/internal/transport/httpx"
	"github.com/go-chi/chi"
	"github.com/rs/zerolog"
)

type Handler struct {
	service service
	logger  zerolog.Logger
}

func NewHandler(service service, logger zerolog.Logger) *Handler {
	return &Handler{service: service, logger: logger}
}

func (h *Handler) Register(r chi.Router) {
	// Read
	r.Get("/", h.getQueryLogs)
}

func (h *Handler) getQueryLogs(w http.ResponseWriter, r *http.Request) {
	// --- Parse/validate query params (transport concern)
	var body pihole.FetchQueryLogClusterRequest
	ctxLogger := h.logger.With()

	cursor := r.URL.Query().Get("cursor")
	if cursor != "" {
		// Cursor request: only cursor and optional length override
		body.Cursor = &cursor
		if v := r.URL.Query().Get("length"); v != "" {
			i, err := strconv.Atoi(v)
			if err != nil || i < 0 {
				httpx.WriteJSONError(w, "invalid length", http.StatusBadRequest)
				return
			}
			body.Length = &i
			ctxLogger.Int("length", i)
		}
		ctxLogger.Str("cursor", cursor)
	} else {
		// --- Parse optional timestamps (RFC3339)
		fromStr := r.URL.Query().Get("from")
		untilStr := r.URL.Query().Get("until")

		if fromStr == "" && untilStr == "" {
			until := time.Now().UTC()
			from := until.Add(-5 * time.Minute)
			tu, tf := until.Unix(), from.Unix()
			body.Filters.Until = &tu
			body.Filters.From = &tf
			ctxLogger.Int64("from", tf).Int64("until", tu)
		} else {
			if fromStr != "" {
				t, err := time.Parse(time.RFC3339, fromStr)
				if err != nil {
					httpx.WriteJSONError(w, "invalid 'from' time", http.StatusBadRequest)
					return
				}
				ts := t.Unix()
				body.Filters.From = &ts
				ctxLogger.Int64("from", ts)
			}
			if untilStr != "" {
				t, err := time.Parse(time.RFC3339, untilStr)
				if err != nil {
					httpx.WriteJSONError(w, "invalid 'until' time", http.StatusBadRequest)
					return
				}
				ts := t.Unix()
				body.Filters.Until = &ts
				ctxLogger.Int64("until", ts)
			}
		}

		// --- Parse filters only when not using cursor
		if v := r.URL.Query().Get("length"); v != "" {
			i, err := strconv.Atoi(v)
			if err != nil || i < 0 {
				httpx.WriteJSONError(w, "invalid length", http.StatusBadRequest)
				return
			}
			body.Length = &i
			ctxLogger.Int("length", i)
		}
		if v := r.URL.Query().Get("start"); v != "" {
			i, err := strconv.Atoi(v)
			if err != nil || i < 0 {
				httpx.WriteJSONError(w, "invalid start", http.StatusBadRequest)
				return
			}
			body.Start = &i
			ctxLogger.Int("start", i)
		}
		if v := r.URL.Query().Get("domain"); v != "" {
			ctxLogger.Str("domain", v)
			body.Filters.Domain = &v
		}
		if v := r.URL.Query().Get("client_ip"); v != "" {
			ctxLogger.Str("client_ip", v)
			body.Filters.ClientIP = &v
		}
		if v := r.URL.Query().Get("client_name"); v != "" {
			ctxLogger.Str("client_name", v)
			body.Filters.ClientName = &v
		}
		if v := r.URL.Query().Get("upstream"); v != "" {
			ctxLogger.Str("upstream", v)
			body.Filters.Upstream = &v
		}
		if v := r.URL.Query().Get("type"); v != "" {
			ctxLogger.Str("type", v)
			body.Filters.Type = &v
		}
		if v := r.URL.Query().Get("status"); v != "" {
			ctxLogger.Str("status", v)
			body.Filters.Status = &v
		}
		if v := r.URL.Query().Get("reply"); v != "" {
			ctxLogger.Str("reply", v)
			body.Filters.Reply = &v
		}
		if v := r.URL.Query().Get("dnssec"); v != "" {
			ctxLogger.Str("dnssec", v)
			body.Filters.DNSSEC = &v
		}
		if v := r.URL.Query().Get("disk"); v != "" {
			b, err := strconv.ParseBool(v)
			if err != nil {
				httpx.WriteJSONError(w, "invalid start", http.StatusBadRequest)
				return
			}
			body.Filters.Disk = &b
			ctxLogger.Bool("disk", b)
		}
	}

	logger := ctxLogger.Logger()
	logger.Debug().Msg("fetching query logs")

	res, err := h.service.Fetch(r.Context(), body)
	if err != nil {
		httpx.WriteJSONError(w, err.Error(), http.StatusBadRequest)
		return
	}

	for _, nr := range res.Results {
		if nr.Error != nil {
			h.logger.Warn().Err(nr.Error).Msg("partial failure fetching logs")
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(res)
}
