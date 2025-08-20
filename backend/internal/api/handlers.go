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
	"github.com/auto-dns/pihole-cluster-admin/internal/crypto"
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

	userResponse := UserResponse{
		Id:        user.Id,
		Username:  user.Username,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}

	http.SetCookie(w, h.sessions.Cookie(sessionID))
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(userResponse)
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
		Id:        user.Id,
		Username:  user.Username,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(userResponse)
}

func (h *Handler) PatchUser(w http.ResponseWriter, r *http.Request) {
	// Parse request body
	var body struct {
		Username *string `json:"username"`
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

	currentUserId, ok := r.Context().Value(sessions.UserIdContextKey).(int64)
	if !ok {
		h.logger.Error().Err(err).Msg("error getting current user id from context")
		writeJSONError(w, "internal server error", http.StatusInternalServerError)
		return
	}

	if id != currentUserId {
		h.logger.Error().Err(err).Int64("current_user_id", currentUserId).Int64("id", id).Msg("user tried to upate user id other than own")
		writeJSONError(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	currentUser, err := h.userStore.GetUser(id)
	if err != nil {
		h.logger.Error().Err(err).Msg("error fetching user")
		writeJSONError(w, "internal server error", http.StatusInternalServerError)
		return
	} else if currentUser == nil {
		h.logger.Error().Msg("error fetching user")
		writeJSONError(w, "internal server error", http.StatusInternalServerError)
		return
	}

	// Validate at least one update field set
	if body.Username == nil {
		h.logger.Error().Msg("no fields provided")
		writeJSONError(w, "must provide at least one field to update", http.StatusBadRequest)
		return
	}
	// Validate content
	if body.Username != nil {
		if strings.TrimSpace(*body.Username) == "" {
			h.logger.Error().Msg("username empty")
			writeJSONError(w, "username must not be empty", http.StatusBadRequest)
			return
		}
		if strings.TrimSpace(*body.Username) == strings.TrimSpace(currentUser.Username) {
			h.logger.Error().Msg("new username matched current username")
			writeJSONError(w, "username must not be empty", http.StatusBadRequest)
			return
		}
	}

	// Call user store to update the node
	updateParams := store.UpdateUserParams{
		Username: body.Username,
	}

	updatedNode, err := h.userStore.UpdateUser(id, updateParams)
	safe := func(p *string) string {
		if p == nil {
			return ""
		}
		return *p
	}
	if err != nil {
		// generic fallback
		h.logger.Error().Err(err).Str("username", safe(updateParams.Username)).Msg("error adding user")
		writeJSONError(w, "failed to update user", http.StatusInternalServerError)
		return
	}

	h.logger.Debug().Int64("id", updatedNode.Id).Str("username", updatedNode.Username).Msg("updated user")

	response := UserResponse{
		Id:        updatedNode.Id,
		Username:  updatedNode.Username,
		CreatedAt: updatedNode.CreatedAt,
		UpdatedAt: updatedNode.UpdatedAt,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func (h *Handler) UpdateUserPassword(w http.ResponseWriter, r *http.Request) {
	// Parse request body
	var body struct {
		CurrentPassword string `json:"currentPassword"`
		NewPassword     string `json:"newPassword"`
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

	currentUserId, ok := r.Context().Value(sessions.UserIdContextKey).(int64)
	if !ok {
		h.logger.Error().Err(err).Msg("error getting current user id from context")
		writeJSONError(w, "internal server error", http.StatusInternalServerError)
		return
	}

	if id != currentUserId {
		h.logger.Error().Err(err).Int64("current_user_id", currentUserId).Int64("id", id).Msg("user tried to upate user id other than own")
		writeJSONError(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	currentUserAuth, err := h.userStore.GetUserAuth(id)
	if err != nil {
		h.logger.Error().Err(err).Msg("error fetching password hash")
		writeJSONError(w, "internal server error", http.StatusInternalServerError)
		return
	} else if currentUserAuth == nil {
		h.logger.Error().Msg("error fetching password hash")
		writeJSONError(w, "internal server error", http.StatusInternalServerError)
		return
	}

	// Validate content
	if crypto.CompareHashAndPassword(currentUserAuth.PasswordHash, body.CurrentPassword) != nil {
		h.logger.Error().Msg("current password provided does not match actual current password")
		writeJSONError(w, "current password incorrect", http.StatusUnauthorized)
		return
	}
	if strings.TrimSpace(body.NewPassword) == "" {
		h.logger.Error().Msg("new password empty")
		writeJSONError(w, "new password must not be empty", http.StatusBadRequest)
		return
	}
	if len(strings.TrimSpace(body.NewPassword)) < 8 {
		h.logger.Error().Msg("new password less than 8 characters")
		writeJSONError(w, "new password must be 8 or more characters", http.StatusBadRequest)
		return
	}
	if crypto.CompareHashAndPassword(currentUserAuth.PasswordHash, body.NewPassword) == nil {
		h.logger.Error().Msg("new password matched current password")
		writeJSONError(w, "new password must not match current password", http.StatusBadRequest)
		return
	}

	// Call user store to update the node
	updateParams := store.UpdateUserParams{
		Password: &body.NewPassword,
	}

	updatedNode, err := h.userStore.UpdateUser(id, updateParams)
	if err != nil {
		// generic fallback
		h.logger.Error().Err(err).Int64("id", id).Msg("error updating user password")
		writeJSONError(w, "failed to update user", http.StatusInternalServerError)
		return
	}

	h.logger.Debug().Int64("id", updatedNode.Id).Msg("updated password")

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNoContent)
}

// User CRUD routes

type UserResponse struct {
	Id        int64     `json:"id"`
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
		Id:        user.Id,
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
