package v1

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/auto-dns/pihole-cluster-admin/internal/errs"
	requestctx "github.com/auto-dns/pihole-cluster-admin/internal/http/context"
	"github.com/auto-dns/pihole-cluster-admin/internal/http/transport"
	user_s "github.com/auto-dns/pihole-cluster-admin/internal/service/user"
	"github.com/go-chi/chi"
)

func registerUser(r chi.Router, d Deps) {
	r.Patch("/user/{id}", userPatch(d))
	r.Post("/user/{id}/password", userUpdatePassword(d))
}

func userPatch(d Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Parse request body
		var body patchUserRequestDTO
		if err := transport.DecodeJSONBody(w, r, &body, 1<<20); err != nil {
			d.Logger.Error().Err(err).Msg("invalid JSON body")
			transport.WriteBadRequestErr(w, "invalid JSON body", errors.New("invalid JSON body"))
			return
		}

		// Validate request
		idString := chi.URLParam(r, "id")
		id, err := strconv.ParseInt(idString, 10, 64)
		if err != nil {
			d.Logger.Error().Err(err).Msg("error converting path parameter id to int64")
			transport.WriteBadRequestErr(w, "error processing id path parameter", errors.New("error processing id path parameter"))
			return
		}
		if id <= 0 {
			d.Logger.Error().Msg("invalid id (<= 0)")
			transport.WriteBadRequestErr(w, "invalid id (<= 0)", errors.New("invalid id (<= 0)"))
			return
		}

		// Must be current user
		currentUserId, ok := requestctx.UserID(r.Context())
		if !ok {
			d.Logger.Error().Err(err).Msg("error getting current user id from context")
			transport.WriteErr(w, errs.New(errs.KindUnknown, "internal server error", errors.New("error getting current user id from context")))
			return
		}

		if id != currentUserId {
			d.Logger.Error().Err(err).Int64("current_user_id", currentUserId).Int64("id", id).Msg("user tried to upate user id other than own")
			transport.WriteUnauthorizedErr(w, "unauthorized")
			return
		}

		// Validate at least one update field set
		if body.Username == nil {
			d.Logger.Error().Msg("no fields provided")
			transport.WriteBadRequestErr(w, "must provide at least one field to update", errors.New("must provide at least one field to update"))
			return
		} else if strings.TrimSpace(*body.Username) == "" {
			d.Logger.Error().Msg("username empty")
			transport.WriteBadRequestErr(w, "username must not be empty", errors.New("username must not be empty"))
			return
		}

		cmd := user_s.PatchUserCommand{
			Username: body.Username,
		}
		updatedUser, err := d.UserService.Patch(id, cmd)

		safe := func(p *string) string {
			if p == nil {
				return ""
			}
			return *p
		}
		if err != nil {
			d.Logger.Error().Err(err).Str("username", safe(body.Username)).Msg("error updating user")
			transport.WriteErr(w, err)
			return
		}

		res := fromDomainUser(updatedUser)

		d.Logger.Debug().Int64("id", updatedUser.Id).Str("username", updatedUser.Username).Msg("updated user")
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(res)
	}
}

func userUpdatePassword(d Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Parse request body
		var body updatePasswordRequestDTO
		if err := transport.DecodeJSONBody(w, r, &body, 1<<20); err != nil {
			d.Logger.Error().Err(err).Msg("invalid JSON body")
			transport.WriteBadRequestErr(w, "invalid JSON body", errors.New("invalid JSON body"))
			return
		}

		// Validate request
		idString := chi.URLParam(r, "id")
		id, err := strconv.ParseInt(idString, 10, 64)
		if err != nil {
			d.Logger.Error().Err(err).Msg("error converting path parameter id to int64")
			transport.WriteBadRequestErr(w, "error processing id path parameter", errors.New("error processing id path parameter"))
			return
		}
		if id <= 0 {
			d.Logger.Error().Msg("invalid id (<= 0)")
			transport.WriteBadRequestErr(w, "invalid id (<= 0)", errors.New("invalid id (<= 0)"))
			return
		}

		// Validate must be current user
		currentUserId, ok := requestctx.UserID(r.Context())
		if !ok {
			d.Logger.Error().Err(err).Msg("error getting current user id from context")
			transport.WriteErr(w, errs.New(errs.KindUnknown, "internal server error", errors.New("error getting current user id from context")))
			return
		}

		if id != currentUserId {
			d.Logger.Error().Err(err).Int64("current_user_id", currentUserId).Int64("id", id).Msg("user tried to upate user id other than own")
			transport.WriteUnauthorizedErr(w, "unauthorized")
			return
		}

		// Basic validation
		if strings.TrimSpace(body.NewPassword) == "" {
			d.Logger.Error().Msg("new password empty")
			transport.WriteBadRequestErr(w, "new password must not be empty", errors.New("new password must not be empty"))
			return
		}
		if len(strings.TrimSpace(body.NewPassword)) < 8 {
			d.Logger.Error().Msg("new password less than 8 characters")
			transport.WriteBadRequestErr(w, "new password must be 8 or more characters", errors.New("new password must be 8 or more characters"))
			return
		}

		cmd := user_s.UpdatePasswordCommand{
			CurrentPassword: body.CurrentPassword,
			NewPassword:     body.NewPassword,
		}
		updatedUser, err := d.UserService.UpdatePassword(id, cmd)
		if err != nil {
			d.Logger.Error().Err(err).Int64("id", id).Msg("updating password")
			transport.WriteErr(w, err)
			return
		}

		d.Logger.Debug().Int64("id", updatedUser.Id).Msg("updated password")

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusNoContent)
	}
}
