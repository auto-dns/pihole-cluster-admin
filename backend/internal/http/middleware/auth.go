package middleware

import (
	"net/http"

	"github.com/auto-dns/pihole-cluster-admin/internal/config"
	requestctx "github.com/auto-dns/pihole-cluster-admin/internal/http/context"
	"github.com/rs/zerolog"
)

type AuthDeps struct {
	Sessions sessionManager
	Cfg      config.SessionConfig
	Logger   zerolog.Logger
}

func RequireAuth(d AuthDeps) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie(d.Cfg.CookieName)
			if err != nil {
				d.Logger.Warn().Str("cookie_name", d.Cfg.CookieName).Msg("error accessing cookie")
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			userId, ok, err := d.Sessions.GetUserId(cookie.Value)
			if err != nil {
				d.Logger.Warn().Str("session_id", truncateSessionID(cookie.Value)).Msg("error retrieving session")
				http.Error(w, "internal server error", http.StatusInternalServerError)
				return
			} else if !ok {
				d.Logger.Warn().Str("session_id", truncateSessionID(cookie.Value)).Msg("session not found")
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			// Pass username to request context
			ctx := requestctx.WithUserID(r.Context(), userId)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// TODO: deduplicate this from internal/sessions/session_manager.go
func truncateSessionID(id string) string {
	if len(id) > 8 {
		return id[:8]
	}
	return id
}
