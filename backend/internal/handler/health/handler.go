package health

import (
	"encoding/json"
	"net/http"

	health_s "github.com/auto-dns/pihole-cluster-admin/internal/service/health"
	"github.com/go-chi/chi"
	"github.com/rs/zerolog"
)

type Handler struct {
	service service
	logger  zerolog.Logger
}

func NewHandler(service service, logger zerolog.Logger) *Handler {
	return &Handler{service: service, logger: logger}
}

func (h *Handler) Register(r chi.Router) {
	r.Get("/summary", h.getSummary)
	r.Get("/node", h.getNodeHealth)
}

func (h *Handler) getSummary(w http.ResponseWriter, r *http.Request) {
	healthSummary := h.service.GetSummary()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(healthSummary)
}

func (h *Handler) getNodeHealth(w http.ResponseWriter, r *http.Request) {
	nodeHealth := h.service.GetNodeHealth()
	nodeHealthSlice := make([]health_s.NodeHealth, 0, len(nodeHealth))
	for _, value := range nodeHealth {
		nodeHealthSlice = append(nodeHealthSlice, value)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(nodeHealthSlice)
}
