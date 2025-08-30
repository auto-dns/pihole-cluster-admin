package v1

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/auto-dns/pihole-cluster-admin/internal/http/helpers"
	"github.com/go-chi/chi"
)

func registerClusterBlocking(r chi.Router, d Deps) {
	r.Get("/cluster/blocking", clusterBlockingGet(d))
	r.Post("/cluster/blocking", clusterBlockingPost(d))
}

func clusterBlockingGet(d Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		state, err := d.ClusterBlockingService.GetState(r.Context())
		if err != nil {
			helpers.WriteErr(w, err)
			return
		}

		dto := getClusterBlockingResponseDTO{
			Summary: clusterBlockingSummaryDTO{
				Mode:      state.Summary.Mode,
				Unanimous: state.Summary.Unanimous,
			},
			Nodes: make(map[int64]clusterBlockingNodeDTO, len(state.Nodes)),
		}
		dto.Summary.Counts = struct {
			Total    int `json:"total"`
			Enabled  int `json:"enabled"`
			Disabled int `json:"disabled"`
			Failed   int `json:"failed"`
			Errors   int `json:"errors"`
		}{
			Total:    state.Summary.Total,
			Enabled:  state.Summary.Enabled,
			Disabled: state.Summary.Disabled,
			Failed:   state.Summary.Failed,
			Errors:   state.Summary.Errors,
		}

		if state.Summary.MinTimer != nil || state.Summary.MaxTimer != nil {
			dto.Summary.Timers.Present = true
			if state.Summary.MinTimer != nil {
				v := int64(state.Summary.MinTimer.Round(time.Second).Seconds())
				dto.Summary.Timers.MinSeconds = &v
			}
			if state.Summary.MaxTimer != nil {
				v := int64(state.Summary.MaxTimer.Round(time.Second).Seconds())
				dto.Summary.Timers.MaxSeconds = &v
			}
		}
		dto.Summary.Took.MaxSeconds = state.Summary.MaxTook.Seconds()
		dto.Summary.Took.AvgSeconds = state.Summary.AvgTook.Seconds()

		for id, n := range state.Nodes {
			node := clusterBlockingNodeDTO{
				Blocking: "unknown",
				Took:     0,
				Error:    n.ErrorMessage(),
			}
			node.Node.Id = n.PiholeNode.Id
			node.Node.Name = n.PiholeNode.Name
			node.Node.Host = n.PiholeNode.Host

			if n.Success && n.Response != nil {
				node.Blocking = string(n.Response.Status)
				node.Took = n.Response.Took.Seconds()
				if n.Response.TimerLeft != nil {
					v := int64(n.Response.TimerLeft.Round(time.Second).Seconds())
					node.Timer = &v
				}
			}

			dto.Nodes[id] = node
		}

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(dto)
	}
}

func clusterBlockingPost(_ Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(struct{}{})
	}
}
