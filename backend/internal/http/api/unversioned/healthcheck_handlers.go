package unversioned

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi"
)

func registerHealthcheck(r chi.Router, d Deps) {
	r.Get("/health/live", healthcheckLive())
	r.Get("/health/ready", healthcheckReady(d))
}

func healthcheckLive() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		res := healthcheckResponseDTO{Status: HealthcheckStatusOk}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(res)
	}
}

func healthcheckReady(d Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 500*time.Millisecond)
		defer cancel()

		status := http.StatusOK
		res := healthcheckResponseDTO{Status: HealthcheckStatusOk}

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		if err := d.Db.PingContext(ctx); err != nil {
			d.Logger.Warn().Err(err).Msg("readiness: db ping failed")
			status = http.StatusServiceUnavailable
			res.Status = HealthcheckStatusUnready
		}

		w.WriteHeader(status)
		_ = json.NewEncoder(w).Encode(res)
	}
}
