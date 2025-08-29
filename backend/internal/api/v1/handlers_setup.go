package v1

import (
	"encoding/json"
	"net/http"
	"strings"

	setup_s "github.com/auto-dns/pihole-cluster-admin/internal/service/setup"
	"github.com/auto-dns/pihole-cluster-admin/internal/transport/httpx"
	"github.com/go-chi/chi"
)

func registerSetupPublic(r chi.Router, d Deps) {
	r.Get("/initialized", setupGetIsInitialized(d))
	r.Post("/user", setupCreateUser(d))
}

func registerSetupPrivate(r chi.Router, d Deps) {
	r.Get("/status", setupGetInitializationStatus(d))
	r.Patch("/status/pihole", setupUpdatePiholeInitializationStatus(d))
}

func setupGetIsInitialized(d Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		initialized, err := d.SetupService.IsInitialized()
		if err != nil {
			d.Logger.Error().Err(err).Msg("failed to get app initialization status")
			httpx.WriteJSONError(w, "server error", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")

		res := isInitializedResponseDTO{
			Initialized: initialized,
		}

		json.NewEncoder(w).Encode(res)
	}
}

func setupCreateUser(d Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Parse request body
		var body createUserRequestDTO
		if err := httpx.DecodeJSONBody(w, r, &body, 1<<20); err != nil {
			d.Logger.Error().Err(err).Msg("invalid JSON body")
			httpx.WriteJSONError(w, "invalid JSON body", http.StatusBadRequest)
			return
		}

		// Validate request params
		if strings.TrimSpace(body.Username) == "" {
			d.Logger.Error().Msg("empty username in body")
			httpx.WriteJSONError(w, "empty username in body", http.StatusBadRequest)
			return
		}
		if strings.TrimSpace(body.Password) == "" {
			d.Logger.Error().Msg("empty password in body")
			httpx.WriteJSONError(w, "empty password in body", http.StatusBadRequest)
			return
		}

		cmd := setup_s.CreateUserCommand{
			Username: body.Username,
			Password: body.Password,
		}
		user, sessionId, err := d.SetupService.CreateUser(r.Context(), cmd)
		if err != nil {
			d.Logger.Error().Err(err).Msg("error creating user and session")
			httpx.WriteJSONErrorFromErr(w, err)
			return
		}

		res := FromDomainUser(user)

		http.SetCookie(w, d.HttpCookieFactory.Cookie(sessionId))
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(res)
	}
}

func setupGetInitializationStatus(d Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		initializationStatus, err := d.SetupService.GetInitializationStatus()
		if err != nil {
			d.Logger.Error().Err(err).Msg("failed to get app initialization status")
			httpx.WriteJSONErrorFromErr(w, err)
			return
		}

		res := ToInitStatusFromDomain(*initializationStatus)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(res)
	}
}

func setupUpdatePiholeInitializationStatus(d Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Parse request body
		var body updatePiholeInitializationStatusRequestDTO
		if err := httpx.DecodeJSONBody(w, r, &body, 1<<20); err != nil {
			d.Logger.Error().Err(err).Msg("invalid JSON body")
			httpx.WriteJSONError(w, "invalid JSON body", http.StatusBadRequest)
			return
		}
		logger := d.Logger.With().Str("new_pihole_status", string(body.Status)).Logger()

		cmd := ToUpdatePiholeInitStatusCommand(body)
		if valid := cmd.Status.IsValid(); !valid {
			err := httpx.NewHttpError(httpx.ErrValidation, "unsupported \"status\" value")
			d.Logger.Error().Err(err).Msg("invalid JSON body")
			httpx.WriteJSONErrorFromErr(w, err)
			return
		}

		if err := d.SetupService.UpdatePiholeInitializationStatus(cmd); err != nil {
			logger.Error().Err(err).Msg("setting pihole initialization status")
			httpx.WriteJSONErrorFromErr(w, err)
			return
		}

		logger.Debug().Msg("updated pihole init status in store")
		w.WriteHeader(http.StatusNoContent)
	}
}
