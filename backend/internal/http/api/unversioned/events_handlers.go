package unversioned

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/auto-dns/pihole-cluster-admin/internal/domain"
	v1 "github.com/auto-dns/pihole-cluster-admin/internal/http/api/v1"
	"github.com/auto-dns/pihole-cluster-admin/internal/realtime"
	"github.com/go-chi/chi"
)

func registerEvents(r chi.Router, d Deps) {
	r.Get("/events", eventsHandle(d))
}

func eventsHandle(d Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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
		io.WriteString(w, ": hello\n")
		io.WriteString(w, "retry: 3000\n\n")
		flusher.Flush()

		events, cancel := d.EventsService.Subscribe(r.Context(), topics)
		defer cancel()

		heartbeat := time.NewTicker(time.Duration(d.Cfg.HeartbeatSeconds) * time.Second)
		defer heartbeat.Stop()

		writeEvent := func(topic string, payload any) error {
			var body any
			switch topic {
			case realtime.TopicClusterBlockingV1:
				if s, ok := payload.(domain.ClusterBlockingState); ok {
					body = v1.ToClusterBlockingStateEvent(s)
				} else {
					return nil // or log type mismatch
				}
			case realtime.TopicClusterHealthV1:
				if s, ok := payload.(domain.ClusterHealth); ok {
					body = v1.ToClusterHealthEvent(s)
				} else {
					return nil // or log type mismatch
				}
			default:
				// Unknown topic: ignore
				return nil
			}

			var buf bytes.Buffer
			enc := json.NewEncoder(&buf)
			enc.SetEscapeHTML(true)
			if err := enc.Encode(body); err != nil {
				return err
			}
			jsonLine := strings.TrimRight(buf.String(), "\n")

			// SSE format: event + data lines
			if topic != "" {
				if _, err := io.WriteString(w, fmt.Sprintf("event: %s\n", topic)); err != nil {
					return err
				}
			}
			// Make sure we prefix each line with "data: "
			for _, line := range strings.Split(jsonLine, "\n") {
				if _, err := io.WriteString(w, fmt.Sprintf("data: %s\n", line)); err != nil {
					return err
				}
			}
			_, _ = io.WriteString(w, "\n")
			flusher.Flush()
			return nil
		}

		for {
			select {
			case event, ok := <-events:
				if !ok {
					return
				}
				if err := writeEvent(event.Topic, event.Payload); err != nil {
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
}

func parseTopics(val string) []string {
	if strings.TrimSpace(val) == "" {
		return []string{realtime.TopicClusterBlockingV1, realtime.TopicClusterHealthV1}
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
