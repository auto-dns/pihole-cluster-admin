package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
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
	var req pihole.FetchQueryLogRequest

	cursor := r.URL.Query().Get("cursor")
	if cursor != "" {
		// Cursor request: only cursor and optional length override
		req.CursorID = &cursor
		if v := r.URL.Query().Get("length"); v != "" {
			if i, err := strconv.Atoi(v); err == nil {
				req.Length = &i
			}
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
					http.Error(w, "invalid 'from' time", http.StatusBadRequest)
					return
				}
				req.Filters.From = ptrInt64(fromTime.Unix())
			}
			if untilStr != "" {
				untilTime, err := time.Parse(time.RFC3339, untilStr)
				if err != nil {
					http.Error(w, "invalid 'until' time", http.StatusBadRequest)
					return
				}
				req.Filters.Until = ptrInt64(untilTime.Unix())
			}
		}

		// --- Parse filters only when not using cursor
		if v := r.URL.Query().Get("length"); v != "" {
			if i, err := strconv.Atoi(v); err == nil {
				req.Length = &i
			}
		}
		if v := r.URL.Query().Get("start"); v != "" {
			if i, err := strconv.Atoi(v); err == nil {
				req.Start = &i
			}
		}
		if v := r.URL.Query().Get("domain"); v != "" {
			req.Filters.Domain = &v
		}
		if v := r.URL.Query().Get("client_ip"); v != "" {
			req.Filters.ClientIP = &v
		}
		if v := r.URL.Query().Get("client_name"); v != "" {
			req.Filters.ClientName = &v
		}
		if v := r.URL.Query().Get("upstream"); v != "" {
			req.Filters.Upstream = &v
		}
		if v := r.URL.Query().Get("type"); v != "" {
			req.Filters.Type = &v
		}
		if v := r.URL.Query().Get("status"); v != "" {
			req.Filters.Status = &v
		}
		if v := r.URL.Query().Get("reply"); v != "" {
			req.Filters.Reply = &v
		}
		if v := r.URL.Query().Get("dnssec"); v != "" {
			req.Filters.DNSSEC = &v
		}
		if v := r.URL.Query().Get("disk"); v != "" {
			b, err := strconv.ParseBool(v)
			if err == nil {
				req.Filters.Disk = &b
			}
		}
	}

	// --- Call cluster client
	res, err := h.cluster.FetchQueryLogs(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	for _, nr := range res.Results {
		if nr.Error != "" {
			h.logger.Warn().Str("error", nr.Error).Msg("partial failure fetching logs")
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(res); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}

func parseDomainPath(parts []string) (typeParam, kindParam, domainParam *string) {
	switch len(parts) {
	case 0:
		// nothing: get all domains
		return nil, nil, nil

	case 1:
		// 1-part combos:
		p := parts[0]
		switch p {
		case "allow", "deny":
			typeParam = &p
		case "exact", "regex":
			kindParam = &p
		default:
			domainParam = &p
		}

	case 2:
		// 2-part combos:
		p1 := parts[0]
		p2 := parts[1]
		if p1 == "allow" || p1 == "deny" {
			typeParam = &p1
			if p2 == "exact" || p2 == "regex" {
				kindParam = &p2
			} else {
				domainParam = &p2
			}
		} else if p1 == "exact" || p1 == "regex" {
			kindParam = &p1
			domainParam = &p2
		} else {
			// fallback: treat first as domain, second ignored (shouldn't happen in spec)
			domainParam = &p1
		}

	case 3:
		// 3-part combo: /allow|deny/exact|regex/domain
		p1 := parts[0]
		p2 := parts[1]
		p3 := parts[2]
		if p1 == "allow" || p1 == "deny" {
			typeParam = &p1
		}
		if p2 == "exact" || p2 == "regex" {
			kindParam = &p2
		}
		domainParam = &p3
	}
	return
}

func (h *Handler) HandleGetDomainRules(w http.ResponseWriter, r *http.Request) {
	// Path after /api/domains
	suffix := strings.TrimPrefix(r.URL.Path, "/api/domains")
	suffix = strings.TrimPrefix(suffix, "/")
	parts := []string{}
	if suffix != "" {
		parts = strings.Split(suffix, "/")
	}

	typeParam, kindParam, domainParam := parseDomainPath(parts)

	opts := pihole.GetDomainRulesOptions{
		Type:   typeParam,
		Kind:   kindParam,
		Domain: domainParam,
	}

	results := h.cluster.GetDomainRules(opts)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(results); err != nil {
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
