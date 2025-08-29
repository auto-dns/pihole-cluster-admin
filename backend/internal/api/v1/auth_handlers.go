package v1

import (
	"encoding/json"
	"net/http"
	"time"

	auth_s "github.com/auto-dns/pihole-cluster-admin/internal/service/auth"
	"github.com/auto-dns/pihole-cluster-admin/internal/sessions"
	"github.com/auto-dns/pihole-cluster-admin/internal/transport/httpx"
	"github.com/go-chi/chi"
)

func registerAuthPublic(r chi.Router, d Deps) {
	r.Post("/auth/login", authLogin(d))
	r.Post("/auth/logout", authLogout(d))
}

func registerAuthPrivate(r chi.Router, d Deps) {
	r.Get("/auth/session/user", authGetSessionUser(d))
}

func authLogin(d Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body auth_s.LoginCommand
		if err := httpx.DecodeJSONBody(w, r, &body, 1<<20); err != nil {
			httpx.WriteJSONError(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		user, sessionId, err := d.AuthService.Login(body)
		if err != nil {
			d.Logger.Error().Err(err).Msg("logging in")
			httpx.WriteJSONErrorFromErr(w, err)
			return
		}

		res := fromDomainUser(user)

		http.SetCookie(w, d.HttpCookieFactory.Cookie(sessionId))
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(res)
	}
}

func authLogout(d Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(d.HttpCookieFactory.CookieName())
		if err != nil {
			w.WriteHeader(http.StatusOK)
			return
		}

		_ = d.AuthService.Logout(cookie.Value)

		expired := d.HttpCookieFactory.Cookie("")
		expired.Expires = time.Now().Add(-1 * time.Hour)
		http.SetCookie(w, expired)
		w.WriteHeader(http.StatusOK)
	}
}

func authGetSessionUser(d Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userId, ok := r.Context().Value(sessions.UserIdContextKey).(int64)
		if !ok {
			httpx.WriteJSONError(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		user, err := d.AuthService.GetUser(userId)
		if err != nil {
			d.Logger.Error().Err(err).Int64("id", userId).Msg("error getting user")
			httpx.WriteJSONErrorFromErr(w, err)
			return
		}

		d.Logger.Debug().Int64("id", user.Id).Str("username", user.Username).Msg("user fetched from database")

		res := fromDomainUser(user)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(res)
	}
}
