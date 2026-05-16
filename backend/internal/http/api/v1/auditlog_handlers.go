package v1

import (
	"net/http"
	"strconv"

	"github.com/auto-dns/pihole-cluster-admin/internal/domain"
	"github.com/auto-dns/pihole-cluster-admin/internal/http/transport"
	"github.com/go-chi/chi"
)


func registerAuditLog(r chi.Router, d Deps) {
	r.Get("/audit", auditLogList(d))
	r.Post("/audit/{id}/rollback", auditLogRollback(d))
}

func auditLogRollback(d Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := chi.URLParam(r, "id")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil || id <= 0 {
			transport.WriteBadRequestErr(w, "invalid id", err)
			return
		}

		entry, err := d.AuditLogService.GetById(r.Context(), id)
		if err != nil {
			d.Logger.Error().Err(err).Int64("id", id).Msg("audit entry not found for rollback")
			transport.WriteErr(w, err)
			return
		}

		if entry.Action != domain.AuditActionAddDomainRule && entry.Action != domain.AuditActionRemoveDomainRule {
			transport.WriteBadRequestErr(w, "only domain rule entries can be rolled back", nil)
			return
		}

		if entry.TargetDomain == nil || entry.RuleType == nil || entry.RuleKind == nil {
			d.Logger.Error().Int64("id", id).Msg("audit entry missing required fields for rollback")
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		ruleType := domain.RuleType(*entry.RuleType)
		ruleKind := domain.RuleKind(*entry.RuleKind)
		targetDomain := *entry.TargetDomain

		var nodes []rollbackNodeResultDTO
		if entry.Action == domain.AuditActionAddDomainRule {
			results := d.DomainRuleService.Remove(r.Context(), domain.RemoveDomainRuleCommand{
				Type:   ruleType,
				Kind:   ruleKind,
				Domain: targetDomain,
			})
			nodes = make([]rollbackNodeResultDTO, 0, len(results))
			for _, nr := range results {
				node := rollbackNodeResultDTO{
					NodeId:   nr.PiholeNode.Id,
					NodeName: nr.PiholeNode.Name,
					Success:  nr.Success,
				}
				if nr.Error != nil {
					node.Error = nr.Error.Error()
				}
				nodes = append(nodes, node)
			}
		} else {
			results := d.DomainRuleService.Add(r.Context(), domain.AddDomainRulesCommand{
				Type:    ruleType,
				Kind:    ruleKind,
				Domains: []string{targetDomain},
			})
			nodes = make([]rollbackNodeResultDTO, 0, len(results))
			for _, nr := range results {
				node := rollbackNodeResultDTO{
					NodeId:   nr.PiholeNode.Id,
					NodeName: nr.PiholeNode.Name,
					Success:  nr.Success,
				}
				if nr.Error != nil {
					node.Error = nr.Error.Error()
				}
				nodes = append(nodes, node)
			}
		}

		transport.WriteJSON(w, http.StatusOK, rollbackResponseDTO{
			OriginalId: id,
			Nodes:      nodes,
		})
	}
}

func auditLogList(d Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		limit := 50
		offset := 0

		if v := r.URL.Query().Get("limit"); v != "" {
			if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 200 {
				limit = n
			}
		}
		if v := r.URL.Query().Get("offset"); v != "" {
			if n, err := strconv.Atoi(v); err == nil && n >= 0 {
				offset = n
			}
		}

		entries, total, err := d.AuditLogService.List(r.Context(), domain.ListAuditEntriesQuery{
			Limit:  limit,
			Offset: offset,
		})
		if err != nil {
			d.Logger.Error().Err(err).Msg("failed to list audit log entries")
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		dtos := make([]auditEntryDTO, 0, len(entries))
		for _, e := range entries {
			dtos = append(dtos, toAuditEntryDTO(e))
		}

		transport.WriteJSON(w, http.StatusOK, listAuditResponseDTO{
			Entries: dtos,
			Total:   total,
			Limit:   limit,
			Offset:  offset,
		})
	}
}
