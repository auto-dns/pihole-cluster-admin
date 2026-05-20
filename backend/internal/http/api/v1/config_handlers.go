package v1

import (
	"errors"
	"net/http"

	"github.com/auto-dns/pihole-cluster-admin/internal/http/transport"
	"github.com/go-chi/chi"
)

func registerConfig(r chi.Router, d Deps) {
	r.Route("/config", func(r chi.Router) {
		r.Get("/", configGet(d))
		r.Patch("/", configPatch(d))
	})
}

func configGet(d Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		result := d.ConfigService.GetClusterConfig(r.Context())

		dto := getConfigResponseDTO{
			Drifted: result.Drifted,
			PerNode: make(map[int64]*configDTO, len(result.PerNode)),
		}
		if result.Consensus != nil {
			c := toConfigDTO(*result.Consensus)
			dto.Consensus = &c
		}
		for id, cfg := range result.PerNode {
			c := toConfigDTO(*cfg)
			dto.PerNode[id] = &c
		}

		transport.WriteJSON(w, http.StatusOK, dto)
	}
}

func configPatch(d Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body patchConfigRequestDTO
		if err := transport.DecodeJSONBody(w, r, &body, 1<<20); err != nil {
			transport.WriteErr(w, err)
			return
		}
		if body.DNS == nil && body.Misc == nil && body.FTL == nil && body.Webserver == nil && body.Resolver == nil {
			transport.WriteBadRequestErr(w, "no fields to update", errors.New("no fields to update"))
			return
		}

		patch := patchRequestToDomain(body)
		results := d.ConfigService.PatchConfig(r.Context(), patch)

		dto := patchConfigResponseDTO{
			Nodes: make(map[int64]patchConfigNodeDTO, len(results)),
		}
		successCount := 0
		for id, nr := range results {
			node := patchConfigNodeDTO{
				Node: piholeNodeRefDTO{
					Id:   nr.PiholeNode.Id,
					Name: nr.PiholeNode.Name,
					Host: nr.PiholeNode.Host,
				},
				Success: nr.Success,
			}
			if nr.Error != nil {
				node.Error = nr.Error.Error()
			}
			if nr.Success {
				successCount++
			}
			dto.Nodes[id] = node
		}

		status := http.StatusOK
		if successCount < len(results) {
			status = http.StatusMultiStatus
		}
		transport.WriteJSON(w, status, dto)
	}
}
