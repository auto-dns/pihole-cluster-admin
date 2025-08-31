package v1

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/auto-dns/pihole-cluster-admin/internal/http/transport"
	setupsvc "github.com/auto-dns/pihole-cluster-admin/internal/service/setup"
	"github.com/go-chi/chi"
)

func registerSetupPublic(r chi.Router, d Deps) {
	r.Get("/setup/initialized", setupGetIsInitialized(d))
	r.Post("/setup/user", setupCreateUser(d))
}

func registerSetupPrivate(r chi.Router, d Deps) {
	r.Get("/setup/status", setupGetInitializationStatus(d))
	r.Patch("/setup/status/pihole", setupUpdatePiholeInitializationStatus(d))
}

func setupGetIsInitialized(d Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		initialized, err := d.SetupService.IsInitialized()
		if err != nil {
			d.Logger.Error().Err(err).Msg("failed to get app initialization status")
			transport.WriteErr(w, err)
			return
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")

		res := isInitializedResponseDTO{
			Initialized: initialized,
		}

		_ = json.NewEncoder(w).Encode(res)
	}
}

func setupCreateUser(d Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Parse request body
		var body createUserRequestDTO
		if err := transport.DecodeJSONBody(w, r, &body, 1<<20); err != nil {
			d.Logger.Error().Err(err).Msg("invalid JSON body")
			transport.WriteErr(w, err)
			return
		}

		// Validate request params
		if strings.TrimSpace(body.Username) == "" {
			d.Logger.Error().Msg("empty username in body")
			transport.WriteBadRequestErr(w, "empty username in body", errors.New("empty username in body"))
			return
		}
		if strings.TrimSpace(body.Password) == "" {
			d.Logger.Error().Msg("empty password in body")
			transport.WriteBadRequestErr(w, "empty password in body", errors.New("empty password in body"))
			return
		}

		cmd := setupsvc.CreateUserCommand{
			Username: body.Username,
			Password: body.Password,
		}
		user, sessionId, err := d.SetupService.CreateUser(r.Context(), cmd)
		if err != nil {
			d.Logger.Error().Err(err).Msg("error creating user and session")
			transport.WriteErr(w, err)
			return
		}

		res := fromDomainUser(user)

		http.SetCookie(w, d.HttpCookieFactory.Make(sessionId))
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(res)
	}
}

func setupGetInitializationStatus(d Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		initializationStatus, err := d.SetupService.GetInitializationStatus()
		if err != nil {
			d.Logger.Error().Err(err).Msg("failed to get app initialization status")
			transport.WriteErr(w, err)
			return
		}

		res := toInitStatusFromDomain(initializationStatus)

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		_ = json.NewEncoder(w).Encode(res)
	}
}

func setupUpdatePiholeInitializationStatus(d Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Parse request body
		var body updatePiholeInitializationStatusRequestDTO
		if err := transport.DecodeJSONBody(w, r, &body, 1<<20); err != nil {
			d.Logger.Error().Err(err).Msg("invalid JSON body")
			transport.WriteBadRequestErr(w, "invalid JSON body", errors.New("invalid JSON body"))
			return
		}
		logger := d.Logger.With().Str("new_pihole_status", string(body.Status)).Logger()

		cmd := toUpdatePiholeInitStatusCommand(body)
		if valid := cmd.Status.IsValid(); !valid {
			transport.WriteBadRequestErr(w, "unsupported \"status\" value", errors.New("unsupported \"status\" value"))
			return
		}

		if err := d.SetupService.UpdatePiholeInitializationStatus(cmd); err != nil {
			logger.Error().Err(err).Msg("setting pihole initialization status")
			transport.WriteErr(w, err)
			return
		}

		logger.Debug().Msg("updated pihole init status in store")
		w.WriteHeader(http.StatusNoContent)
	}
}
