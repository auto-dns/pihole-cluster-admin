package cluster

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi"
	"github.com/rs/zerolog"
)

type Handler struct {
	service service
	logger  zerolog.Logger
}

func NewHandler(service service, logger zerolog.Logger) *Handler {
	return &Handler{
		service: service,

		logger: logger,
	}
}

func (h *Handler) Register(r chi.Router) {
	r.Get("/", h.get)
	r.Post("/", h.post)
}

func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	state, err := h.service.GetState(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}

	dto := getResponseDTO{
		Summary: getSummaryDTO{
			Mode:      state.Summary.Mode,
			Unanimous: state.Summary.Unanimous,
		},
		Nodes: make(map[int64]getNodeDTO, len(state.Nodes)),
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
		node := getNodeDTO{
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

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(dto)
}

func (h *Handler) post(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(struct{}{})
}
