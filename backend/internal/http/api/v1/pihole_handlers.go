package v1

import (
	"errors"
	"net/http"
	"strings"

	"github.com/auto-dns/pihole-cluster-admin/internal/http/transport"
	piholesvc "github.com/auto-dns/pihole-cluster-admin/internal/service/pihole"
	"github.com/go-chi/chi"
)

func registerPihole(r chi.Router, d Deps) {
	r.Route("/pihole", func(r chi.Router) {
		r.Get("/", piholeGetAll(d))                      // GET  /api/v1/pihole
		r.Post("/", piholeAdd(d))                        // POST /api/v1/pihole
		r.Post("/test", piholeTestInstanceConnection(d)) // POST /api/v1/pihole/test
		r.Route("/{id}", func(r chi.Router) {
			r.Patch("/", piholeUpdate(d))                           // PATCH  /api/v1/pihole/{id}
			r.Delete("/", piholeRemove(d))                          // DELETE /api/v1/pihole/{id}
			r.Post("/test", piholeTestExistingConnection(d))        // POST /api/v1/pihole/{id}/test
			r.Post("/rotate-password", piholeRotatePassword(d))    // POST /api/v1/pihole/{id}/rotate-password
		})
	})
}

func piholeGetAll(d Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		piholes, err := d.PiholeService.GetAll()
		if err != nil {
			d.Logger.Error().Err(err).Msg("error getting pihole nodes from database")
			transport.WriteErr(w, err)
			return
		}

		d.Logger.Debug().Int("count", len(piholes)).Msg("fetched pihole nodes from database")

		res := make([]piholeNodeDTO, 0, len(piholes))
		for _, n := range piholes {
			res = append(res, fromDomainPiholeNode(n))
		}

		transport.WriteJSON(w, http.StatusOK, res)
	}
}

func piholeAdd(d Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Parse request body
		var body piholeNodeAddRequestDTO
		if err := transport.DecodeJSONBody(w, r, &body, 1<<20); err != nil {
			d.Logger.Error().Err(err).Msg("invalid JSON body")
			transport.WriteBadRequestErr(w, "invalid JSON body", errors.New("invalid JSON body"))
			return
		}

		// Validate the inputs
		if body.Scheme != "http" && body.Scheme != "https" {
			d.Logger.Error().Msg("scheme must be http or https")
			transport.WriteBadRequestErr(w, "scheme must be http or https", errors.New("scheme must be http or https"))
			return
		}
		if strings.TrimSpace(body.Host) == "" {
			d.Logger.Error().Msg("host must not be empty")
			transport.WriteBadRequestErr(w, "host must not be empty", errors.New("host must not be empty"))
			return
		}
		if body.Port <= 0 || body.Port > 65535 {
			d.Logger.Error().Msg("port must be a valid TCP port")
			transport.WriteBadRequestErr(w, "port must be a valid TCP port", errors.New("port must be a valid TCP port"))
			return
		}
		if strings.TrimSpace(body.Name) == "" {
			d.Logger.Error().Msg("name must not be empty")
			transport.WriteBadRequestErr(w, "name must not be empty", errors.New("name must not be empty"))
			return
		}
		if strings.TrimSpace(body.Password) == "" {
			d.Logger.Error().Msg("password must not be empty")
			transport.WriteBadRequestErr(w, "password must not be empty", errors.New("password must not be empty"))
			return
		}

		addParams := piholesvc.AddNodeCommand{
			Scheme:      body.Scheme,
			Host:        body.Host,
			Port:        body.Port,
			Name:        body.Name,
			Description: body.Description,
			Password:    body.Password,
		}

		insertedNode, err := d.PiholeService.Add(r.Context(), addParams)
		if err != nil {
			d.Logger.Error().Err(err).Str("host", addParams.Host).Int("port", addParams.Port).Msg("adding node")
			transport.WriteErr(w, err)
			return
		}
		d.Logger.Debug().Int64("id", insertedNode.Id).Str("scheme", insertedNode.Scheme).Str("host", insertedNode.Host).Int("port", insertedNode.Port).Str("name", insertedNode.Name).Time("created_at", insertedNode.CreatedAt).Time("updated_at", insertedNode.UpdatedAt).Msg("added pihole node")

		res := fromDomainPiholeNode(insertedNode)
		transport.WriteJSON(w, http.StatusCreated, res)
	}
}

