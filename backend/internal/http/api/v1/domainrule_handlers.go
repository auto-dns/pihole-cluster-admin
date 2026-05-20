package v1

import (
	"errors"
	"net/http"
	"strings"

	"github.com/auto-dns/pihole-cluster-admin/internal/domain"
	"github.com/auto-dns/pihole-cluster-admin/internal/http/transport"
	"github.com/auto-dns/pihole-cluster-admin/internal/util"
	"github.com/go-chi/chi"
	"github.com/rs/zerolog/log"
)

func registerDomainRules(r chi.Router, d Deps) {
	r.Route("/domain", func(r chi.Router) {
		// Read
		r.Get("/", domainRuleGetByTypeKindDomain(d))
		r.Get("/type/{type}", domainRuleGetByTypeKindDomain(d))
		r.Get("/kind/{kind}", domainRuleGetByTypeKindDomain(d))
		r.Get("/domain/{domain}", domainRuleGetByTypeKindDomain(d))
		r.Get("/type/{type}/kind/{kind}", domainRuleGetByTypeKindDomain(d))
		r.Get("/type/{type}/kind/{kind}/domain/{domain}", domainRuleGetByTypeKindDomain(d))
		// Write
		r.Post("/type/{type}/kind/{kind}", domainRuleAddDomainRule(d))
		r.Delete("/type/{type}/kind/{kind}/domain/{domain}", domainRuleRemoveDomainRule(d))
		// Parity sync
		r.Post("/parity/sync", domainRuleSyncParityRule(d))
		// Regex tester
		r.Post("/regex/test", domainRuleTestRegex(d))
	})
}

// GET /domains, /type/{type}, etc.
func domainRuleGetByTypeKindDomain(d Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		typeString := chi.URLParam(r, "type")
		kindString := chi.URLParam(r, "kind")
		domainString := chi.URLParam(r, "domain")

		q := domain.ListDomainRulesQuery{}

		if typeString != "" {
			ruleType, ok := parseRuleType(typeString)
			if !ok {
				d.Logger.Error().Msg("bad \"type\" parameter")
				transport.WriteBadRequestErr(w, "bad \"type\" parameter", errors.New("bad \"type\" parameter"))
				return
			}
			q.Type = &ruleType
		}

		if kindString != "" {
			ruleKind, ok := parseRuleKind(kindString)
			if !ok {
				d.Logger.Error().Msg("bad \"kind\" parameter")
				transport.WriteBadRequestErr(w, "bad \"kind\" parameter", errors.New("bad \"kind\" parameter"))
				return
			}
			q.Kind = &ruleKind
		}

		if domainString != "" {
			q.Domain = &domainString
		}
		results := d.DomainRuleService.List(r.Context(), q)

		dto := listDomainRulesResponseDTO{
			Nodes: make(map[int64]listNodeDTO, len(results)),
		}
		dto.Summary.TotalNodes = len(results)

		for id, nr := range results {
			node := listNodeDTO{
				Node: piholeNodeRefDTO{
					Id:   nr.PiholeNode.Id,
					Name: nr.PiholeNode.Name,
					Host: nr.PiholeNode.Host,
				},
				TookMS: 0,
				Error:  util.ErrorString(nr.Error),
			}

			if nr.Success && nr.Response != nil {
				// map rules
				if len(nr.Response.Rules) > 0 {
					node.Rules = make([]domainRuleDTO, 0, len(nr.Response.Rules))
					for _, r := range nr.Response.Rules {
						node.Rules = append(node.Rules, toDomainRuleDTO(r))
					}
					dto.Summary.TotalRules += len(nr.Response.Rules)
				} else {
					node.Rules = []domainRuleDTO{}
				}
				node.TookMS = nr.Response.Took.Milliseconds()
				dto.Summary.OkNodes++
			} else {
				dto.Summary.ErrorNodes++
				if node.Rules == nil {
					node.Rules = []domainRuleDTO{}
				}
			}

			dto.Nodes[id] = node
			if nr.Error != nil {
				d.Logger.Warn().Err(nr.Error).Msg("partial failure getting domain rules")
			}
		}

		transport.WriteJSON(w, http.StatusOK, dto)
	}
}

