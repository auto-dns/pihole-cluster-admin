package domainrulehandler

import (
	"encoding/json"
	"net/http"

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

func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()
	// Read
	r.Get("/", h.getAll)
	r.Get("/type/{type}", h.getByType)
	r.Get("/kind/{kind}", h.getByKind)
	r.Get("/domain/{domain}", h.getByDomain)
	r.Get("/type/{type}/kind/{kind}", h.getByTypeKind)
	r.Get("/type/{type}/kind/{kind}/domain/{domain}", h.getByTypeKindDomain)
	// Write
	r.Post("/type/{type}/kind/{kind}", h.addDomainRule)
	r.Delete("/type/{type}/kind/{kind}/domain/{domain}", h.removeDomainRule)
	return r
}

func (h *Handler) getAll(w http.ResponseWriter, r *http.Request) {
	results := h.service.GetAll(r.Context())

	for _, nr := range results {
		if nr.Error != nil {
			h.logger.Warn().Err(nr.Error).Msg("partial failure getting domain rule")
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(results); err != nil {
		h.logger.Error().Err(err).Msg("failed to encode response")
		httpx.WriteJSONError(w, "failed to encode response", http.StatusInternalServerError)
	}
}

func (h *Handler) getByType(w http.ResponseWriter, r *http.Request) {
	typeString := chi.URLParam(r, "type")
	ruleType, ok := pihole.ParseRuleType(typeString)
	if !ok {
		h.logger.Error().Msg("bad \"type\" parameter")
		httpx.WriteJSONError(w, "bad \"type\" parameter", http.StatusBadRequest)
		return
	}

	opts := pihole.GetDomainRulesByTypeOptions{
		Type: ruleType,
	}
	results := h.service.GetByType(r.Context(), opts)

	for _, nr := range results {
		if nr.Error != nil {
			h.logger.Warn().Err(nr.Error).Msg("partial failure getting domain rules")
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(results); err != nil {
		h.logger.Error().Err(err).Msg("failed to encode response")
		httpx.WriteJSONError(w, "failed to encode response", http.StatusInternalServerError)
	}
}

func (h *Handler) getByKind(w http.ResponseWriter, r *http.Request) {
	kindString := chi.URLParam(r, "kind")
	ruleKind, ok := pihole.ParseRuleKind(kindString)
	if !ok {
		h.logger.Error().Msg("bad \"kind\" parameter")
		httpx.WriteJSONError(w, "bad \"kind\" parameter", http.StatusBadRequest)
		return
	}

	opts := pihole.GetDomainRulesByKindOptions{
		Kind: ruleKind,
	}
	results := h.service.GetByKind(r.Context(), opts)

	for _, nr := range results {
		if nr.Error != nil {
			h.logger.Warn().Err(nr.Error).Msg("partial failure getting domain rules")
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(results); err != nil {
		h.logger.Error().Err(err).Msg("failed to encode response")
		httpx.WriteJSONError(w, "failed to encode response", http.StatusInternalServerError)
	}
}

func (h *Handler) getByDomain(w http.ResponseWriter, r *http.Request) {
	domainString := chi.URLParam(r, "domain")

	if len(domainString) == 0 {
		h.logger.Error().Msg("empty \"domain\" parmeter")
		httpx.WriteJSONError(w, "empty \"domain\" parmeter", http.StatusBadRequest)
		return
	}

	opts := pihole.GetDomainRulesByDomainOptions{
		Domain: domainString,
	}
	results := h.service.GetByDomain(r.Context(), opts)

	for _, nr := range results {
		if nr.Error != nil {
			h.logger.Warn().Err(nr.Error).Msg("partial failure getting domain rules")
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(results); err != nil {
		h.logger.Error().Err(err).Msg("failed to encode response")
		httpx.WriteJSONError(w, "failed to encode response", http.StatusInternalServerError)
	}
}

func (h *Handler) getByTypeKind(w http.ResponseWriter, r *http.Request) {
	typeString := chi.URLParam(r, "type")
	kindString := chi.URLParam(r, "kind")

	ruleType, ok := pihole.ParseRuleType(typeString)
	if !ok {
		h.logger.Error().Msg("bad \"type\" parameter")
		httpx.WriteJSONError(w, "bad \"type\" parameter", http.StatusBadRequest)
		return
	}

	ruleKind, ok := pihole.ParseRuleKind(kindString)
	if !ok {
		h.logger.Error().Msg("bad \"kind\" parameter")
		httpx.WriteJSONError(w, "bad \"kind\" parameter", http.StatusBadRequest)
		return
	}

	opts := pihole.GetDomainRulesByTypeKindOptions{
		Type: ruleType,
		Kind: ruleKind,
	}
	results := h.service.GetByTypeKind(r.Context(), opts)

	for _, nr := range results {
		if nr.Error != nil {
			h.logger.Warn().Err(nr.Error).Msg("partial failure getting domain rules")
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(results); err != nil {
		h.logger.Error().Err(err).Msg("failed to encode response")
		httpx.WriteJSONError(w, "failed to encode response", http.StatusInternalServerError)
	}
}

func (h *Handler) getByTypeKindDomain(w http.ResponseWriter, r *http.Request) {
	typeString := chi.URLParam(r, "type")
	kindString := chi.URLParam(r, "kind")
	domainString := chi.URLParam(r, "domain")

	ruleType, ok := pihole.ParseRuleType(typeString)
	if !ok {
		h.logger.Error().Msg("bad \"type\" parameter")
		httpx.WriteJSONError(w, "bad \"type\" parameter", http.StatusBadRequest)
		return
	}

	ruleKind, ok := pihole.ParseRuleKind(kindString)
	if !ok {
		h.logger.Error().Msg("bad \"kind\" parameter")
		httpx.WriteJSONError(w, "bad \"kind\" parameter", http.StatusBadRequest)
		return
	}

	if len(domainString) == 0 {
		h.logger.Error().Msg("empty \"domain\" parmeter")
		httpx.WriteJSONError(w, "empty \"domain\" parmeter", http.StatusBadRequest)
		return
	}

	opts := pihole.GetDomainRulesByTypeKindDomainOptions{
		Type:   ruleType,
		Kind:   ruleKind,
		Domain: domainString,
	}
	results := h.service.GetByTypeKindDomain(r.Context(), opts)

	for _, nr := range results {
		if nr.Error != nil {
			h.logger.Warn().Err(nr.Error).Msg("partial failure getting domain rules")
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(results); err != nil {
		h.logger.Error().Err(err).Msg("failed to encode response")
		httpx.WriteJSONError(w, "failed to encode response", http.StatusInternalServerError)
	}
}

func (h *Handler) addDomainRule(w http.ResponseWriter, r *http.Request) {
	typeString := chi.URLParam(r, "type")
	kindString := chi.URLParam(r, "kind")

	ruleType, ok := pihole.ParseRuleType(typeString)
	if !ok {
		h.logger.Error().Msg("bad \"type\" parameter")
		httpx.WriteJSONError(w, "bad \"type\" parameter", http.StatusBadRequest)
		return
	}

	ruleKind, ok := pihole.ParseRuleKind(kindString)
	if !ok {
		h.logger.Error().Msg("bad \"kind\" parameter")
		httpx.WriteJSONError(w, "bad \"kind\" parameter", http.StatusBadRequest)
		return
	}

	logger := h.logger.With().Str("type", string(ruleType)).Str("kind", string(ruleKind)).Logger()

	// --- Parse JSON body
	var body pihole.AddDomainPayload
	if err := httpx.DecodeJSONBody(w, r, &body, 1<<20); err != nil {
		logger.Error().Err(err).Msg("invalid JSON body")
		httpx.WriteJSONError(w, "invalid JSON body", http.StatusBadRequest)
		return
	}

	// --- Normalize Domain to []string
	var domains []string
	switch v := body.Domain.(type) {
	case string:
		domains = []string{v}
	case []interface{}:
		for _, item := range v {
			s, ok := item.(string)
			if !ok {
				logger.Error().Interface("domain", item).Msg("domain list must conain only strings")
				httpx.WriteJSONError(w, "domain list must contain only strings", http.StatusBadRequest)
				return
			}
			domains = append(domains, s)
		}
	default:
		logger.Error().Interface("domain", v).Msg("domain must be string or array of strings")
		httpx.WriteJSONError(w, "domain must be string or array of strings", http.StatusBadRequest)
		return
	}

	body.Domain = domains

	logger.Debug().Strs("domains", domains).Msg("adding domain rule")

	opts := pihole.AddDomainRuleOptions{
		Type:    ruleType,
		Kind:    ruleKind,
		Payload: body,
	}
	results := h.service.Add(r.Context(), opts)

	for _, nr := range results {
		if nr.Error != nil {
			logger.Warn().Err(nr.Error).Msg("partial failure adding domain rule")
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(results); err != nil {
		logger.Error().Err(err).Msg("failed to encode response")
		httpx.WriteJSONError(w, "failed to encode response", http.StatusInternalServerError)
	}
}

func (h *Handler) removeDomainRule(w http.ResponseWriter, r *http.Request) {
	typeString := chi.URLParam(r, "type")
	kindString := chi.URLParam(r, "kind")
	domainString := chi.URLParam(r, "domain")

	ruleType, ok := pihole.ParseRuleType(typeString)
	if !ok {
		h.logger.Error().Msg("bad \"type\" parameter")
		httpx.WriteJSONError(w, "bad \"type\" parameter", http.StatusBadRequest)
		return
	}

	ruleKind, ok := pihole.ParseRuleKind(kindString)
	if !ok {
		h.logger.Error().Msg("bad \"kind\" parameter")
		httpx.WriteJSONError(w, "bad \"kind\" parameter", http.StatusBadRequest)
		return
	}

	if len(domainString) == 0 {
		h.logger.Error().Msg("empty \"domain\" parmeter")
		httpx.WriteJSONError(w, "empty \"domain\" parmeter", http.StatusBadRequest)
		return
	}

	logger := h.logger.With().Str("type", string(ruleType)).Str("kind", string(ruleKind)).Str("domain", domainString).Logger()
	logger.Debug().Msg("removing domain rule")

	opts := pihole.RemoveDomainRuleOptions{
		Type:   ruleType,
		Kind:   ruleKind,
		Domain: domainString,
	}
	results := h.service.Remove(r.Context(), opts)

	for _, nr := range results {
		if nr.Error != nil {
			logger.Warn().Err(nr.Error).Msg("partial failure removing domain rule")
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(results); err != nil {
		logger.Error().Err(err).Msg("failed to encode response")
		httpx.WriteJSONError(w, "failed to encode response", http.StatusInternalServerError)
	}
}

func parseDomainPath(parts []string) (typeParam *pihole.RuleType, kindParam *pihole.RuleKind, domainParam *string) {
	switch len(parts) {
	case 0:
		// nothing: get all domains
		return nil, nil, nil

	case 1:
		// 1-part combos:
		p := parts[0]
		if rt, ok := pihole.ParseRuleType(p); ok {
			typeParam = &rt
			return
		}
		if rk, ok := pihole.ParseRuleKind(p); ok {
			kindParam = &rk
			return
		}
		domainParam = &p
		return

	case 2:
		// 2-part combos:
		p1, p2 := parts[0], parts[1]
		if rt, ok := pihole.ParseRuleType(p1); ok {
			typeParam = &rt
			if rk, ok := pihole.ParseRuleKind(p2); ok {
				kindParam = &rk
			} else {
				domainParam = &p2
			}
		} else if rk, ok := pihole.ParseRuleKind(p1); ok {
			kindParam = &rk
			domainParam = &p2
		} else {
			// fallback: treat first as domain, second ignored (shouldn't happen in spec)
			domainParam = &p1
		}

	case 3:
		// 3-part combo: /allow|deny/exact|regex/domain
		p1, p2, p3 := parts[0], parts[1], parts[2]
		if rt, ok := pihole.ParseRuleType(p1); ok {
			typeParam = &rt
		}
		if rk, ok := pihole.ParseRuleKind(p2); ok {
			kindParam = &rk
		}
		domainParam = &p3
	}
	return
}