func piholeUpdate(d Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Parse request body
		var body piholeNodeUpdateRequestDTO
		if err := transport.DecodeJSONBody(w, r, &body, 1<<20); err != nil {
			d.Logger.Error().Err(err).Msg("invalid JSON body")
			transport.WriteErr(w, err)
			return
		}

		// Validate request
		id, ok := ParseInt64Param(r, "id", 1)
		if !ok {
			d.Logger.Error().Msg("invalid id path parameter")
			transport.WriteBadRequestErr(w, "error processing id path parameter", errors.New("error processing id path parameter"))
			return
		}
		// Validate at least one update field set
		if body.Scheme == nil && body.Host == nil && body.Port == nil && body.Name == nil && body.Description == nil && body.Password == nil {
			d.Logger.Error().Msg("must provide at least one field to update")
			transport.WriteBadRequestErr(w, "must provide at least one field to update", errors.New("must provide at least one field to update"))
			return
		}
		// Validate content
		if body.Scheme != nil && *body.Scheme != "http" && *body.Scheme != "https" {
			d.Logger.Error().Msg("scheme must be http or https")
			transport.WriteBadRequestErr(w, "scheme must be http or https", errors.New("scheme must be http or https"))
			return
		}
		if body.Host != nil && strings.TrimSpace(*body.Host) == "" {
			d.Logger.Error().Msg("host must not be empty")
			transport.WriteBadRequestErr(w, "host must not be empty", errors.New("host must not be empty"))
			return
		}
		if body.Port != nil && (*body.Port <= 0 || *body.Port > 65535) {
			d.Logger.Error().Msg("port must be a valid TCP port")
			transport.WriteBadRequestErr(w, "port must be a valid TCP port", errors.New("port must be a valid TCP port"))
			return
		}
		if body.Name != nil && strings.TrimSpace(*body.Name) == "" {
			d.Logger.Error().Msg("name must not be empty")
			transport.WriteBadRequestErr(w, "name must not be empty", errors.New("name must not be empty"))
			return
		}
		if body.Password != nil && strings.TrimSpace(*body.Password) == "" {
			d.Logger.Error().Msg("password must not be empty")
			transport.WriteBadRequestErr(w, "password must not be empty", errors.New("password must not be empty"))
			return
		}

		updateParams := piholesvc.UpdateNodeCommand{
			Scheme:      body.Scheme,
			Host:        body.Host,
			Port:        body.Port,
			Name:        body.Name,
			Description: body.Description,
			Password:    body.Password,
		}

		updatedNode, err := d.PiholeService.Update(r.Context(), id, updateParams)
		safe := func(p *string) string {
			if p == nil {
				return ""
			}
			return *p
		}
		safeInt := func(p *int) int {
			if p == nil {
				return 0
			}
			return *p
		}
		if err != nil {
			d.Logger.Error().Err(err).Str("host", safe(updateParams.Host)).Int("port", safeInt(updateParams.Port)).Msg("updating node")
			transport.WriteErr(w, err)
			return
		}

		d.Logger.Debug().Int64("id", updatedNode.Id).Str("scheme", updatedNode.Scheme).Str("host", updatedNode.Host).Int("port", updatedNode.Port).Time("created_at", updatedNode.CreatedAt).Str("name", updatedNode.Name).Time("updated_at", updatedNode.UpdatedAt).Msg("updated pihole node")

		res := fromDomainPiholeNode(updatedNode)
		transport.WriteJSON(w, http.StatusOK, res)
	}
}

