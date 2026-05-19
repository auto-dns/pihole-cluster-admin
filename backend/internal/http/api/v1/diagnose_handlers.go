package v1

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/auto-dns/pihole-cluster-admin/internal/domain"
	"github.com/go-chi/chi"
)

type diagnoseEventDTO struct {
	Domain    string `json:"domain"`
	Client    string `json:"client"`
	Node      string `json:"node"`
	Status    string `json:"status"`
	Timestamp string `json:"timestamp"`
}

func registerDiagnose(r chi.Router, d Deps) {
	r.Get("/diagnose/stream", diagnoseStream(d))
}

func diagnoseStream(d Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		clientIP := strings.TrimSpace(r.URL.Query().Get("client_ip"))

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

		io.WriteString(w, ": hello\n\n")
		flusher.Flush()

		since := time.Now().UTC()
		seen := make(map[string]bool)

		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-r.Context().Done():
				return
			case <-ticker.C:
				until := time.Now().UTC()

				query := domain.QueryLogQuery{
					Filters: domain.QueryLogFilters{
						From:  &since,
						Until: &until,
					},
				}
				if clientIP != "" {
					query.Filters.ClientIP = &clientIP
				}

				resp, err := d.QueryLogService.Fetch(r.Context(), query)
				if err != nil {
					d.Logger.Warn().Err(err).Msg("diagnose: failed to fetch query logs")
					since = until
					continue
				}

				for _, nr := range resp.Results {
					if !nr.Success || nr.Response == nil {
						continue
					}
					for _, entry := range nr.Response.Entries {
						if !diagnoseIsBlocked(entry.Status) || seen[entry.Domain] {
							continue
						}
						seen[entry.Domain] = true

						evt := diagnoseEventDTO{
							Domain:    entry.Domain,
							Client:    entry.ClientIP,
							Node:      nr.PiholeNode.Name,
							Status:    entry.Status,
							Timestamp: entry.Time.UTC().Format(time.RFC3339),
						}
						data, err := json.Marshal(evt)
						if err != nil {
							continue
						}
						fmt.Fprintf(w, "event: blocked\ndata: %s\n\n", data)
						flusher.Flush()
					}
				}

				since = until
			}
		}
	}
}

func diagnoseIsBlocked(status string) bool {
	switch status {
	case "GRAVITY", "BLACKLIST", "REGEX_BLACKLIST",
		"GRAVITY_CNAME", "REGEX_CNAME", "BLACKLIST_CNAME":
		return true
	}
	return strings.HasPrefix(status, "BLOCKED") || strings.HasPrefix(status, "EXTERNAL_BLOCKED")
}
