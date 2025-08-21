package eventshandler

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/auto-dns/pihole-cluster-admin/internal/config"
	"github.com/go-chi/chi"
	"github.com/rs/zerolog"
)

type Handler struct {
	service service
	cfg     config.ServerSideEventsConfig
	logger  zerolog.Logger
}

func NewHandler(cfg config.ServerSideEventsConfig, service service, logger zerolog.Logger) *Handler {
	return &Handler{
		service: service,
		cfg:     cfg,
		logger:  logger,
	}
}

func (h *Handler) Register(r chi.Router) {
	r.Get("/", h.handleEvents)
}

func (h *Handler) handleEvents(w http.ResponseWriter, r *http.Request) {
	topics := parseTopics(r.URL.Query().Get("topics"))

	// Server Side Events headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache, no-transform")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("X-Accel-Buffering", "no")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}

	// Initial comment to allow proxies to keep the connection alive
	_, _ = io.WriteString(w, ": hello\n")
	_, _ = io.WriteString(w, "retry: 3000\n\n")
	flusher.Flush()

	events, cancel := h.service.Subscribe(r.Context(), topics)
	defer cancel()

	heartbeat := time.NewTicker(time.Duration(h.cfg.HeartbeatSeconds) * time.Second)
	defer heartbeat.Stop()

	writeEvent := func(topic string, data []byte) error {
		if topic != "" {
			if _, err := fmt.Fprintf(w, "event: %s\n", topic); err != nil {
				return err
			}
		}
		if _, err := w.Write([]byte("data: ")); err != nil {
			return err
		}
		if _, err := w.Write(data); err != nil {
			return err
		}
		if _, err := w.Write([]byte("\n\n")); err != nil {
			return err
		}
		flusher.Flush()
		return nil
	}

	for {
		select {
		case event, ok := <-events:
			if !ok {
				return
			}
			if err := writeEvent(event.Topic, event.Data); err != nil {
				return // Client likely disconnected
			}
		case <-heartbeat.C:
			if _, err := io.WriteString(w, ": ping\n\n"); err != nil {
				return
			}
			flusher.Flush()
		case <-r.Context().Done():
			return
		}
	}
}

func parseTopics(val string) []string {
	if strings.TrimSpace(val) == "" {
		return []string{"health_summary", "node_health"}
	}
	parts := strings.Split(val, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if t := strings.TrimSpace(p); t != "" {
			out = append(out, t)
		}
	}
	return out
}
