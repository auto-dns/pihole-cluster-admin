package domainrule

import (
	"encoding/json"
	"net/http"

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
	r.Get("/", h.getByTypeKindDomain)
	r.Get("/type/{type}", h.getByTypeKindDomain)
	r.Get("/kind/{kind}", h.getByTypeKindDomain)
	r.Get("/domain/{domain}", h.getByTypeKindDomain)
	r.Get("/type/{type}/kind/{kind}", h.getByTypeKindDomain)
	r.Get("/type/{type}/kind/{kind}/domain/{domain}", h.getByTypeKindDomain)
	// Write
	r.Post("/type/{type}/kind/{kind}", h.addDomainRule)
	r.Delete("/type/{type}/kind/{kind}/domain/{domain}", h.removeDomainRule)
}

// GET /domains, /type/{type}, etc.
func (h *Handler) getByTypeKindDomain(w http.ResponseWriter, r *http.Request) {
	typeString := chi.URLParam(r, "type")
	kindString := chi.URLParam(r, "kind")
	domainString := chi.URLParam(r, "domain")

	ruleType, ok := parseRuleType(typeString)
	if !ok {
		h.logger.Error().Msg("bad \"type\" parameter")
		httpx.WriteJSONError(w, "bad \"type\" parameter", http.StatusBadRequest)
		return
	}

	ruleKind, ok := parseRuleKind(kindString)
	if !ok {
		h.logger.Error().Msg("bad \"kind\" parameter")
		httpx.WriteJSONError(w, "bad \"kind\" parameter", http.StatusBadRequest)
		return
	}

	if domainString == "" {
		h.logger.Error().Msg("empty \"domain\" parmeter")
		httpx.WriteJSONError(w, "empty \"domain\" parmeter", http.StatusBadRequest)
		return
	}

	q := domain.ListDomainRulesQuery{
		Type:   &ruleType,
		Kind:   &ruleKind,
		Domain: &domainString,
	}
	results := h.service.List(r.Context(), q)

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

	ruleType, ok := parseRuleType(typeString)
	if !ok {
		h.logger.Error().Msg("bad \"type\" parameter")
		httpx.WriteJSONError(w, "bad \"type\" parameter", http.StatusBadRequest)
		return
	}

	ruleKind, ok := parseRuleKind(kindString)
	if !ok {
		h.logger.Error().Msg("bad \"kind\" parameter")
		httpx.WriteJSONError(w, "bad \"kind\" parameter", http.StatusBadRequest)
		return
	}

	logger := h.logger.With().Str("type", string(ruleType)).Str("kind", string(ruleKind)).Logger()

	// --- Parse JSON body
	var body addDTO
	if err := httpx.DecodeJSONBody(w, r, &body, 1<<20); err != nil {
		logger.Error().Err(err).Msg("invalid JSON body")
		httpx.WriteJSONError(w, "invalid JSON body", http.StatusBadRequest)
		return
	}
	domains, err := normalizeDomains(body.Domain)
	if err != nil {
		logger.Error().Err(err).Msg("invalid JSON body")
		httpx.WriteJSONError(w, err.Error(), http.StatusBadRequest)
	}

	cmd := domain.AddDomainRulesCommand{
		Type:    ruleType,
		Kind:    ruleKind,
		Domains: domains,
	}

	logger.Debug().Strs("domains", domains).Msg("adding domain rule")
	results := h.service.Add(r.Context(), cmd)

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

	ruleType, ok := parseRuleType(typeString)
	if !ok {
		h.logger.Error().Msg("bad \"type\" parameter")
		httpx.WriteJSONError(w, "bad \"type\" parameter", http.StatusBadRequest)
		return
	}

	ruleKind, ok := parseRuleKind(kindString)
	if !ok {
		h.logger.Error().Msg("bad \"kind\" parameter")
		httpx.WriteJSONError(w, "bad \"kind\" parameter", http.StatusBadRequest)
		return
	}

	if domainString == "" {
		h.logger.Error().Msg("empty \"domain\" parmeter")
		httpx.WriteJSONError(w, "empty \"domain\" parmeter", http.StatusBadRequest)
		return
	}

	logger := h.logger.With().Str("type", string(ruleType)).Str("kind", string(ruleKind)).Str("domain", domainString).Logger()
	logger.Debug().Msg("removing domain rule")

	cmd := domain.RemoveDomainRuleCommand{
		Type:   ruleType,
		Kind:   ruleKind,
		Domain: domainString,
	}
	results := h.service.Remove(r.Context(), cmd)

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
