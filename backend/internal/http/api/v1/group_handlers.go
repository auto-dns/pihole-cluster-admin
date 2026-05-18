package v1

import (
	"errors"
	"net/http"

	"github.com/auto-dns/pihole-cluster-admin/internal/domain"
	"github.com/auto-dns/pihole-cluster-admin/internal/http/transport"
	"github.com/go-chi/chi"
)

func registerGroups(r chi.Router, d Deps) {
	r.Route("/groups", func(r chi.Router) {
		r.Get("/", groupList(d))
		r.Post("/", groupAdd(d))
		r.Put("/{name}", groupUpdate(d))
		r.Delete("/{name}", groupRemove(d))
	})
}

func groupList(d Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		results := d.GroupService.List(r.Context())
		transport.WriteJSON(w, http.StatusOK, listGroupsResponseFromDomain(results))
	}
}

func groupAdd(d Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body addGroupRequestDTO
		if err := transport.DecodeJSONBody(w, r, &body, 1<<20); err != nil {
			d.Logger.Error().Err(err).Msg("invalid JSON body")
			transport.WriteErr(w, err)
			return
		}

		if body.Name == "" {
			transport.WriteBadRequestErr(w, "\"name\" is required", errors.New("\"name\" is required"))
			return
		}

		enabled := true
		if body.Enabled != nil {
			enabled = *body.Enabled
		}

		cmd := domain.AddGroupCommand{
			Name:        body.Name,
			Description: body.Description,
			Enabled:     enabled,
		}

		results := d.GroupService.Add(r.Context(), cmd)
		transport.WriteJSON(w, http.StatusOK, groupMutateResponseFromDomain(results))
	}
}

func groupUpdate(d Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		name := chi.URLParam(r, "name")
		if name == "" {
			transport.WriteBadRequestErr(w, "group name is required", errors.New("group name is required"))
			return
		}

		var body updateGroupRequestDTO
		if err := transport.DecodeJSONBody(w, r, &body, 1<<20); err != nil {
			d.Logger.Error().Err(err).Msg("invalid JSON body")
			transport.WriteErr(w, err)
			return
		}

		cmd := domain.UpdateGroupCommand{
			Name:        name,
			Description: body.Description,
			Enabled:     body.Enabled,
		}

		results := d.GroupService.Update(r.Context(), cmd)
		transport.WriteJSON(w, http.StatusOK, groupMutateResponseFromDomain(results))
	}
}

func groupRemove(d Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		name := chi.URLParam(r, "name")
		if name == "" {
			transport.WriteBadRequestErr(w, "group name is required", errors.New("group name is required"))
			return
		}

		cmd := domain.RemoveGroupCommand{Name: name}
		results := d.GroupService.Remove(r.Context(), cmd)
		transport.WriteJSON(w, http.StatusOK, removeGroupResponseFromDomain(results))
	}
}