func domainRuleAddDomainRule(d Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		typeString := chi.URLParam(r, "type")
		kindString := chi.URLParam(r, "kind")

		ruleType, ok := parseRuleType(typeString)
		if !ok {
			d.Logger.Error().Msg("bad \"type\" parameter")
			transport.WriteBadRequestErr(w, "bad \"type\" parameter", errors.New("bad \"type\" parameter"))
			return
		}

		ruleKind, ok := parseRuleKind(kindString)
		if !ok {
			d.Logger.Error().Msg("bad \"kind\" parameter")
			transport.WriteBadRequestErr(w, "bad \"kind\" parameter", errors.New("bad \"kind\" parameter"))
			return
		}

		logger := d.Logger.With().Str("type", string(ruleType)).Str("kind", string(ruleKind)).Logger()

		// --- Parse JSON body
		var body addDomainRuleRequestDTO
		if err := transport.DecodeJSONBody(w, r, &body, 1<<20); err != nil {
			logger.Error().Err(err).Msg("invalid JSON body")
			transport.WriteErr(w, err)
			return
		}
		domains, err := normalizeDomains(body.Domain)
		if err != nil {
			logger.Error().Err(err).Msg("invalid JSON body")
			transport.WriteErr(w, err)
			return
		}

		cmd := domain.AddDomainRulesCommand{
			Type:    ruleType,
			Kind:    ruleKind,
			Domains: domains,
		}

		logger.Debug().Strs("domains", domains).Msg("adding domain rule")
		results := d.DomainRuleService.Add(r.Context(), cmd)

		dto := addDomainRuleResponseDTO{
			Nodes: make(map[int64]addDomainRuleNodeDTO, len(results)),
		}

		for id, nr := range results {
			node := addDomainRuleNodeDTO{
				Node: piholeNodeRefDTO{
					Id:   nr.PiholeNode.Id,
					Name: nr.PiholeNode.Name,
					Host: nr.PiholeNode.Host,
				},
				Error: util.ErrorString(nr.Error),
			}
			if nr.Response != nil {
				node.Result.TookMS = nr.Response.Took.Milliseconds()
			}

			if nr.Success && nr.Response != nil {
				if len(nr.Response.Rules) > 0 {
					node.Result.Domains = make([]domainRuleDTO, 0, len(nr.Response.Rules))
					for _, dr := range nr.Response.Rules {
						node.Result.Domains = append(node.Result.Domains, toDomainRuleDTO(dr))
					}
				}
				node.Result.Processed = toProcessedDTO(nr.Response.Processed)
			}

			dto.Nodes[id] = node
		}

		for _, nr := range results {
			if nr.Error != nil {
				logger.Warn().Err(nr.Error).Msg("partial failure adding domain rule")
			}
		}

		transport.WriteJSON(w, http.StatusOK, dto)
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
			transport.WriteBadRequestErr(w, "bad \"type\" parameter", errors.New("bad \"type\" parameter"))
			return
		}

		ruleKind, ok := parseRuleKind(kindString)
		if !ok {
			d.Logger.Error().Msg("bad \"kind\" parameter")
			transport.WriteBadRequestErr(w, "bad \"kind\" parameter", errors.New("bad \"kind\" parameter"))
			return
		}

		if domainString == "" {
			d.Logger.Error().Msg("empty \"domain\" parameter")
			transport.WriteBadRequestErr(w, "empty \"domain\" parameter", errors.New("empty \"domain\" parameter"))
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

		dto := removeDomainRuleResponseDTO{
			Nodes: make(map[int64]removeDomainRuleNodeDTO, len(results)),
		}

		// Totals
		total := len(results)
		removed := 0
		errors := 0

		for id, nr := range results {
			node := removeDomainRuleNodeDTO{
				Node: piholeNodeRefDTO{
					Id:   nr.PiholeNode.Id,
					Name: nr.PiholeNode.Name,
					Host: nr.PiholeNode.Host,
				},
				Removed: nr.Success && nr.Error == nil,
			}
			if nr.Error != nil {
				node.Error = nr.Error.Error()
			}

			if node.Removed {
				removed++
			}
			if node.Error != "" {
				errors++
				log.Warn().Err(nr.Error).Int64("node_id", id).Msg("partial failure removing domain rule")
			}

			dto.Nodes[id] = node
		}

		dto.Summary = removeSummaryDTO{
			Total:   total,
			Removed: removed,
			Failed:  total - removed,
			Errors:  errors,
		}

		transport.WriteJSON(w, http.StatusOK, dto)
	}
}

