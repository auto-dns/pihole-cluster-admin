package setup

import (
	"encoding/json"
	"net/http"
	"strings"

	setup_s "github.com/auto-dns/pihole-cluster-admin/internal/service/setup"
	"github.com/auto-dns/pihole-cluster-admin/internal/transport/httpx"
	"github.com/go-chi/chi"
	"github.com/rs/zerolog"
)

type Handler struct {
	service           service
	httpCookieFactory httpCookieFactory
	logger            zerolog.Logger
}

func NewHandler(service service, httpCookieFactory httpCookieFactory, logger zerolog.Logger) *Handler {
	return &Handler{
		service:           service,
		httpCookieFactory: httpCookieFactory,
		logger:            logger,
	}
}

func (h *Handler) RegisterPublic(r chi.Router) {
	r.Get("/initialized", h.getIsInitialized)
	r.Post("/user", h.createUser)
}

func (h *Handler) RegisterPrivate(r chi.Router) {
	r.Get("/status", h.getInitializationStatus)
	r.Patch("/status/pihole", h.updatePiholeInitializationStatus)
}

func (h *Handler) getIsInitialized(w http.ResponseWriter, r *http.Request) {
	initialized, err := h.service.IsInitialized()
	if err != nil {
		h.logger.Error().Err(err).Msg("failed to get app initialization status")
		httpx.WriteJSONError(w, "server error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"initialized": initialized})
}

func (h *Handler) createUser(w http.ResponseWriter, r *http.Request) {
	// Parse request body
	var body setup_s.CreateUserParams
	if err := httpx.DecodeJSONBody(w, r, &body, 1<<20); err != nil {
		h.logger.Error().Err(err).Msg("invalid JSON body")
		httpx.WriteJSONError(w, "invalid JSON body", http.StatusBadRequest)
		return
	}

	// Validate request params
	if strings.TrimSpace(body.Username) == "" {
		h.logger.Error().Msg("empty username in body")
		httpx.WriteJSONError(w, "empty username in body", http.StatusBadRequest)
		return
	}
	if strings.TrimSpace(body.Password) == "" {
		h.logger.Error().Msg("empty password in body")
		httpx.WriteJSONError(w, "empty password in body", http.StatusBadRequest)
		return
	}

	user, sessionId, err := h.service.CreateUser(body)
	if err != nil {
		h.logger.Error().Err(err).Msg("error creating user and session")
		httpx.WriteJSONErrorFromErr(w, err)
		return
	}

	http.SetCookie(w, h.httpCookieFactory.Cookie(sessionId))
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
}

func (h *Handler) getInitializationStatus(w http.ResponseWriter, r *http.Request) {
	initializationStatus, err := h.service.GetInitializationStatus()
	if err != nil {
		h.logger.Error().Err(err).Msg("failed to get app initialization status")
		httpx.WriteJSONErrorFromErr(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(initializationStatus)
}

func (h *Handler) updatePiholeInitializationStatus(w http.ResponseWriter, r *http.Request) {
	// Parse request body
	var body setup_s.UpdatePiholeInitializationStatusParams
	if err := httpx.DecodeJSONBody(w, r, &body, 1<<20); err != nil {
		h.logger.Error().Err(err).Msg("invalid JSON body")
		httpx.WriteJSONError(w, "invalid JSON body", http.StatusBadRequest)
		return
	}
	logger := h.logger.With().Str("new_pihole_status", string(body.Status)).Logger()

	if err := h.service.UpdatePiholeInitializationStatus(body); err != nil {
		logger.Error().Err(err).Msg("setting pihole initialization status")
		httpx.WriteJSONErrorFromErr(w, err)
		return
	}

	logger.Debug().Msg("updated pihole init status in store")
	w.WriteHeader(http.StatusNoContent)
}
