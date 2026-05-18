package v1

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/auto-dns/pihole-cluster-admin/internal/domain"
	"github.com/auto-dns/pihole-cluster-admin/internal/http/transport"
	"github.com/go-chi/chi"
)

func registerPiholeClients(r chi.Router, d Deps) {
	r.Route("/clients", func(r chi.Router) {
		r.Get("/", clientList(d))
		r.Put("/{id}", clientUpdate(d))
		r.Delete("/{id}", clientRemove(d))
	})
}

func clientList(d Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		results := d.PiholeClientService.List(r.Context())
		transport.WriteJSON(w, http.StatusOK, listClientsResponseFromDomain(results))
	}
}

func clientUpdate(d Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := chi.URLParam(r, "id")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil || id <= 0 {
			transport.WriteBadRequestErr(w, "invalid client id", errors.New("invalid client id"))
			return
		}

		var body updateClientRequestDTO
		if err := transport.DecodeJSONBody(w, r, &body, 1<<20); err != nil {
			d.Logger.Error().Err(err).Msg("invalid JSON body")
			transport.WriteErr(w, err)
			return
		}

		groups := body.Groups
		if groups == nil {
			groups = []int{}
		}

		cmd := domain.UpdatePiholeClientCommand{
			Id:      id,
			Groups:  groups,
			Comment: body.Comment,
		}

		results := d.PiholeClientService.Update(r.Context(), cmd)
		transport.WriteJSON(w, http.StatusOK, clientMutateResponseFromDomain(results))
	}
}

func clientRemove(d Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := chi.URLParam(r, "id")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil || id <= 0 {
			transport.WriteBadRequestErr(w, "invalid client id", errors.New("invalid client id"))
			return
		}

		cmd := domain.RemovePiholeClientCommand{Id: id}
		results := d.PiholeClientService.Remove(r.Context(), cmd)
		transport.WriteJSON(w, http.StatusOK, removeClientResponseFromDomain(results))
	}
}
