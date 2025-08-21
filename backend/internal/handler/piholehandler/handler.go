package piholehandler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/auto-dns/pihole-cluster-admin/internal/service/piholeservice"
	"github.com/auto-dns/pihole-cluster-admin/internal/store"
	"github.com/auto-dns/pihole-cluster-admin/internal/transport/httpx"
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
	// Read
	r.Get("/", h.getAll)
	// Write
	r.Post("/", h.add)
	r.Patch("/{id}", h.update)
	r.Delete("/{id}", h.remove)
	r.Post("/test", h.testInstanceConnection)
	r.Post("/{id}/test", h.testExistingConnection)
}

func (h *Handler) getAll(w http.ResponseWriter, r *http.Request) {
	piholes, err := h.service.GetAll()
	if err != nil {
		h.logger.Error().Err(err).Msg("error getting pihole nodes from database")
		httpx.WriteJSONError(w, "internal server error", http.StatusInternalServerError)
		return
	}

	h.logger.Debug().Int("count", len(piholes)).Msg("fetched pihole nodes from database")

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(piholes)
}

func (h *Handler) add(w http.ResponseWriter, r *http.Request) {
	// Parse request body
	var body struct {
		Scheme      string `json:"scheme"`
		Host        string `json:"host"`
		Port        int    `json:"port"`
		Name        string `json:"name"`
		Description string `json:"description"`
		Password    string `json:"password"`
	}
	if err := httpx.DecodeJSONBody(w, r, &body, 1<<20); err != nil {
		h.logger.Error().Err(err).Msg("invalid JSON body")
		httpx.WriteJSONError(w, "invalid JSON body", http.StatusBadRequest)
		return
	}

	// Validate the inputs
	if body.Scheme != "http" && body.Scheme != "https" {
		h.logger.Error().Msg("scheme must be http or https")
		httpx.WriteJSONError(w, "scheme must be http or https", http.StatusBadRequest)
		return
	}
	if strings.TrimSpace(body.Host) == "" {
		h.logger.Error().Msg("host must not be empty")
		httpx.WriteJSONError(w, "host must not be empty", http.StatusBadRequest)
		return
	}
	if body.Port <= 0 || body.Port > 65535 {
		h.logger.Error().Msg("port must be a valid TCP port")
		httpx.WriteJSONError(w, "port must be a valid TCP port", http.StatusBadRequest)
		return
	}
	if strings.TrimSpace(body.Name) == "" {
		h.logger.Error().Msg("name must not be empty")
		httpx.WriteJSONError(w, "name must not be empty", http.StatusBadRequest)
		return
	}
	if strings.TrimSpace(body.Password) == "" {
		h.logger.Error().Msg("password must not be empty")
		httpx.WriteJSONError(w, "password must not be empty", http.StatusBadRequest)
		return
	}

	// Call user store to add the node
	addParams := store.AddPiholeParams{
		Scheme:      body.Scheme,
		Host:        body.Host,
		Port:        body.Port,
		Name:        body.Name,
		Description: body.Description,
		Password:    body.Password,
	}

	insertedNode, err := h.service.Add(r.Context(), addParams)
	if err != nil {
		h.logger.Error().Err(err).Str("host", addParams.Host).Int("port", addParams.Port).Msg("adding node")
		httpx.WriteJSONErrorFromErr(w, err)
		return
	}
	h.logger.Debug().Int64("id", insertedNode.Id).Str("scheme", insertedNode.Scheme).Str("host", insertedNode.Host).Int("port", insertedNode.Port).Str("name", insertedNode.Name).Time("created_at", insertedNode.CreatedAt).Time("updated_at", insertedNode.UpdatedAt).Msg("added pihole node")

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(insertedNode)
}