func piholeRemove(d Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, ok := ParseInt64Param(r, "id", 1)
		if !ok {
			d.Logger.Error().Msg("invalid id path parameter")
			transport.WriteBadRequestErr(w, "invalid id (<= 0)", nil)
			return
		}

		found, err := d.PiholeService.Remove(r.Context(), id)
		if err != nil {
			d.Logger.Error().Err(err).Int64("id", id).Msg("error removing pihole node")
			transport.WriteErr(w, err)
			return
		}

		if !found {
			d.Logger.Error().Int64("id", id).Msg("pihole not found")
			w.WriteHeader(http.StatusNotFound)
			return
		}

		d.Logger.Debug().Int64("id", id).Msg("pihole removed")
		w.WriteHeader(http.StatusNoContent)
	}
}

func piholeTestInstanceConnection(d Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Used to test a pihole instance that hasn't been turned into a cluster yet
		var body piholeTestInstanceConnectionRequestDTO
		if err := transport.DecodeJSONBody(w, r, &body, 1<<20); err != nil {
			transport.WriteErr(w, err)
			return
		}

		// Validate scheme
		body.Scheme = strings.ToLower(strings.TrimSpace(body.Scheme))
		switch body.Scheme {
		case "http", "https":
		default:
			transport.WriteBadRequestErr(w, "scheme must be http or https", errors.New("scheme must be http or https"))
			return
		}
		// Validate host
		body.Host = strings.TrimSpace(body.Host)
		if body.Host == "" {
			transport.WriteBadRequestErr(w, "host is required", errors.New("host is required"))
			return
		}
		// Validate port
		if body.Port == 0 {
			if body.Scheme == "https" {
				body.Port = 443
			} else {
				body.Port = 80
			}
		}
		if body.Port < 1 || body.Port > 65535 {
			transport.WriteBadRequestErr(w, "invalid port", errors.New("invalid port"))
			return
		}
		// Validate password
		if body.Password == "" {
			transport.WriteBadRequestErr(w, "password is required", errors.New("password is required"))
			return
		}

		cmd := piholesvc.TestInstanceConnectionCommand{
			Scheme:   body.Scheme,
			Host:     body.Host,
			Port:     body.Port,
			Password: body.Password,
		}

		if err := d.PiholeService.TestInstanceConnection(r.Context(), cmd); err != nil {
			transport.WriteErr(w, err)
			return
		}

		d.Logger.Debug().Str("scheme", body.Scheme).Str("host", body.Host).Int("port", body.Port).Msg("successfully logged in with pihole instance")

		w.WriteHeader(http.StatusNoContent)
	}
}

func piholeRotatePassword(d Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, ok := ParseInt64Param(r, "id", 1)
		if !ok {
			transport.WriteBadRequestErr(w, "invalid id path parameter", errors.New("invalid id path parameter"))
			return
		}

		var body piholeRotatePasswordRequestDTO
		if err := transport.DecodeJSONBody(w, r, &body, 1<<20); err != nil {
			d.Logger.Error().Err(err).Msg("invalid JSON body")
			transport.WriteErr(w, err)
			return
		}

		if strings.TrimSpace(body.NewPassword) == "" {
			transport.WriteBadRequestErr(w, "newPassword must not be empty", errors.New("newPassword must not be empty"))
			return
		}

		cmd := piholesvc.RotatePasswordCommand{NewPassword: body.NewPassword}
		if err := d.PiholeService.RotatePassword(r.Context(), id, cmd); err != nil {
			d.Logger.Error().Err(err).Int64("id", id).Msg("rotating Pi-hole password")
			transport.WriteErr(w, err)
			return
		}

		d.Logger.Info().Int64("id", id).Msg("rotated Pi-hole node password")
		w.WriteHeader(http.StatusNoContent)
	}
}

func piholeTestExistingConnection(d Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, ok := ParseInt64Param(r, "id", 0)
		if !ok {
			transport.WriteBadRequestErr(w, "invalid id", nil)
			return
		}

		var body piholeTestExistingConnectionRequestDTO
		if err := transport.DecodeJSONBody(w, r, &body, 1<<20); err != nil {
			transport.WriteErr(w, err)
			return
		}

		cmd := piholesvc.TestExistingConnectionCommand{
			Scheme:   body.Scheme,
			Host:     body.Host,
			Port:     body.Port,
			Password: body.Password,
		}

		if err := d.PiholeService.TestExistingConnection(r.Context(), id, cmd); err != nil {
			transport.WriteErr(w, err)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}
