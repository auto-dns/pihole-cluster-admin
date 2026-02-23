package v1

import (
	"net/http"
	"time"

	"github.com/auto-dns/pihole-cluster-admin/internal/domain"
	"github.com/auto-dns/pihole-cluster-admin/internal/http/transport"
	"github.com/go-chi/chi"
)

func registerClusterBlocking(r chi.Router, d Deps) {
	r.Route("/cluster/blocking", func(r chi.Router) {
		r.Get("/", clusterBlockingGet(d))
		r.Post("/", clusterBlockingPost(d))
		r.Post("/nodes/{id}", clusterBlockingPostNode(d))
	})
}

func clusterBlockingGet(d Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		state, err := d.ClusterBlockingService.GetState(r.Context())
		if err != nil {
			transport.WriteErr(w, err)
			return
		}

		dto := clusterBlockingResponseFromDomain(state)
		transport.WriteJSON(w, http.StatusOK, dto)
	}
}

func clusterBlockingPost(d Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body setClusterBlockingStateReqDTO
		if err := transport.DecodeJSONBody(w, r, &body, 1<<20); err != nil {
			transport.WriteErr(w, err)
			return
		}

		state, err := d.ClusterBlockingService.SetState(r.Context(), body.Blocking, body.Timer)
		if err != nil {
			transport.WriteErr(w, err)
			return
		}

		dto := clusterBlockingResponseFromDomain(state)
		transport.WriteJSON(w, http.StatusOK, dto)
	}
}

func clusterBlockingPostNode(d Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		nodeID, ok := ParseInt64Param(r, "id", 1)
		if !ok {
			transport.WriteBadRequestErr(w, "invalid node id", nil)
			return
		}

		var body setClusterBlockingStateReqDTO
		if err := transport.DecodeJSONBody(w, r, &body, 1<<20); err != nil {
			transport.WriteErr(w, err)
			return
		}

		state, err := d.ClusterBlockingService.SetStateForNode(r.Context(), nodeID, body.Blocking, body.Timer)
		if err != nil {
			transport.WriteErr(w, err)
			return
		}

		dto := clusterBlockingResponseFromDomain(state)
		transport.WriteJSON(w, http.StatusOK, dto)
	}
}

func clusterBlockingResponseFromDomain(state *domain.ClusterBlockingState) getClusterBlockingResponseDTO {
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
		var errStr string
		if n.Error != nil {
			errStr = n.Error.Error()
		}
		node := clusterBlockingNodeDTO{
			Blocking: "unknown",
			Took:     0,
			Error:    errStr,
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

	return dto
}
