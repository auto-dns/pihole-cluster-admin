package v1

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"time"

	requestctx "github.com/auto-dns/pihole-cluster-admin/internal/http/context"
	"github.com/auto-dns/pihole-cluster-admin/internal/http/transport"
	authsvc "github.com/auto-dns/pihole-cluster-admin/internal/service/auth"
	"github.com/go-chi/chi"
)

func generateCSRFToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

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
		if err := transport.DecodeJSONBody(w, r, &body, 1<<20); err != nil {
			transport.WriteErr(w, err)
			return
		}

		cmd := authsvc.LoginCommand{
			Username: body.Username,
			Password: body.Password,
		}
		user, sessionId, err := d.AuthService.Login(cmd)
		if err != nil {
			d.Logger.Error().Err(err).Msg("logging in")
			transport.WriteErr(w, err)
			return
		}

		res := fromDomainUser(user)

		csrfToken, err := generateCSRFToken()
		if err != nil {
			d.Logger.Error().Err(err).Msg("generating CSRF token")
			transport.WriteErr(w, err)
			return
		}

		http.SetCookie(w, d.HttpCookieFactory.Make(sessionId))
		http.SetCookie(w, d.HttpCookieFactory.MakeCSRF(csrfToken))
		transport.WriteJSON(w, http.StatusOK, res)
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

		expiredCSRF := d.HttpCookieFactory.MakeCSRF("")
		expiredCSRF.Expires = time.Now().Add(-1 * time.Hour)
		expiredCSRF.MaxAge = -1
		http.SetCookie(w, expiredCSRF)

		w.WriteHeader(http.StatusOK)
	}
}

func authGetSessionUser(d Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userId, ok := requestctx.UserID(r.Context())
		if !ok {
			transport.WriteUnauthorizedErr(w, "unauthorized")
			return
		}

		user, err := d.AuthService.GetUser(userId)
		if err != nil {
			d.Logger.Error().Err(err).Int64("id", userId).Msg("error getting user")
			transport.WriteErr(w, err)
			return
		}

		d.Logger.Debug().Int64("id", user.Id).Str("username", user.Username).Msg("user fetched from database")

		res := fromDomainUser(user)
		transport.WriteJSON(w, http.StatusOK, res)
	}
}
