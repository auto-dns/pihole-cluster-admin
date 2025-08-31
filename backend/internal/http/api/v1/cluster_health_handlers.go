package v1

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi"
)

func registerHealth(r chi.Router, d Deps) {
	r.Get("/cluster/health", healthGet(d))
}

func healthGet(d Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		clusterHealth := d.HealthService.GetClusterHealth(r.Context())
		res := fromDomainToClusterHealthDTO(clusterHealth)

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(res)
	}
}
