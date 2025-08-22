package auth

import (
	"encoding/json"
	"net/http"
	"time"

	auth_s "github.com/auto-dns/pihole-cluster-admin/internal/service/auth"
	"github.com/auto-dns/pihole-cluster-admin/internal/sessions"
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
	r.Post("/login", h.login)
	r.Post("/logout", h.logout)
}

func (h *Handler) RegisterPrivate(r chi.Router) {
	r.Get("/session/user", h.getSessionUser)
}

func (h *Handler) login(w http.ResponseWriter, r *http.Request) {
	var body auth_s.LoginParams
	if err := httpx.DecodeJSONBody(w, r, &body, 1<<20); err != nil {
		httpx.WriteJSONError(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	user, sessionId, err := h.service.Login(body)
	if err != nil {
		h.logger.Error().Err(err).Msg("logging in")
		httpx.WriteJSONErrorFromErr(w, err)
		return
	}

	http.SetCookie(w, h.httpCookieFactory.Cookie(sessionId))
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(user)
}

func (h *Handler) logout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(h.httpCookieFactory.CookieName())
	if err != nil {
		w.WriteHeader(http.StatusOK)
		return
	}

	_ = h.service.Logout(cookie.Value)

	expired := h.httpCookieFactory.Cookie("")
	expired.Expires = time.Now().Add(-1 * time.Hour)
	http.SetCookie(w, expired)
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) getSessionUser(w http.ResponseWriter, r *http.Request) {
	userId, ok := r.Context().Value(sessions.UserIdContextKey).(int64)
	if !ok {
		httpx.WriteJSONError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	user, err := h.service.GetUser(userId)
	if err != nil {
		h.logger.Error().Err(err).Int64("id", userId).Str("username", user.Username).Msg("error getting user")
		httpx.WriteJSONErrorFromErr(w, err)
		return
	}

	h.logger.Debug().Int64("id", user.Id).Str("username", user.Username).Msg("user fetched from database")

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(user)
}
