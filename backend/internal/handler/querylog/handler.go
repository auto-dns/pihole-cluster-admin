package querylog

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/auto-dns/pihole-cluster-admin/internal/domain"
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
	var req domain.QueryLogRequest
	ctxLogger := h.logger.With()

	cursor := r.URL.Query().Get("cursor")
	if cursor != "" {
		// Cursor request: only cursor and optional length override
		req.Cursor = &cursor
		if v := r.URL.Query().Get("length"); v != "" {
			i, err := strconv.Atoi(v)
			if err != nil || i < 0 {
				httpx.WriteJSONError(w, "invalid length", http.StatusBadRequest)
				return
			}
			req.Length = &i
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
			req.Filters.From = &from
			req.Filters.Until = &until
			ctxLogger.Time("from", from).Time("until", until)
		} else {
			if fromStr != "" {
				t, err := time.Parse(time.RFC3339, fromStr)
				if err != nil {
					httpx.WriteJSONError(w, "invalid 'from' time", http.StatusBadRequest)
					return
				}
				req.Filters.From = &t
				ctxLogger.Time("from", t)
			}
			if untilStr != "" {
				t, err := time.Parse(time.RFC3339, untilStr)
				if err != nil {
					httpx.WriteJSONError(w, "invalid 'until' time", http.StatusBadRequest)
					return
				}
				req.Filters.Until = &t
				ctxLogger.Time("until", t)
			}
		}

		// --- Parse filters only when not using cursor
		if v := r.URL.Query().Get("length"); v != "" {
			i, err := strconv.Atoi(v)
			if err != nil || i < 0 {
				httpx.WriteJSONError(w, "invalid length", http.StatusBadRequest)
				return
			}
			req.Length = &i
			ctxLogger.Int("length", i)
		}
		if v := r.URL.Query().Get("start"); v != "" {
			i, err := strconv.Atoi(v)
			if err != nil || i < 0 {
				httpx.WriteJSONError(w, "invalid start", http.StatusBadRequest)
				return
			}
			req.Start = &i
			ctxLogger.Int("start", i)
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
			if err != nil {
				httpx.WriteJSONError(w, "invalid start", http.StatusBadRequest)
				return
			}
			req.Filters.Disk = &b
			ctxLogger.Bool("disk", b)
		}
	}

	logger := ctxLogger.Logger()
	logger.Debug().Msg("fetching query logs")

	res, err := h.service.Fetch(r.Context(), req)
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
	json.NewEncoder(w).Encode(res)
}
