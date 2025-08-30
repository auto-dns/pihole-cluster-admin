package v1

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/auto-dns/pihole-cluster-admin/internal/http/helpers"
	"github.com/auto-dns/pihole-cluster-admin/internal/http/requestctx"
	auth_s "github.com/auto-dns/pihole-cluster-admin/internal/service/auth"
	"github.com/go-chi/chi"
)

// TODO: should login / logout be versioned?

func registerAuthPublic(r chi.Router, d Deps) {
	r.Post("/login", authLogin(d))
	r.Post("/logout", authLogout(d))
}

func registerAuthPrivate(r chi.Router, d Deps) {
	r.Get("/session/user", authGetSessionUser(d))
}

func authLogin(d Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body loginRequestDTO
		if err := helpers.DecodeJSONBody(w, r, &body, 1<<20); err != nil {
			helpers.WriteErr(w, err)
			return
		}

		cmd := auth_s.LoginCommand{
			Username: body.Username,
			Password: body.Password,
		}
		user, sessionId, err := d.AuthService.Login(cmd)
		if err != nil {
			d.Logger.Error().Err(err).Msg("logging in")
			helpers.WriteErr(w, err)
			return
		}

		res := fromDomainUser(user)

		http.SetCookie(w, d.HttpCookieFactory.Make(sessionId))
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(res)
	}
}

func authLogout(d Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(d.HttpCookieFactory.Name())
		if err != nil {
			w.WriteHeader(http.StatusOK)
			return
		}

		_ = d.AuthService.Logout(cookie.Value)

		expired := d.HttpCookieFactory.Make("")
		expired.Expires = time.Now().Add(-1 * time.Hour)
		http.SetCookie(w, expired)
		w.WriteHeader(http.StatusOK)
	}
}

func authGetSessionUser(d Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userId, ok := requestctx.UserID(r.Context())
		if !ok {
			helpers.WriteUnauthorizedErr(w, "unauthorized")
			return
		}

		user, err := d.AuthService.GetUser(userId)
		if err != nil {
			d.Logger.Error().Err(err).Int64("id", userId).Msg("error getting user")
			helpers.WriteErr(w, err)
			return
		}

		d.Logger.Debug().Int64("id", user.Id).Str("username", user.Username).Msg("user fetched from database")

		res := fromDomainUser(user)

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(res)
	}
}
