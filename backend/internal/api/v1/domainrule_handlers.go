package v1

import (
	"encoding/json"
	"net/http"

	"github.com/auto-dns/pihole-cluster-admin/internal/domain"
	"github.com/auto-dns/pihole-cluster-admin/internal/transport/httpx"
	"github.com/go-chi/chi"
)

func registerDomainRules(r chi.Router, d Deps) {
	// Read
	r.Get("/domain", domainRuleGetByTypeKindDomain(d))
	r.Get("/domain/type/{type}", domainRuleGetByTypeKindDomain(d))
	r.Get("/domain/kind/{kind}", domainRuleGetByTypeKindDomain(d))
	r.Get("/domain/domain/{domain}", domainRuleGetByTypeKindDomain(d))
	r.Get("/domain/type/{type}/kind/{kind}", domainRuleGetByTypeKindDomain(d))
	r.Get("/domain/type/{type}/kind/{kind}/domain/{domain}", domainRuleGetByTypeKindDomain(d))
	// Write
	r.Post("/domain/type/{type}/kind/{kind}", domainRuleAddDomainRule(d))
	r.Delete("/domain/type/{type}/kind/{kind}/domain/{domain}", domainRuleRemoveDomainRule(d))
}

// GET /domains, /type/{type}, etc.
func domainRuleGetByTypeKindDomain(d Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		typeString := chi.URLParam(r, "type")
		kindString := chi.URLParam(r, "kind")
		domainString := chi.URLParam(r, "domain")

		ruleType, ok := parseRuleType(typeString)
		if !ok {
			d.Logger.Error().Msg("bad \"type\" parameter")
			httpx.WriteJSONError(w, "bad \"type\" parameter", http.StatusBadRequest)
			return
		}

		ruleKind, ok := parseRuleKind(kindString)
		if !ok {
			d.Logger.Error().Msg("bad \"kind\" parameter")
			httpx.WriteJSONError(w, "bad \"kind\" parameter", http.StatusBadRequest)
			return
		}

		if domainString == "" {
			d.Logger.Error().Msg("empty \"domain\" parmeter")
			httpx.WriteJSONError(w, "empty \"domain\" parmeter", http.StatusBadRequest)
			return
		}

		q := domain.ListDomainRulesQuery{
			Type:   &ruleType,
			Kind:   &ruleKind,
			Domain: &domainString,
		}
		results := d.DomainRuleService.List(r.Context(), q)

		for _, nr := range results {
			if nr.Error != nil {
				d.Logger.Warn().Err(nr.Error).Msg("partial failure getting domain rules")
			}
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(results); err != nil {
			d.Logger.Error().Err(err).Msg("failed to encode response")
			httpx.WriteJSONError(w, "failed to encode response", http.StatusInternalServerError)
		}
	}
}

func domainRuleAddDomainRule(d Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		typeString := chi.URLParam(r, "type")
		kindString := chi.URLParam(r, "kind")

		ruleType, ok := parseRuleType(typeString)
		if !ok {
			d.Logger.Error().Msg("bad \"type\" parameter")
			httpx.WriteJSONError(w, "bad \"type\" parameter", http.StatusBadRequest)
			return
		}

		ruleKind, ok := parseRuleKind(kindString)
		if !ok {
			d.Logger.Error().Msg("bad \"kind\" parameter")
			httpx.WriteJSONError(w, "bad \"kind\" parameter", http.StatusBadRequest)
			return
		}

		logger := d.Logger.With().Str("type", string(ruleType)).Str("kind", string(ruleKind)).Logger()

		// --- Parse JSON body
		var body addDomainRuleResponseDTO
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
		results := d.DomainRuleService.Add(r.Context(), cmd)

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
}

func domainRuleRemoveDomainRule(d Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		typeString := chi.URLParam(r, "type")
		kindString := chi.URLParam(r, "kind")
		domainString := chi.URLParam(r, "domain")

		ruleType, ok := parseRuleType(typeString)
		if !ok {
			d.Logger.Error().Msg("bad \"type\" parameter")
			httpx.WriteJSONError(w, "bad \"type\" parameter", http.StatusBadRequest)
			return
		}

		ruleKind, ok := parseRuleKind(kindString)
		if !ok {
			d.Logger.Error().Msg("bad \"kind\" parameter")
			httpx.WriteJSONError(w, "bad \"kind\" parameter", http.StatusBadRequest)
			return
		}

		if domainString == "" {
			d.Logger.Error().Msg("empty \"domain\" parmeter")
			httpx.WriteJSONError(w, "empty \"domain\" parmeter", http.StatusBadRequest)
			return
		}

		logger := d.Logger.With().Str("type", string(ruleType)).Str("kind", string(ruleKind)).Str("domain", domainString).Logger()
		logger.Debug().Msg("removing domain rule")

		cmd := domain.RemoveDomainRuleCommand{
			Type:   ruleType,
			Kind:   ruleKind,
			Domain: domainString,
		}
		results := d.DomainRuleService.Remove(r.Context(), cmd)

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
}