func (h *Handler) update(w http.ResponseWriter, r *http.Request) {
	// Parse request body
	var body struct {
		Scheme      *string `json:"scheme"`
		Host        *string `json:"host"`
		Port        *int    `json:"port"`
		Name        *string `json:"name"`
		Description *string `json:"description"`
		Password    *string `json:"password"`
	}
	if err := httpx.DecodeJSONBody(w, r, &body, 1<<20); err != nil {
		h.logger.Error().Err(err).Msg("invalid JSON body")
		httpx.WriteJSONError(w, "invalid JSON body", http.StatusBadRequest)
		return
	}

	// Validate request
	idString := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idString, 10, 64)
	if err != nil {
		h.logger.Error().Err(err).Msg("error converting path parameter id to int64")
		httpx.WriteJSONError(w, "error processing id path parameter", http.StatusBadRequest)
		return
	}
	if id <= 0 {
		h.logger.Error().Msg("invalid id (<= 0)")
		httpx.WriteJSONError(w, "invalid id (<= 0)", http.StatusBadRequest)
		return
	}
	// Validate at least one update field set
	if body.Scheme == nil && body.Host == nil && body.Port == nil && body.Name == nil && body.Description == nil && body.Password == nil {
		h.logger.Error().Msg("must provide at least one field to update")
		httpx.WriteJSONError(w, "must provide at least one field to update", http.StatusBadRequest)
		return
	}
	// Validate content
	if body.Scheme != nil && *body.Scheme != "http" && *body.Scheme != "https" {
		h.logger.Error().Msg("scheme must be http or https")
		httpx.WriteJSONError(w, "scheme must be http or https", http.StatusBadRequest)
		return
	}
	if body.Host != nil && strings.TrimSpace(*body.Host) == "" {
		h.logger.Error().Msg("host must not be empty")
		httpx.WriteJSONError(w, "host must not be empty", http.StatusBadRequest)
		return
	}
	if body.Port != nil && (*body.Port <= 0 || *body.Port > 65535) {
		h.logger.Error().Msg("port must be a valid TCP port")
		httpx.WriteJSONError(w, "port must be a valid TCP port", http.StatusBadRequest)
		return
	}
	if body.Name != nil && strings.TrimSpace(*body.Name) == "" {
		h.logger.Error().Msg("name must not be empty")
		httpx.WriteJSONError(w, "name must not be empty", http.StatusBadRequest)
		return
	}
	if body.Password != nil && strings.TrimSpace(*body.Password) == "" {
		h.logger.Error().Msg("password must not be empty")
		httpx.WriteJSONError(w, "password must not be empty", http.StatusBadRequest)
		return
	}

	// Call user store to update the node
	updateParams := store.UpdatePiholeParams{
		Scheme:      body.Scheme,
		Host:        body.Host,
		Port:        body.Port,
		Name:        body.Name,
		Description: body.Description,
		Password:    body.Password,
	}

	updatedNode, err := h.service.Update(r.Context(), id, updateParams)
	safe := func(p *string) string {
		if p == nil {
			return ""
		}
		return *p
	}
	safeInt := func(p *int) int {
		if p == nil {
			return 0
		}
		return *p
	}
	if err != nil {
		h.logger.Error().Err(err).Str("host", safe(updateParams.Host)).Int("port", safeInt(updateParams.Port)).Msg("updating node")
		httpx.WriteJSONErrorFromErr(w, err)
		return
	}

	h.logger.Debug().Int64("id", updatedNode.Id).Str("scheme", updatedNode.Scheme).Str("host", updatedNode.Host).Int("port", updatedNode.Port).Time("created_at", updatedNode.CreatedAt).Str("name", updatedNode.Name).Time("updated_at", updatedNode.UpdatedAt).Msg("updated pihole node")

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(updatedNode)
}

func (h *Handler) remove(w http.ResponseWriter, r *http.Request) {
	idString := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idString, 10, 64)
	if err != nil {
		h.logger.Error().Err(err).Msg("error converting path parameter id to int64")
		httpx.WriteJSONError(w, "error processing id path parameter", http.StatusBadRequest)
		return
	}
	if id <= 0 {
		h.logger.Error().Msg("invalid id (<= 0)")
		httpx.WriteJSONError(w, "invalid id (<= 0)", http.StatusBadRequest)
		return
	}

	found, err := h.service.Remove(r.Context(), id)
	if err != nil {
		h.logger.Error().Err(err).Int64("id", id).Msg("error removing pihole node")
		httpx.WriteJSONError(w, "internal server error", http.StatusInternalServerError)
		return
	}

	if !found {
		h.logger.Error().Int64("id", id).Msg("pihole not found")
		w.WriteHeader(http.StatusNotFound)
		return
	}

	h.logger.Debug().Int64("id", id).Msg("pihole removed")
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) testInstanceConnection(w http.ResponseWriter, r *http.Request) {
	// Used to test a pihole instance that hasn't been turned into a cluster yet
	var body piholeservice.TestInstanceConnectionParams
	if err := httpx.DecodeJSONBody(w, r, &body, 1<<20); err != nil {
		httpx.WriteJSONError(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Validate scheme
	body.Scheme = strings.ToLower(strings.TrimSpace(body.Scheme))
	switch body.Scheme {
	case "http", "https":
	default:
		httpx.WriteJSONError(w, "scheme must be http or https", http.StatusBadRequest)
		return
	}
	// Validate host
	body.Host = strings.TrimSpace(body.Host)
	if body.Host == "" {
		httpx.WriteJSONError(w, "host is required", http.StatusBadRequest)
		return
	}
	// Validate port
	if body.Port == 0 {
		if body.Scheme == "https" {
			body.Port = 443
		} else {
			body.Port = 80
		}
	}
	if body.Port < 1 || body.Port > 65535 {
		httpx.WriteJSONError(w, "invalid port", http.StatusBadRequest)
		return
	}
	// Validate password
	if body.Password == "" {
		httpx.WriteJSONError(w, "password is required", http.StatusBadRequest)
		return
	}

	if err := h.service.TestInstanceConnection(r.Context(), body); err != nil {
		httpx.WriteJSONError(w, "internal server error", http.StatusInternalServerError)
		return
	}

	h.logger.Debug().Str("scheme", body.Scheme).Str("host", body.Host).Int("port", body.Port).Msg("successfully logged in with pihole instance")

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) testExistingConnection(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id < 0 {
		httpx.WriteJSONError(w, "invalid id", http.StatusBadRequest)
		return
	}

	var body piholeservice.TestExistingConnectionParams
	if err := httpx.DecodeJSONBody(w, r, &body, 1<<20); err != nil {
		httpx.WriteJSONError(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if err := h.service.TestExistingConnection(r.Context(), id, body); err != nil {
		httpx.WriteJSONError(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
