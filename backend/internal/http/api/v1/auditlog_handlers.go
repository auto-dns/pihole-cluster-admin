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
