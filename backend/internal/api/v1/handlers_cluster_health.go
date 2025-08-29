package v1

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi"
)

func registerHealth(r chi.Router, d Deps) {
	r.Get("/cluster/summary", healthGetSummary(d))
	r.Get("/cluster/node", healthGetNodeHealth(d))
}

func healthGetSummary(d Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		healthSummary := d.HealthService.GetSummary()

		res := fromDomainClusterHealthSummary(healthSummary)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(res)
	}
}

func healthGetNodeHealth(d Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		nodeHealth := d.HealthService.GetNodeHealth()
		nodeHealthSlice := make([]clusterNodeHealthDTO, 0, len(nodeHealth))
		for _, value := range nodeHealth {
			nodeHealthSlice = append(nodeHealthSlice, fromDomainClusterNodeHealth(value))
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(nodeHealthSlice)
	}
}
