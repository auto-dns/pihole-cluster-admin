package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/auto-dns/pihole-cluster-admin/internal/config"
	"github.com/auto-dns/pihole-cluster-admin/internal/domain"
	"github.com/auto-dns/pihole-cluster-admin/internal/health"
	"github.com/auto-dns/pihole-cluster-admin/internal/pihole"
	"github.com/auto-dns/pihole-cluster-admin/internal/sessions"
	"github.com/auto-dns/pihole-cluster-admin/internal/store"
	"github.com/go-chi/chi"
	"github.com/rs/zerolog"
)

func ptrInt64(v int64) *int64 { return &v }

type Handler struct {
	cluster         piholeCluster
	sessions        sessionDeps
	initStatusStore initStatusStore
	piholeStore     piholeStore
	userStore       userStore
	healthService   healthService
	eventSubscriber eventSubscriber
	logger          zerolog.Logger
	cfg             config.ServerConfig
}

func NewHandler(cluster piholeCluster, sessions sessionDeps, initStatusStore initStatusStore, piholeStore piholeStore, userStore userStore, healthService healthService, eventSubscriber eventSubscriber, cfg config.ServerConfig, logger zerolog.Logger) *Handler {
	return &Handler{
		cluster:         cluster,
		sessions:        sessions,
		initStatusStore: initStatusStore,
		piholeStore:     piholeStore,
		userStore:       userStore,
		healthService:   healthService,
		eventSubscriber: eventSubscriber,
		logger:          logger,
		cfg:             cfg,
	}
}

// Convenience function
func writeJSONError(w http.ResponseWriter, message string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"error": message,
	})
}

func decodeJSONBody(w http.ResponseWriter, r *http.Request, dst any, maxBytes int64) error {
	r.Body = http.MaxBytesReader(w, r.Body, maxBytes)
	defer r.Body.Close()

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	if err := dec.Decode(dst); err != nil {
		return err
	}
	return nil
}

// Handlers

func (h *Handler) AuthMiddleware(next http.Handler) http.Handler {
	return h.sessions.AuthMiddleware(next)
}

// Unauthenticated routes