func domainRuleSyncParityRule(d Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body syncDomainRuleRequestDTO
		if err := transport.DecodeJSONBody(w, r, &body, 1<<20); err != nil {
			d.Logger.Error().Err(err).Msg("invalid JSON body")
			transport.WriteErr(w, err)
			return
		}

		ruleType, ok := parseRuleType(body.Type)
		if !ok {
			transport.WriteBadRequestErr(w, "bad \"type\"", errors.New("bad \"type\""))
			return
		}
		ruleKind, ok := parseRuleKind(body.Kind)
		if !ok {
			transport.WriteBadRequestErr(w, "bad \"kind\"", errors.New("bad \"kind\""))
			return
		}
		if body.Domain == "" {
			transport.WriteBadRequestErr(w, "\"domain\" is required", errors.New("\"domain\" is required"))
			return
		}

		cmd := domain.SyncDomainRuleCommand{
			Type:    ruleType,
			Kind:    ruleKind,
			Domain:  body.Domain,
			Comment: body.Comment,
		}

		logger := d.Logger.With().Str("type", string(ruleType)).Str("kind", string(ruleKind)).Str("domain", body.Domain).Logger()
		logger.Debug().Msg("syncing domain rule parity")

		results := d.DomainRuleService.SyncRule(r.Context(), cmd)

		dto := syncDomainRuleResponseDTO{
			Nodes: make(map[int64]syncDomainRuleNodeDTO, len(results)),
		}
		for id, nr := range results {
			node := syncDomainRuleNodeDTO{
				Node: piholeNodeRefDTO{
					Id:   nr.PiholeNode.Id,
					Name: nr.PiholeNode.Name,
					Host: nr.PiholeNode.Host,
				},
			}
			if nr.Error != nil {
				node.Error = nr.Error.Error()
			}
			if nr.Response != nil {
				node.AlreadyPresent = nr.Response.AlreadyPresent
				node.Added = nr.Response.Added
			}
			dto.Nodes[id] = node
			dto.Summary.TotalNodes++
			switch {
			case nr.Response != nil && nr.Response.AlreadyPresent:
				dto.Summary.AlreadyPresentNodes++
			case nr.Success:
				dto.Summary.SyncedNodes++
			default:
				dto.Summary.FailedNodes++
			}
		}

		transport.WriteJSON(w, http.StatusOK, dto)
	}
}

func domainRuleTestRegex(d Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body regexTestRequestDTO
		if err := transport.DecodeJSONBody(w, r, &body, 1<<20); err != nil {
			transport.WriteErr(w, err)
			return
		}
		testDomain := strings.TrimSpace(body.Domain)
		if testDomain == "" {
			transport.WriteBadRequestErr(w, "\"domain\" is required", errors.New("\"domain\" is required"))
			return
		}

		results := d.DomainRuleService.TestRegex(r.Context(), testDomain)

		dto := regexTestResponseDTO{
			Domain: testDomain,
			Nodes:  make([]regexTestNodeDTO, 0, len(results)),
		}
		for _, nr := range results {
			node := regexTestNodeDTO{
				NodeId:   nr.PiholeNode.Id,
				NodeName: nr.PiholeNode.Name,
				Success:  nr.Success,
				Matches:  make([]regexMatchDTO, 0),
			}
			if nr.Error != nil {
				node.Error = nr.Error.Error()
			}
			if nr.Success && nr.Response != nil {
				for _, m := range nr.Response.Matches {
					node.Matches = append(node.Matches, regexMatchDTO{
						ID:      m.ID,
						Pattern: m.Pattern,
						Kind:    m.Kind,
						Enabled: m.Enabled,
					})
				}
			}
			dto.Nodes = append(dto.Nodes, node)
		}

		transport.WriteJSON(w, http.StatusOK, dto)
	}
}
