package api

import (
	"net/http"

	"github.com/rs/zerolog"
)

type Handler struct {
	Logger zerolog.Logger
}

func NewHandler(logger zerolog.Logger) *Handler {
	return &Handler{
		Logger: logger,
	}
}

func (h *Handler) Healthcheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"status": "OK"}`))
}
