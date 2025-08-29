package v1

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	user_s "github.com/auto-dns/pihole-cluster-admin/internal/service/user"
	"github.com/auto-dns/pihole-cluster-admin/internal/sessions"
	"github.com/auto-dns/pihole-cluster-admin/internal/transport/httpx"
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
		if err := httpx.DecodeJSONBody(w, r, &body, 1<<20); err != nil {
			d.Logger.Error().Err(err).Msg("invalid JSON body")
			httpx.WriteJSONError(w, "invalid JSON body", http.StatusBadRequest)
			return
		}

		// Validate request
		idString := chi.URLParam(r, "id")
		id, err := strconv.ParseInt(idString, 10, 64)
		if err != nil {
			d.Logger.Error().Err(err).Msg("error converting path parameter id to int64")
			httpx.WriteJSONError(w, "error processing id path parameter", http.StatusBadRequest)
			return
		}
		if id <= 0 {
			d.Logger.Error().Msg("invalid id (<= 0)")
			httpx.WriteJSONError(w, "invalid id (<= 0)", http.StatusBadRequest)
			return
		}

		// Must be current user
		currentUserId, ok := r.Context().Value(sessions.UserIdContextKey).(int64)
		if !ok {
			d.Logger.Error().Err(err).Msg("error getting current user id from context")
			httpx.WriteJSONError(w, "internal server error", http.StatusInternalServerError)
			return
		}

		if id != currentUserId {
			d.Logger.Error().Err(err).Int64("current_user_id", currentUserId).Int64("id", id).Msg("user tried to upate user id other than own")
			httpx.WriteJSONError(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		// Validate at least one update field set
		if body.Username == nil {
			d.Logger.Error().Msg("no fields provided")
			httpx.WriteJSONError(w, "must provide at least one field to update", http.StatusBadRequest)
			return
		} else if strings.TrimSpace(*body.Username) == "" {
			d.Logger.Error().Msg("username empty")
			httpx.WriteJSONError(w, "username must not be empty", http.StatusBadRequest)
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
			httpx.WriteJSONErrorFromErr(w, err)
			return
		}

		res := fromDomainUser(updatedUser)

		d.Logger.Debug().Int64("id", updatedUser.Id).Str("username", updatedUser.Username).Msg("updated user")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(res)
	}
}

func userUpdatePassword(d Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Parse request body
		var body updatePasswordRequestDTO
		if err := httpx.DecodeJSONBody(w, r, &body, 1<<20); err != nil {
			d.Logger.Error().Err(err).Msg("invalid JSON body")
			httpx.WriteJSONError(w, "invalid JSON body", http.StatusBadRequest)
			return
		}

		// Validate request
		idString := chi.URLParam(r, "id")
		id, err := strconv.ParseInt(idString, 10, 64)
		if err != nil {
			d.Logger.Error().Err(err).Msg("error converting path parameter id to int64")
			httpx.WriteJSONError(w, "error processing id path parameter", http.StatusBadRequest)
			return
		}
		if id <= 0 {
			d.Logger.Error().Msg("invalid id (<= 0)")
			httpx.WriteJSONError(w, "invalid id (<= 0)", http.StatusBadRequest)
			return
		}

		// Validate must be current user
		currentUserId, ok := r.Context().Value(sessions.UserIdContextKey).(int64)
		if !ok {
			d.Logger.Error().Err(err).Msg("error getting current user id from context")
			httpx.WriteJSONError(w, "internal server error", http.StatusInternalServerError)
			return
		}

		if id != currentUserId {
			d.Logger.Error().Err(err).Int64("current_user_id", currentUserId).Int64("id", id).Msg("user tried to upate user id other than own")
			httpx.WriteJSONError(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		// Basic validation
		if strings.TrimSpace(body.NewPassword) == "" {
			d.Logger.Error().Msg("new password empty")
			httpx.WriteJSONError(w, "new password must not be empty", http.StatusBadRequest)
			return
		}
		if len(strings.TrimSpace(body.NewPassword)) < 8 {
			d.Logger.Error().Msg("new password less than 8 characters")
			httpx.WriteJSONError(w, "new password must be 8 or more characters", http.StatusBadRequest)
			return
		}

		cmd := user_s.UpdatePasswordCommand{
			CurrentPassword: body.CurrentPassword,
			NewPassword:     body.NewPassword,
		}
		updatedUser, err := d.UserService.UpdatePassword(id, cmd)
		if err != nil {
			d.Logger.Error().Err(err).Int64("id", id).Msg("updating password")
			httpx.WriteJSONErrorFromErr(w, err)
			return
		}

		d.Logger.Debug().Int64("id", updatedUser.Id).Msg("updated password")

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNoContent)
	}
}