func (h *Handler) Healthcheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status": "OK"}`))
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var creds struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := decodeJSONBody(w, r, &creds, 1<<20); err != nil {
		writeJSONError(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Validate against the database
	user, err := h.userStore.ValidateUser(creds.Username, creds.Password)
	var wrongPasswordErr *store.WrongPasswordError
	switch {
	case errors.Is(err, sql.ErrNoRows):
		h.logger.Warn().Str("username", creds.Username).Msg("Invalid username or password")
		writeJSONError(w, "Invalid credentials", http.StatusUnauthorized)
		return
	case errors.As(err, &wrongPasswordErr):
		h.logger.Warn().Str("username", creds.Username).Msg("Invalid username or password")
		writeJSONError(w, "Invalid credentials", http.StatusUnauthorized)
		return
	case err != nil:
		h.logger.Error().Err(err).Str("username", creds.Username).Msg("error validating user")
		writeJSONError(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Successful login → create session
	h.logger.Info().Int64("userId", user.Id).Msg("user login success")
	sessionID, err := h.sessions.CreateSession(user.Id)
	if err != nil {
		h.logger.Error().Err(err).Int64("userId", user.Id).Msg("error creating session")
	}

	http.SetCookie(w, h.sessions.Cookie(sessionID))
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(user)
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(h.cfg.Session.CookieName)
	if err == nil {
		sessionId := cookie.Value
		userId, ok, err := h.sessions.GetUserId(sessionId)
		if err != nil {
			h.logger.Error().Err(err).Int64("userId", userId).Msg("error getting user session")
		} else if ok {
			h.logger.Info().Int64("userId", userId).Msg("user logged out")
		} else {
			h.logger.Warn().Msg("user attempted logout, but no username was found in the session")
		}
		h.sessions.DestroySession(sessionId)
		expired := h.sessions.Cookie("")
		expired.Expires = time.Now().Add(-1 * time.Hour)
		http.SetCookie(w, expired)
	} else {
		h.logger.Info().Msg("user attempted logout but did not have a session")
	}
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) GetIsInitialized(w http.ResponseWriter, r *http.Request) {
	initialized, err := h.userStore.IsInitialized()
	if err != nil {
		h.logger.Error().Err(err).Msg("failed to get app initialization status")
		writeJSONError(w, "server error", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(map[string]bool{"initialized": initialized})
}

func (h *Handler) GetInitializationStatus(w http.ResponseWriter, r *http.Request) {
	initializationStatus, err := h.initStatusStore.GetInitializationStatus()
	if err != nil {
		h.logger.Error().Err(err).Msg("failed to get app initialization status")
		writeJSONError(w, "server error", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(initializationStatus)
}

func (h *Handler) UpdatePiholeInitializationStatus(w http.ResponseWriter, r *http.Request) {
	// Parse request body
	var body struct {
		Status domain.PiholeStatus `json:"status"`
	}
	if err := decodeJSONBody(w, r, &body, 1<<20); err != nil {
		h.logger.Error().Err(err).Msg("invalid JSON body")
		writeJSONError(w, "invalid JSON body", http.StatusBadRequest)
		return
	}
	logger := h.logger.With().Str("new_pihole_status", string(body.Status)).Logger()

	// Fetch current initialization status from store
	currStatus, err := h.initStatusStore.GetInitializationStatus()
	if err != nil {
		logger.Error().Err(err).Msg("failed to get app initialization status")
		writeJSONError(w, "server error", http.StatusInternalServerError)
		return
	}
	logger = logger.With().Str("current_pihole_status", string(currStatus.PiholeStatus)).Logger()

	// Disallow updating to same status as current
	if body.Status == currStatus.PiholeStatus {
		logger.Error().Msg("illegal operation: new status is same as current status")
		writeJSONError(w, "cannot update status to same status", http.StatusBadRequest)
		return
	}

	// Handle each inbound status
	switch body.Status {
	// Requesting to set uninitialized
	case domain.PiholeUninitialized:
		logger.Error().Msg("illegal operation: cannot update status to UNINITIALIZED")
		writeJSONError(w, "cannot update status to UNINITIALIZED", http.StatusBadRequest)
		return
	// Requesting to set added
	case domain.PiholeAdded:
		// Allow setting to "added" from all statuses
	// Requesting to set skipped
	case domain.PiholeSkipped:
		// Disallow setting to "skipped" from "added"
		if currStatus.PiholeStatus == domain.PiholeAdded {
			logger.Error().Msg("illegal operation: cannot update status from ADDED to SKIPPED")
			writeJSONError(w, "cannot update status from ADDED to SKIPPED", http.StatusBadRequest)
			return
		}
	}

	err = h.initStatusStore.SetPiholeStatus(body.Status)
	if err != nil {
		logger.Error().Err(err).Msg("setting pihole initialization status in store")
		writeJSONError(w, "internal server error", http.StatusInternalServerError)
		return
	}

	logger.Debug().Msg("updated pihole init status in store")
	w.WriteHeader(http.StatusNoContent)
}

// Authenticated routes

// Event Streaming

func (h *Handler) HandleEvents(w http.ResponseWriter, r *http.Request) {
	topicsParam := r.URL.Query().Get("topics")
	var topics []string
	if topicsParam == "" {
		topics = []string{"health_summary", "node_health"}
	} else {
		for _, t := range strings.Split(topicsParam, ",") {
			if t = strings.TrimSpace(t); t != "" {
				topics = append(topics, t)
			}
		}
	}

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

	ch, cancel := h.eventSubscriber.Subscribe(topics)
	defer cancel()

	heartbeat := time.NewTicker(time.Duration(h.cfg.ServerSideEvents.HeartbeatSeconds) * time.Second)
	defer heartbeat.Stop()

	// Initial comment to allow proxies to keep the connection alive
	_, _ = io.WriteString(w, ": hello\n")
	_, _ = io.WriteString(w, "retry: 3000\n\n")
	flusher.Flush()

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
		case event := <-ch:
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

// -- Health Status

func (h *Handler) GetHealthSummary(w http.ResponseWriter, r *http.Request) {
	healthSummary := h.healthService.Summary()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(healthSummary)
}

func (h *Handler) GetNodeHealth(w http.ResponseWriter, r *http.Request) {
	nodeHealth := h.healthService.NodeHealth()
	nodeHealthSlice := make([]health.NodeHealth, 0, len(nodeHealth))
	for _, value := range nodeHealth {
		nodeHealthSlice = append(nodeHealthSlice, value)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(nodeHealthSlice)
}

// -- User

func (h *Handler) GetSessionUser(w http.ResponseWriter, r *http.Request) {
	userId, ok := r.Context().Value(sessions.UserIdContextKey).(int64)
	if !ok {
		writeJSONError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	user, err := h.userStore.GetUser(userId)
	if err != nil {
		h.logger.Error().Err(err).Msg("user session not found")
		writeJSONError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	h.logger.Debug().Int64("id", user.Id).Str("username", user.Username).Msg("user fetched from database")

	userResponse := UserResponse{
		Username:  user.Username,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(userResponse)
}

// -- Pihole CRUD routes

type PiholeResponse struct {
	Id          int64  `json:"id"`
	Scheme      string `json:"scheme"`
	Host        string `json:"host"`
	Port        int    `json:"port"`
	Name        string `json:"name"`
	Description string `json:"description"`
	// Omit the password
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func (h *Handler) AddPiholeNode(w http.ResponseWriter, r *http.Request) {
	// Parse request body
	var body struct {
		Scheme      string `json:"scheme"`
		Host        string `json:"host"`
		Port        int    `json:"port"`
		Name        string `json:"name"`
		Description string `json:"description"`
		Password    string `json:"password"`
	}
	if err := decodeJSONBody(w, r, &body, 1<<20); err != nil {
		h.logger.Error().Err(err).Msg("invalid JSON body")
		writeJSONError(w, "invalid JSON body", http.StatusBadRequest)
		return
	}

	// Validate the inputs
	if body.Scheme != "http" && body.Scheme != "https" {
		h.logger.Error().Msg("scheme must be http or https")
		writeJSONError(w, "scheme must be http or https", http.StatusBadRequest)
		return
	}
	if strings.TrimSpace(body.Host) == "" {
		h.logger.Error().Msg("host must not be empty")
		writeJSONError(w, "host must not be empty", http.StatusBadRequest)
		return
	}
	if body.Port <= 0 || body.Port > 65535 {
		h.logger.Error().Msg("port must be a valid TCP port")
		writeJSONError(w, "port must be a valid TCP port", http.StatusBadRequest)
		return
	}
	if strings.TrimSpace(body.Name) == "" {
		h.logger.Error().Msg("name must not be empty")
		writeJSONError(w, "name must not be empty", http.StatusBadRequest)
		return
	}
	if strings.TrimSpace(body.Password) == "" {
		h.logger.Error().Msg("password must not be empty")
		writeJSONError(w, "password must not be empty", http.StatusBadRequest)
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

	insertedNode, err := h.piholeStore.AddPiholeNode(addParams)
	if err != nil {
		if strings.Contains(err.Error(), "piholes.host") {
			h.logger.Error().Err(err).Str("host", addParams.Host).Int("port", addParams.Port).Msg("duplicate host:port")
			writeJSONError(w, "duplicate host:port", http.StatusConflict)
			return
		}
		if strings.Contains(err.Error(), "piholes.name") {
			h.logger.Error().Err(err).Str("name", addParams.Name).Msg("duplicate name")
			writeJSONError(w, "duplicate name", http.StatusConflict)
			return
		}
		// generic fallback
		h.logger.Error().Err(err).Str("scheme", addParams.Scheme).Str("name", addParams.Name).Str("host", addParams.Host).Int("port", addParams.Port).Str("description", addParams.Description).Msg("error adding pihole")
		writeJSONError(w, "failed to add pihole node", http.StatusInternalServerError)
		return
	}
	h.logger.Debug().Int64("id", insertedNode.Id).Str("scheme", insertedNode.Scheme).Str("host", insertedNode.Host).Int("port", insertedNode.Port).Str("name", insertedNode.Name).Time("created_at", insertedNode.CreatedAt).Time("updated_at", insertedNode.UpdatedAt).Msg("added pihole node to database")

	nodeSecret, err := h.piholeStore.GetPiholeNodeSecret(insertedNode.Id)
	if err != nil {
		h.logger.Error().Err(err).Str("scheme", insertedNode.Scheme).Str("name", insertedNode.Name).Str("host", insertedNode.Host).Int("port", insertedNode.Port).Str("description", insertedNode.Description).Msg("error getting pihole secret")
		writeJSONError(w, "failed to add pihole node", http.StatusInternalServerError)
		return
	}

	// Add client to cluster
	cfg := &pihole.ClientConfig{
		Id:       insertedNode.Id,
		Name:     insertedNode.Name,
		Scheme:   insertedNode.Scheme,
		Host:     insertedNode.Host,
		Port:     insertedNode.Port,
		Password: nodeSecret.Password,
	}
	client := pihole.NewClient(cfg, h.logger)
	err = h.cluster.AddClient(r.Context(), client)
	if err != nil {
		h.logger.Error().Err(err).Int64("id", insertedNode.Id).Msg("adding client to cluster")
		return
	}
	h.logger.Debug().Int64("id", insertedNode.Id).Msg("added pihole node to cluster")

	response := PiholeResponse{
		Id:          insertedNode.Id,
		Scheme:      insertedNode.Scheme,
		Host:        insertedNode.Host,
		Port:        insertedNode.Port,
		Name:        insertedNode.Name,
		Description: insertedNode.Description,
		CreatedAt:   insertedNode.CreatedAt,
		UpdatedAt:   insertedNode.UpdatedAt,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

func (h *Handler) UpdatePiholeNode(w http.ResponseWriter, r *http.Request) {
	// Parse request body
	var body struct {
		Scheme      *string `json:"scheme"`
		Host        *string `json:"host"`
		Port        *int    `json:"port"`
		Name        *string `json:"name"`
		Description *string `json:"description"`
		Password    *string `json:"password"` // Updating this field is optional
	}
	if err := decodeJSONBody(w, r, &body, 1<<20); err != nil {
		h.logger.Error().Err(err).Msg("invalid JSON body")
		writeJSONError(w, "invalid JSON body", http.StatusBadRequest)
		return
	}

	// Validate request
	idString := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idString, 10, 64)
	if err != nil {
		h.logger.Error().Err(err).Msg("error converting path parameter id to int64")
		writeJSONError(w, "error processing id path parameter", http.StatusBadRequest)
		return
	}
	if id <= 0 {
		h.logger.Error().Msg("invalid id (<= 0)")
		writeJSONError(w, "invalid id (<= 0)", http.StatusBadRequest)
		return
	}
	// Validate at least one update field set
	if body.Scheme == nil && body.Host == nil && body.Port == nil && body.Name == nil && body.Description == nil && body.Password == nil {
		h.logger.Error().Msg("must provide at least one field to update")
		writeJSONError(w, "must provide at least one field to update", http.StatusBadRequest)
		return
	}
	// Validate content
	if body.Scheme != nil && *body.Scheme != "http" && *body.Scheme != "https" {
		h.logger.Error().Msg("scheme must be http or https")
		writeJSONError(w, "scheme must be http or https", http.StatusBadRequest)
		return
	}
	if body.Host != nil && strings.TrimSpace(*body.Host) == "" {
		h.logger.Error().Msg("host must not be empty")
		writeJSONError(w, "host must not be empty", http.StatusBadRequest)
		return
	}
	if body.Port != nil && (*body.Port <= 0 || *body.Port > 65535) {
		h.logger.Error().Msg("port must be a valid TCP port")
		writeJSONError(w, "port must be a valid TCP port", http.StatusBadRequest)
		return
	}
	if body.Name != nil && strings.TrimSpace(*body.Name) == "" {
		h.logger.Error().Msg("name must not be empty")
		writeJSONError(w, "name must not be empty", http.StatusBadRequest)
		return
	}
	if body.Password != nil && strings.TrimSpace(*body.Password) == "" {
		h.logger.Error().Msg("password must not be empty")
		writeJSONError(w, "password must not be empty", http.StatusBadRequest)
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

	updatedNode, err := h.piholeStore.UpdatePiholeNode(id, updateParams)
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
		if strings.Contains(err.Error(), "piholes.host") {
			h.logger.Error().Err(err).Str("host", safe(updateParams.Host)).Int("port", safeInt(updateParams.Port)).Msg("duplicate host:port")
			writeJSONError(w, "duplicate host:port", http.StatusConflict)
			return
		}
		if strings.Contains(err.Error(), "piholes.name") {
			h.logger.Error().Err(err).Str("name", safe(updateParams.Name)).Msg("duplicate name")
			writeJSONError(w, "duplicate name", http.StatusConflict)
			return
		}
		// generic fallback
		h.logger.Error().Err(err).Str("scheme", safe(updateParams.Scheme)).Str("name", safe(updateParams.Name)).Str("host", safe(updateParams.Host)).Int("port", safeInt(updateParams.Port)).Str("description", safe(updateParams.Description)).Msg("error adding pihole")
		writeJSONError(w, "failed to update pihole node", http.StatusInternalServerError)
		return
	}

	h.logger.Debug().Int64("id", updatedNode.Id).Str("scheme", updatedNode.Scheme).Str("host", updatedNode.Host).Int("port", updatedNode.Port).Time("created_at", updatedNode.CreatedAt).Str("name", updatedNode.Name).Time("updated_at", updatedNode.UpdatedAt).Msg("updated pihole node")

	nodeSecret, err := h.piholeStore.GetPiholeNodeSecret(updatedNode.Id)
	if err != nil {
		h.logger.Error().Err(err).Str("scheme", updatedNode.Scheme).Str("name", updatedNode.Name).Str("host", updatedNode.Host).Int("port", updatedNode.Port).Str("description", updatedNode.Description).Msg("error getting pihole password")
		writeJSONError(w, "failed to get pihole node auth", http.StatusInternalServerError)
		return
	}

	// Update client in cluster
	cfg := &pihole.ClientConfig{
		Id:       updatedNode.Id,
		Name:     updatedNode.Name,
		Scheme:   updatedNode.Scheme,
		Host:     updatedNode.Host,
		Port:     updatedNode.Port,
		Password: nodeSecret.Password,
	}
	err = h.cluster.UpdateClient(r.Context(), cfg.Id, cfg)
	if err != nil {
		h.logger.Error().Err(err).Int64("id", updatedNode.Id).Msg("updating client in cluster")
		return
	}
	h.logger.Debug().Int64("id", updatedNode.Id).Msg("updated pihole node in cluster")

	response := PiholeResponse{
		Id:          updatedNode.Id,
		Scheme:      updatedNode.Scheme,
		Host:        updatedNode.Host,
		Port:        updatedNode.Port,
		Name:        updatedNode.Name,
		Description: updatedNode.Description,
		CreatedAt:   updatedNode.CreatedAt,
		UpdatedAt:   updatedNode.UpdatedAt,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func (h *Handler) RemovePiholeNode(w http.ResponseWriter, r *http.Request) {
	idString := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idString, 10, 64)
	if err != nil {
		h.logger.Error().Err(err).Msg("error converting path parameter id to int64")
		writeJSONError(w, "error processing id path parameter", http.StatusBadRequest)
		return
	}
	if id <= 0 {
		h.logger.Error().Msg("invalid id (<= 0)")
		writeJSONError(w, "invalid id (<= 0)", http.StatusBadRequest)
		return
	}

	found, err := h.piholeStore.RemovePiholeNode(id)
	if err != nil {
		h.logger.Error().Err(err).Int64("id", id).Msg("error removing pihole node from database")
		writeJSONError(w, "internal server error", http.StatusInternalServerError)
		return
	}

	if h.cluster.HasClient(r.Context(), id) {
		err = h.cluster.RemoveClient(r.Context(), id)
		if err != nil {
			h.logger.Error().Err(err).Int64("id", id).Msg("removing client from cluster")
			writeJSONError(w, "internal server error", http.StatusInternalServerError)
			return
		}
	}

	if found {
		h.logger.Debug().Int64("id", id).Msg("pihole removed")
		w.WriteHeader(http.StatusNoContent)
		return
	} else {
		h.logger.Error().Int64("id", id).Msg("pihole not found")
		w.WriteHeader(http.StatusNotFound)
		return
	}
}

func (h *Handler) GetAllPiholeNodes(w http.ResponseWriter, r *http.Request) {
	piholes, err := h.piholeStore.GetAllPiholeNodes()
	if err != nil {
		h.logger.Error().Err(err).Msg("error getting pihole nodes from database")
		writeJSONError(w, "internal server error", http.StatusInternalServerError)
		return
	}

	h.logger.Debug().Int("count", len(piholes)).Msg("fetched pihole nodes from database")

	response := make([]PiholeResponse, len(piholes))
	for i, pihole := range piholes {
		response[i] = PiholeResponse{
			Id:          pihole.Id,
			Scheme:      pihole.Scheme,
			Host:        pihole.Host,
			Port:        pihole.Port,
			Name:        pihole.Name,
			Description: pihole.Description,
			CreatedAt:   pihole.CreatedAt,
			UpdatedAt:   pihole.UpdatedAt,
		}
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func (h *Handler) TestExistingPiholeConnection(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id < 0 {
		writeJSONError(w, "invalid id", http.StatusBadRequest)
		return
	}

	var body struct {
		Scheme   *string `json:"scheme"`
		Host     *string `json:"host"`
		Port     *int    `json:"port"`
		Password *string `json:"password"`
	}
	if err := decodeJSONBody(w, r, &body, 1<<20); err != nil {
		writeJSONError(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Load client from store
	node, err := h.piholeStore.GetPiholeNode(id)
	if err != nil {
		writeJSONError(w, "not found", http.StatusNotFound)
		return
	}
	nodeSecret, err := h.piholeStore.GetPiholeNodeSecret(id)
	if err != nil {
		writeJSONError(w, "not found", http.StatusNotFound)
		return
	}

	// Merge overrides with existing record
	scheme := node.Scheme
	host := node.Host
	port := node.Port
	pass := nodeSecret.Password

	if body.Scheme != nil {
		scheme = strings.ToLower(strings.TrimSpace(*body.Scheme))
	}
	if body.Host != nil {
		host = strings.TrimSpace(*body.Host)
	}
	if body.Port != nil {
		port = *body.Port
	}
	if body.Password != nil && strings.TrimSpace(*body.Password) != "" {
		pass = *body.Password
	}

	// Validation
	if scheme != "http" && scheme != "https" {
		writeJSONError(w, "scheme must be http or https", http.StatusBadRequest)
		return
	}
	if host == "" {
		writeJSONError(w, "host is required", http.StatusBadRequest)
		return
	}
	if port < 1 || port > 65535 {
		writeJSONError(w, "invalid port", http.StatusBadRequest)
		return
	}
	if pass == "" {
		writeJSONError(w, "password is required", http.StatusBadRequest)
		return
	}

	// Create a new temporary test client
	httpClient := &http.Client{
		Transport: &http.Transport{Proxy: http.ProxyFromEnvironment, DisableKeepAlives: true},
		Timeout:   4 * time.Second,
	}
	cfg := &pihole.ClientConfig{Id: id, Name: node.Name, Scheme: scheme, Host: host, Port: port, Password: pass}
	testClient := pihole.NewClient(cfg, h.logger, pihole.WithHTTPClient(httpClient))

	// Log in
	if err := testClient.Login(r.Context()); err != nil {
		writeJSONError(w, "login failed", http.StatusBadRequest)
		return
	}
	// Log out
	if err := testClient.Logout(r.Context()); err != nil {
		h.logger.Warn().Err(err).Msg("test logout error")
	}
	httpClient.CloseIdleConnections()

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) TestPiholeInstanceConnection(w http.ResponseWriter, r *http.Request) {
	// Used to test a pihole instance that hasn't been turned into a cluster yet
	var body struct {
		Scheme   string `json:"scheme"`
		Host     string `json:"host"`
		Port     int    `json:"port"`
		Password string `json:"password"`
	}
	if err := decodeJSONBody(w, r, &body, 1<<20); err != nil {
		writeJSONError(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Validate scheme
	body.Scheme = strings.ToLower(strings.TrimSpace(body.Scheme))
	switch body.Scheme {
	case "http", "https":
	default:
		writeJSONError(w, "scheme must be http or https", http.StatusBadRequest)
		return
	}
	// Validate host
	body.Host = strings.TrimSpace(body.Host)
	if body.Host == "" {
		writeJSONError(w, "host is required", http.StatusBadRequest)
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
		writeJSONError(w, "invalid port", http.StatusBadRequest)
		return
	}
	// Validate password
	if body.Password == "" {
		writeJSONError(w, "password is required", http.StatusBadRequest)
		return
	}

	logger := h.logger.With().Str("scheme", body.Scheme).Str("host", body.Host).Int("port", body.Port).Logger()

	httpClient := &http.Client{
		Transport: &http.Transport{
			Proxy:             http.ProxyFromEnvironment,
			DisableKeepAlives: true,
		},
		Timeout: 4 * time.Second,
	}

	piholeConfig := &pihole.ClientConfig{
		Id: -1, Name: "",
		Scheme: body.Scheme, Host: body.Host, Port: body.Port, Password: body.Password,
	}
	testClient := pihole.NewClient(piholeConfig, logger, pihole.WithHTTPClient(httpClient))

	// Login
	if err := testClient.Login(r.Context()); err != nil {
		logger.Error().Err(err).Msg("login failed")
		writeJSONError(w, "login failed", http.StatusBadRequest)
		return
	}

	// Logout
	if err := testClient.Logout(r.Context()); err != nil {
		logger.Warn().Err(err).Msg("error logging out of test pihole client")
	}
	httpClient.CloseIdleConnections()

	logger.Debug().Msg("successfully logged in with pihole instance")
	w.WriteHeader(http.StatusNoContent)
}

// User CRUD routes

type UserResponse struct {
	Username  string    `json:"username"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func (h *Handler) CreateUser(w http.ResponseWriter, r *http.Request) {
	// Verify app not initialized
	initialized, err := h.userStore.IsInitialized()
	if err != nil {
		h.logger.Error().Err(err).Msg("error fetching app initialization status")
		writeJSONError(w, "internal server error", http.StatusInternalServerError)
		return
	}

	if initialized {
		h.logger.Error().Msg("app is already initialized")
		writeJSONError(w, "forbidden", http.StatusForbidden)
		return
	}

	// Parse request body
	var body struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := decodeJSONBody(w, r, &body, 1<<20); err != nil {
		h.logger.Error().Err(err).Msg("invalid JSON body")
		writeJSONError(w, "invalid JSON body", http.StatusBadRequest)
		return
	}

	// Validate request params
	if strings.TrimSpace(body.Username) == "" {
		h.logger.Error().Msg("empty username in body")
		writeJSONError(w, "empty username in body", http.StatusBadRequest)
		return
	}
	if strings.TrimSpace(body.Password) == "" {
		h.logger.Error().Msg("empty password in body")
		writeJSONError(w, "empty password in body", http.StatusBadRequest)
		return
	}

	// Create user
	createUserParams := store.CreateUserParams{
		Username: body.Username,
		Password: body.Password,
	}
	user, err := h.userStore.CreateUser(createUserParams)
	if err != nil {
		h.logger.Error().Err(err).Str("username", body.Username).Msg("failed to create user")
		writeJSONError(w, "internal server error", http.StatusInternalServerError)
		return
	}

	err = h.initStatusStore.SetUserCreated(true)
	if err != nil {
		h.logger.Error().Err(err).Msg("error updating initialization status")
	}

	h.logger.Debug().Int64("id", user.Id).Str("username", user.Username).Msg("created user in database")

	userResponse := UserResponse{
		Username:  user.Username,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}
	// Create a session and return a cookie
	sessionId, err := h.sessions.CreateSession(user.Id)
	if err != nil {
		h.logger.Error().Err(err).Str("username", body.Username).Msg("failed to create session")
		writeJSONError(w, "internal server error", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, h.sessions.Cookie(sessionId))
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(userResponse)
}

// Application business logic routes

func (h *Handler) FetchQueryLogs(w http.ResponseWriter, r *http.Request) {
	var req pihole.FetchQueryLogClusterRequest
	ctxLogger := h.logger.With()

	cursor := r.URL.Query().Get("cursor")
	if cursor != "" {
		// Cursor request: only cursor and optional length override
		req.Cursor = &cursor
		if v := r.URL.Query().Get("length"); v != "" {
			if i, err := strconv.Atoi(v); err == nil {
				req.Length = &i
			}
		}
		ctxLogger.Str("cursor", cursor)
		if req.Length != nil {
			ctxLogger.Int("length", *req.Length)
		}
	} else {
		// --- Parse optional timestamps (RFC3339)
		fromStr := r.URL.Query().Get("from")
		untilStr := r.URL.Query().Get("until")

		if fromStr == "" && untilStr == "" {
			until := time.Now().UTC()
			from := until.Add(-5 * time.Minute)
			req.Filters.From = ptrInt64(from.Unix())
			req.Filters.Until = ptrInt64(until.Unix())
		} else {
			if fromStr != "" {
				fromTime, err := time.Parse(time.RFC3339, fromStr)
				if err != nil {
					writeJSONError(w, "invalid 'from' time", http.StatusBadRequest)
					return
				}
				req.Filters.From = ptrInt64(fromTime.Unix())
			}
			if untilStr != "" {
				untilTime, err := time.Parse(time.RFC3339, untilStr)
				if err != nil {
					writeJSONError(w, "invalid 'until' time", http.StatusBadRequest)
					return
				}
				req.Filters.Until = ptrInt64(untilTime.Unix())
			}
		}
		ctxLogger.Int64("from", *req.Filters.From).Int64("until", *req.Filters.Until)

		// --- Parse filters only when not using cursor
		if v := r.URL.Query().Get("length"); v != "" {
			if i, err := strconv.Atoi(v); err == nil {
				req.Length = &i
				ctxLogger.Int("length", i)
			}
		}
		if v := r.URL.Query().Get("start"); v != "" {
			if i, err := strconv.Atoi(v); err == nil {
				req.Start = &i
				ctxLogger.Int("start", i)
			}
		}
		if v := r.URL.Query().Get("domain"); v != "" {
			ctxLogger.Str("domain", v)
			req.Filters.Domain = &v
		}
		if v := r.URL.Query().Get("client_ip"); v != "" {
			ctxLogger.Str("client_ip", v)
			req.Filters.ClientIP = &v
		}
		if v := r.URL.Query().Get("client_name"); v != "" {
			ctxLogger.Str("client_name", v)
			req.Filters.ClientName = &v
		}
		if v := r.URL.Query().Get("upstream"); v != "" {
			ctxLogger.Str("upstream", v)
			req.Filters.Upstream = &v
		}
		if v := r.URL.Query().Get("type"); v != "" {
			ctxLogger.Str("type", v)
			req.Filters.Type = &v
		}
		if v := r.URL.Query().Get("status"); v != "" {
			ctxLogger.Str("status", v)
			req.Filters.Status = &v
		}
		if v := r.URL.Query().Get("reply"); v != "" {
			ctxLogger.Str("reply", v)
			req.Filters.Reply = &v
		}
		if v := r.URL.Query().Get("dnssec"); v != "" {
			ctxLogger.Str("dnssec", v)
			req.Filters.DNSSEC = &v
		}
		if v := r.URL.Query().Get("disk"); v != "" {
			b, err := strconv.ParseBool(v)
			if err == nil {
				ctxLogger.Bool("disk", b)
				req.Filters.Disk = &b
			}
		}
	}

	logger := ctxLogger.Logger()
	logger.Debug().Msg("fetching query logs")

	// --- Call cluster client
	res, err := h.cluster.FetchQueryLogs(r.Context(), req)
	if err != nil {
		writeJSONError(w, err.Error(), http.StatusBadRequest)
		return
	}

	for _, nr := range res.Results {
		if nr.Error != nil {
			h.logger.Warn().Err(nr.Error).Msg("partial failure fetching logs")
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(res); err != nil {
		h.logger.Error().Err(err).Msg("failed to encode response")
		writeJSONError(w, "failed to encode response", http.StatusInternalServerError)
	}
}

func parseDomainPath(parts []string) (typeParam *pihole.RuleType, kindParam *pihole.RuleKind, domainParam *string) {
	switch len(parts) {
	case 0:
		// nothing: get all domains
		return nil, nil, nil

	case 1:
		// 1-part combos:
		p := parts[0]
		if rt, ok := pihole.ParseRuleType(p); ok {
			typeParam = &rt
			return
		}
		if rk, ok := pihole.ParseRuleKind(p); ok {
			kindParam = &rk
			return
		}
		domainParam = &p
		return

	case 2:
		// 2-part combos:
		p1, p2 := parts[0], parts[1]
		if rt, ok := pihole.ParseRuleType(p1); ok {
			typeParam = &rt
			if rk, ok := pihole.ParseRuleKind(p2); ok {
				kindParam = &rk
			} else {
				domainParam = &p2
			}
		} else if rk, ok := pihole.ParseRuleKind(p1); ok {
			kindParam = &rk
			domainParam = &p2
		} else {
			// fallback: treat first as domain, second ignored (shouldn't happen in spec)
			domainParam = &p1
		}

	case 3:
		// 3-part combo: /allow|deny/exact|regex/domain
		p1, p2, p3 := parts[0], parts[1], parts[2]
		if rt, ok := pihole.ParseRuleType(p1); ok {
			typeParam = &rt
		}
		if rk, ok := pihole.ParseRuleKind(p2); ok {
			kindParam = &rk
		}
		domainParam = &p3
	}
	return
}

func (h *Handler) GetDomainRules(w http.ResponseWriter, r *http.Request) {
	// Path after /api/domains
	suffix := strings.TrimPrefix(r.URL.Path, "/api/domains")
	suffix = strings.TrimPrefix(suffix, "/")
	parts := []string{}
	if suffix != "" {
		parts = strings.Split(suffix, "/")
	}

	typeParam, kindParam, domainParam := parseDomainPath(parts)
	ctxLogger := h.logger.With()
	if typeParam != nil {
		ctxLogger.Str("type", string(*typeParam))
	}
	if kindParam != nil {
		ctxLogger.Str("kind", string(*kindParam))
	}
	if domainParam != nil {
		ctxLogger.Str("domain", *domainParam)
	}
	logger := ctxLogger.Logger()
	logger.Debug().Msg("getting domain rules")

	opts := pihole.GetDomainRulesOptions{
		Type:   typeParam,
		Kind:   kindParam,
		Domain: domainParam,
	}
	results := h.cluster.GetDomainRules(r.Context(), opts)

	for _, nr := range results {
		if nr.Error != nil {
			logger.Warn().Err(nr.Error).Msg("partial failure getting domain rule")
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(results); err != nil {
		logger.Error().Err(err).Msg("failed to encode response")
		writeJSONError(w, "failed to encode response", http.StatusInternalServerError)
	}
}

func (h *Handler) AddDomainRule(w http.ResponseWriter, r *http.Request) {
	domainType := chi.URLParam(r, "type")
	domainKind := chi.URLParam(r, "kind")
	logger := h.logger.With().Str("type", domainType).Str("kind", domainKind).Logger()

	// --- Parse JSON body
	var body pihole.AddDomainPayload
	if err := decodeJSONBody(w, r, &body, 1<<20); err != nil {
		logger.Error().Err(err).Msg("invalid JSON body")
		writeJSONError(w, "invalid JSON body", http.StatusBadRequest)
		return
	}

	// --- Normalize Domain to []string
	var domains []string
	switch v := body.Domain.(type) {
	case string:
		domains = []string{v}
	case []interface{}:
		for _, item := range v {
			s, ok := item.(string)
			if !ok {
				logger.Error().Interface("domain", item).Msg("domain list must conain only strings")
				writeJSONError(w, "domain list must contain only strings", http.StatusBadRequest)
				return
			}
			domains = append(domains, s)
		}
	default:
		logger.Error().Interface("domain", v).Msg("domain must be string or array of strings")
		writeJSONError(w, "domain must be string or array of strings", http.StatusBadRequest)
		return
	}

	body.Domain = domains

	logger.Debug().Strs("domains", domains).Msg("adding domain rule")

	opts := pihole.AddDomainRuleOptions{
		Type:    domainType,
		Kind:    domainKind,
		Payload: body,
	}
	results := h.cluster.AddDomainRule(r.Context(), opts)

	for _, nr := range results {
		if nr.Error != nil {
			logger.Warn().Err(nr.Error).Msg("partial failure adding domain rule")
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(results); err != nil {
		logger.Error().Err(err).Msg("failed to encode response")
		writeJSONError(w, "failed to encode response", http.StatusInternalServerError)
	}
}

func (h *Handler) RemoveDomainRule(w http.ResponseWriter, r *http.Request) {
	domainType := chi.URLParam(r, "type")
	domainKind := chi.URLParam(r, "kind")
	domain := chi.URLParam(r, "domain")

	logger := h.logger.With().Str("type", domainType).Str("kind", domainKind).Str("domain", domain).Logger()
	logger.Debug().Msg("removing domain rule")

	opts := pihole.RemoveDomainRuleOptions{
		Type:   domainType,
		Kind:   domainKind,
		Domain: domain,
	}
	results := h.cluster.RemoveDomainRule(r.Context(), opts)

	for _, nr := range results {
		if nr.Error != nil {
			logger.Warn().Err(nr.Error).Msg("partial failure removing domain rule")
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(results); err != nil {
		logger.Error().Err(err).Msg("failed to encode response")
		writeJSONError(w, "failed to encode response", http.StatusInternalServerError)
	}
}
