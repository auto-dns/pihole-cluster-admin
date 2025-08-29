package healthcheck

import (
	"context"
	"net/http"
	"time"

	"github.com/go-chi/chi"
)

func Register(r chi.Router, d Deps) {
	r.Get("/health/live", healthcheckLive())
	r.Get("/health/ready", healthcheckReady(d))
}

func healthcheckLive() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "OK"}`))
	}
}

func healthcheckReady(d Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 500*time.Millisecond)
		defer cancel()

		status := http.StatusOK
		bodyBytes := []byte(`{"status": "OK"}`)

		w.Header().Set("Content-Type", "application/json")
		if err := d.Db.PingContext(ctx); err != nil {
			d.Logger.Warn().Err(err).Msg("readiness: db ping failed")
			status = http.StatusServiceUnavailable
			bodyBytes = []byte(`{"status": "unready"}`)
		}

		w.WriteHeader(status)
		w.Write(bodyBytes)
	}
}
