package healthcheckhandler

import (
	"net/http"

	"github.com/go-chi/chi"
	"github.com/rs/zerolog"
)

type Handler struct {
	logger zerolog.Logger
}

func NewHandler(logger zerolog.Logger) *Handler {
	return &Handler{logger: logger}
}

func (h *Handler) Register(r chi.Router) {
	r.Get("/", h.healthcheck)
}

func (h *Handler) healthcheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status": "OK"}`))
}
