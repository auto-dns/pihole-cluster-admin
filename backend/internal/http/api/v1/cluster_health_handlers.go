package v1

import (
	"net/http"

	"github.com/auto-dns/pihole-cluster-admin/internal/http/transport"
	"github.com/go-chi/chi"
)

func registerHealth(r chi.Router, d Deps) {
	r.Get("/cluster/health", healthGet(d))
}

func healthGet(d Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		clusterHealth := d.HealthService.GetClusterHealth(r.Context())
		res := fromDomainToClusterHealthDTO(clusterHealth)
		transport.WriteJSON(w, http.StatusOK, res)
	}
}
