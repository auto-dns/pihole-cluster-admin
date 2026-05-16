package v1

import (
	"net/http"

	"github.com/auto-dns/pihole-cluster-admin/internal/errs"
	"github.com/auto-dns/pihole-cluster-admin/internal/http/transport"
	"github.com/go-chi/chi"
)

func registerSync(r chi.Router, d Deps) {
	r.Post("/sync", syncFromNode(d))
}

func syncFromNode(d Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body syncRequestDTO
		if err := transport.DecodeJSONBody(w, r, &body, 1<<20); err != nil {
			transport.WriteErr(w, err)
			return
		}
		if body.SourceNodeId <= 0 {
			transport.WriteBadRequestErr(w, "sourceNodeId is required", nil)
			return
		}

		results, err := d.SyncService.SyncFromNode(r.Context(), body.SourceNodeId)
		if err != nil {
			if errs.KindOf(err) == errs.KindNotFound {
				transport.WriteNotFoundErr(w, "source node not found")
			} else if errs.KindOf(err) == errs.KindUnavailable {
				transport.WriteErr(w, err)
			} else {
				d.Logger.Error().Err(err).Msg("sync from node failed")
				http.Error(w, "internal server error", http.StatusInternalServerError)
			}
			return
		}

		dtos := make([]syncNodeResultDTO, 0, len(results))
		for _, r := range results {
			dtos = append(dtos, toSyncNodeResultDTO(r))
		}

		transport.WriteJSON(w, http.StatusOK, syncResponseDTO{
			SourceNodeId: body.SourceNodeId,
			Nodes:        dtos,
		})
	}
}
