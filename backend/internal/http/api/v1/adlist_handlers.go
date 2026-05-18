package v1

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/auto-dns/pihole-cluster-admin/internal/domain"
	"github.com/auto-dns/pihole-cluster-admin/internal/http/transport"
	"github.com/go-chi/chi"
)

func registerAdlists(r chi.Router, d Deps) {
	r.Route("/adlists", func(r chi.Router) {
		r.Get("/", adlistList(d))
		r.Post("/", adlistAdd(d))
		r.Put("/{id}", adlistUpdate(d))
		r.Delete("/{id}", adlistRemove(d))
	})
	r.Post("/gravity/rebuild", adlistRebuildGravity(d))
}

func adlistList(d Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		results := d.AdlistService.List(r.Context())
		transport.WriteJSON(w, http.StatusOK, listAdlistsResponseFromDomain(results))
	}
}

func adlistAdd(d Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body addAdlistRequestDTO
		if err := transport.DecodeJSONBody(w, r, &body, 1<<20); err != nil {
			d.Logger.Error().Err(err).Msg("invalid JSON body")
			transport.WriteErr(w, err)
			return
		}

		if body.Address == "" {
			transport.WriteBadRequestErr(w, "\"address\" is required", errors.New("\"address\" is required"))
			return
		}

		adlistType, ok := parseAdlistType(body.Type)
		if !ok {
			transport.WriteBadRequestErr(w, "\"type\" must be \"block\" or \"allow\"", errors.New("invalid type"))
			return
		}

		enabled := true
		if body.Enabled != nil {
			enabled = *body.Enabled
		}

		cmd := domain.AddAdlistCommand{
			Address: body.Address,
			Type:    adlistType,
			Comment: body.Comment,
			Groups:  body.Groups,
			Enabled: enabled,
		}

		results := d.AdlistService.Add(r.Context(), cmd)
		transport.WriteJSON(w, http.StatusOK, addAdlistResponseFromDomain(results))
	}
}

func adlistUpdate(d Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := chi.URLParam(r, "id")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil || id <= 0 {
			transport.WriteBadRequestErr(w, "invalid adlist id", errors.New("invalid adlist id"))
			return
		}

		var body updateAdlistRequestDTO
		if err := transport.DecodeJSONBody(w, r, &body, 1<<20); err != nil {
			d.Logger.Error().Err(err).Msg("invalid JSON body")
			transport.WriteErr(w, err)
			return
		}

		cmd := domain.UpdateAdlistCommand{
			Id:      id,
			Enabled: body.Enabled,
			Comment: body.Comment,
			Groups:  body.Groups,
		}

		results := d.AdlistService.Update(r.Context(), cmd)
		transport.WriteJSON(w, http.StatusOK, addAdlistResponseFromDomain(results))
	}
}

func adlistRemove(d Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := chi.URLParam(r, "id")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil || id <= 0 {
			transport.WriteBadRequestErr(w, "invalid adlist id", errors.New("invalid adlist id"))
			return
		}

		cmd := domain.RemoveAdlistCommand{Id: id}
		results := d.AdlistService.Remove(r.Context(), cmd)
		transport.WriteJSON(w, http.StatusOK, removeAdlistResponseFromDomain(results))
	}
}

func adlistRebuildGravity(d Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		results := d.AdlistService.RebuildGravity(r.Context())
		transport.WriteJSON(w, http.StatusOK, gravityRebuildResponseFromDomain(results))
	}
}

func parseAdlistType(s string) (domain.AdlistType, bool) {
	switch domain.AdlistType(s) {
	case domain.AdlistTypeBlock, domain.AdlistTypeAllow:
		return domain.AdlistType(s), true
	default:
		return "", false
	}
}
