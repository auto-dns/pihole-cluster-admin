package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/auto-dns/pihole-cluster-admin/internal/pihole"
	"github.com/go-chi/chi"
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

func (h *Handler) FetchQueryLogs(w http.ResponseWriter, r *http.Request) {
	// --- Build QueryOptions
	opts := pihole.FetchQueryLogOptions{}

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
	nodeResults := h.cluster.FetchQueryLogs(opts)

	// Log partial failures (but still return partial results)
	for _, nr := range nodeResults {
		if nr.Error != "" {
			h.logger.Warn().Str("error", nr.Error).Msg("partial failure fetching logs")
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(nodeResults); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}

func (h *Handler) HandleAddDomainRule(w http.ResponseWriter, r *http.Request) {
	domainType := chi.URLParam(r, "type")
	domainKind := chi.URLParam(r, "kind")

	// --- Parse JSON body
	var payload pihole.AddDomainPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}

	// --- Normalize Domain to []string
	var domains []string
	switch v := payload.Domain.(type) {
	case string:
		domains = []string{v}
	case []interface{}:
		for _, item := range v {
			s, ok := item.(string)
			if !ok {
				http.Error(w, "domain list must contain only strings", http.StatusBadRequest)
				return
			}
			domains = append(domains, s)
		}
	default:
		http.Error(w, "domain must be string or array of strings", http.StatusBadRequest)
		return
	}

	payloadNormalized := payload
	payloadNormalized.Domain = domains

	opts := pihole.AddDomainRuleOptions{
		Type:    domainType,
		Kind:    domainKind,
		Payload: payloadNormalized,
	}

	nodeResults := h.cluster.AddDomainRule(opts)

	for _, nr := range nodeResults {
		if nr.Error != "" {
			h.logger.Warn().Str("error", nr.Error).Msg("partial failure adding domain rule")
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(nodeResults); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}

func (h *Handler) HandleRemoveDomainRule(w http.ResponseWriter, r *http.Request) {
	domainType := chi.URLParam(r, "type")
	domainKind := chi.URLParam(r, "kind")
	domain := chi.URLParam(r, "domain")

	results := h.cluster.RemoveDomainRule(pihole.RemoveDomainRuleOptions{
		Type:   domainType,
		Kind:   domainKind,
		Domain: domain,
	})

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(results); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}
