package middleware

import (
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/middleware"
	"github.com/rs/zerolog"
)

func RequestLogger(l zerolog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			requestId := middleware.GetReqID(r.Context())
			if requestId == "" {
				requestId = "none"
			}
			clientIp := realIP(r)
			reqLog := l.With().
				Str("request_id", requestId).
				Str("method", r.Method).
				Str("path", r.URL.Path).
				Str("client_ip", clientIp).
				Logger()

			ctx := reqLog.WithContext(r.Context())
			next.ServeHTTP(ww, r.WithContext(ctx))
			reqLog.Info().
				Int("status", ww.Status()).
				Int("bytes", ww.BytesWritten()).
				Dur("duration", time.Since(start)).
				Str("agent", r.UserAgent()).
				Msg("request completed")
		})
	}
}

func realIP(r *http.Request) string {
	// If you're behind a trusted reverse proxy/load balancer, prefer X-Forwarded-For.
	// Only trust this if added by your infra; otherwise fall back to RemoteAddr.
	if xf := r.Header.Get("X-Forwarded-For"); xf != "" {
		parts := strings.Split(xf, ",")
		return strings.TrimSpace(parts[0])
	}
	if xr := r.Header.Get("X-Real-IP"); xr != "" {
		return xr
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
