package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/auto-dns/pihole-cluster-admin/internal/config"
	"github.com/auto-dns/pihole-cluster-admin/internal/pihole"
	"github.com/auto-dns/pihole-cluster-admin/internal/store"
	"github.com/go-chi/chi"
	"github.com/rs/zerolog"
)

func ptrInt64(v int64) *int64 { return &v }

type Handler struct {
	cluster                   pihole.ClusterInterface
	sessions                  SessionManagerInterface
	initializationStatusStore store.InitializationStatusStoreInterface
	piholeStore               store.PiholeStoreInterface
	userStore                 store.UserStoreInterface
	logger                    zerolog.Logger
	sessionCfg                config.SessionConfig
}

func NewHandler(cluster pihole.ClusterInterface, sessions SessionManagerInterface, initializationStatusStore store.InitializationStatusStoreInterface, piholeStore store.PiholeStoreInterface, userStore store.UserStoreInterface, cfg config.SessionConfig, logger zerolog.Logger) HandlerInterface {
	return &Handler{
		cluster:                   cluster,
		sessions:                  sessions,
		initializationStatusStore: initializationStatusStore,
		piholeStore:               piholeStore,
		userStore:                 userStore,
		logger:                    logger,
	}
}

// Handlers

func (h *Handler) AuthMiddleware(next http.Handler) http.Handler {
	return h.sessions.AuthMiddleware(next)
}

// Unauthenticated routes

func (h *Handler) Healthcheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"status": "OK"}`))
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var creds struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Validate against the database
	user, err := h.userStore.ValidateUser(creds.Username, creds.Password)
	var wrongPasswordErr *store.WrongPasswordError
	switch {
	case errors.Is(err, sql.ErrNoRows):
		h.logger.Warn().Str("username", creds.Username).Msg("invalid login attempt")
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	case errors.As(err, &wrongPasswordErr):
		h.logger.Warn().Str("username", creds.Username).Msg("invalid login attempt")
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	case err != nil:
		h.logger.Error().Err(err).Str("username", creds.Username).Msg("error validating user")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Successful login → create session
	h.logger.Info().Int64("userId", user.Id).Msg("user login success")
	sessionID := h.sessions.CreateSession(user.Id)
	http.SetCookie(w, h.sessions.Cookie(sessionID))
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(h.sessionCfg.CookieName)
	if err == nil {
		sessionId := cookie.Value
		if userId, ok := h.sessions.GetUserId(sessionId); ok {
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
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(map[string]bool{"initialized": initialized})
}

func (h *Handler) GetInitializationStatus(w http.ResponseWriter, r *http.Request) {
	initializationStatus, err := h.initializationStatusStore.GetInitializationStatus()
	if err != nil {
		h.logger.Error().Err(err).Msg("failed to get app initialization status")
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(initializationStatus)
}

// Authenticated routes
// -- User

func (h *Handler) GetSessionUser(w http.ResponseWriter, r *http.Request) {
	userId := r.Context().Value(userIdContextKey).(int64)

	user, err := h.userStore.GetUser(userId)
	if err != nil {
		h.logger.Error().Err(err).Msg("user session not found")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
	}

	h.logger.Debug().Int64("id", user.Id).Str("username", user.Username).Msg("user fetched from database")

	userResponse := UserResponse{
		Username:  user.Username,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}
	// Create a session and return a cookie
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
	type AddPiholeBody struct {
		Scheme      string `json:"scheme"`
		Host        string `json:"host"`
		Port        int    `json:"port"`
		Name        string `json:"name"`
		Description string `json:"description"`
		Password    string `json:"password"`
	}
	var addPiholeBody AddPiholeBody
	if err := json.NewDecoder(r.Body).Decode(&addPiholeBody); err != nil {
		h.logger.Error().Err(err).Msg("invalid JSON body")
		http.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}

	// Validate the inputs
	if addPiholeBody.Scheme != "http" && addPiholeBody.Scheme != "https" {
		h.logger.Error().Msg("scheme must be http or https")
		http.Error(w, "scheme must be http or https", http.StatusBadRequest)
		return
	}
	if strings.TrimSpace(addPiholeBody.Host) == "" {
		h.logger.Error().Msg("host must not be empty")
		http.Error(w, "host must not be empty", http.StatusBadRequest)
		return
	}
	if addPiholeBody.Port <= 0 || addPiholeBody.Port > 65535 {
		h.logger.Error().Msg("port must be a valid TCP port")
		http.Error(w, "port must be a valid TCP port", http.StatusBadRequest)
		return
	}
	if strings.TrimSpace(addPiholeBody.Name) == "" {
		h.logger.Error().Msg("name must not be empty")
		http.Error(w, "name must not be empty", http.StatusBadRequest)
		return
	}
	if strings.TrimSpace(addPiholeBody.Password) == "" {
		h.logger.Error().Msg("password must not be empty")
		http.Error(w, "password must not be empty", http.StatusBadRequest)
		return
	}

	// Call user store to add the node
	addParams := store.AddPiholeParams{
		Scheme:      addPiholeBody.Scheme,
		Host:        addPiholeBody.Host,
		Port:        addPiholeBody.Port,
		Name:        addPiholeBody.Name,
		Description: addPiholeBody.Description,
		Password:    addPiholeBody.Password,
	}

	insertedNode, err := h.piholeStore.AddPiholeNode(addParams)
	if err != nil {
		if strings.Contains(err.Error(), "piholes.host") {
			h.logger.Error().Err(err).Str("host", addParams.Host).Int("port", addParams.Port).Msg("duplicate host:port")
			http.Error(w, "duplicate host:port", http.StatusConflict)
			return
		}
		if strings.Contains(err.Error(), "piholes.name") {
			h.logger.Error().Err(err).Str("name", addParams.Name).Msg("duplicate name")
			http.Error(w, "duplicate name", http.StatusConflict)
			return
		}
		// generic fallback
		h.logger.Error().Err(err).Str("scheme", addParams.Scheme).Str("name", addParams.Name).Str("host", addParams.Host).Int("port", addParams.Port).Str("description", addParams.Description).Msg("error adding pihole")
		http.Error(w, "failed to add pihole node", http.StatusInternalServerError)
		return
	}

	h.logger.Debug().Int64("id", insertedNode.Id).Str("scheme", insertedNode.Scheme).Str("host", insertedNode.Host).Int("port", insertedNode.Port).Str("name", insertedNode.Name).Time("created_at", insertedNode.CreatedAt).Time("updated_at", insertedNode.UpdatedAt).Msg("added pihole node to database")

	// Add client to cluster
	cfg := &pihole.ClientConfig{
		Id:       insertedNode.Id,
		Name:     insertedNode.Name,
		Scheme:   insertedNode.Scheme,
		Host:     insertedNode.Host,
		Port:     insertedNode.Port,
		Password: *insertedNode.Password,
	}
	client := pihole.NewClient(cfg, h.logger)
	err = h.cluster.AddClient(client)
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
	type UpdatePiholeBody struct {
		Scheme      *string `json:"scheme"`
		Host        *string `json:"host"`
		Port        *int    `json:"port"`
		Name        *string `json:"name"`
		Description *string `json:"description"`
		Password    *string `json:"password"` // Updating this field is optional
	}
	var updatePiholeBody UpdatePiholeBody
	if err := json.NewDecoder(r.Body).Decode(&updatePiholeBody); err != nil {
		h.logger.Error().Err(err).Msg("invalid JSON body")
		http.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}

	// Validate request
	idString := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idString, 10, 64)
	if err != nil {
		h.logger.Error().Err(err).Msg("error converting path parameter id to int64")
		http.Error(w, "error processing id path parameter", http.StatusBadRequest)
		return
	}
	if id <= 0 {
		h.logger.Error().Msg("invalid id (<= 0)")
		http.Error(w, "invalid id (<= 0)", http.StatusBadRequest)
		return
	}
	// Validate at least one update field set
	if updatePiholeBody.Scheme == nil && updatePiholeBody.Host == nil && updatePiholeBody.Port == nil && updatePiholeBody.Name == nil && updatePiholeBody.Description == nil && updatePiholeBody.Password == nil {
		h.logger.Error().Msg("must provide at least one field to update")
		http.Error(w, "must provide at least one field to update", http.StatusBadRequest)
		return
	}
	// Validate content
	if updatePiholeBody.Scheme != nil && *updatePiholeBody.Scheme != "http" && *updatePiholeBody.Scheme != "https" {
		h.logger.Error().Msg("scheme must be http or https")
		http.Error(w, "scheme must be http or https", http.StatusBadRequest)
		return
	}
	if updatePiholeBody.Host != nil && strings.TrimSpace(*updatePiholeBody.Host) == "" {
		h.logger.Error().Msg("host must not be empty")
		http.Error(w, "host must not be empty", http.StatusBadRequest)
		return
	}
	if updatePiholeBody.Port != nil && *updatePiholeBody.Port <= 0 || *updatePiholeBody.Port > 65535 {
		h.logger.Error().Msg("port must be a valid TCP port")
		http.Error(w, "port must be a valid TCP port", http.StatusBadRequest)
		return
	}
	if updatePiholeBody.Name != nil && strings.TrimSpace(*updatePiholeBody.Name) == "" {
		h.logger.Error().Msg("name must not be empty")
		http.Error(w, "name must not be empty", http.StatusBadRequest)
		return
	}
	if updatePiholeBody.Password != nil && strings.TrimSpace(*updatePiholeBody.Password) == "" {
		h.logger.Error().Msg("password must not be empty")
		http.Error(w, "password must not be empty", http.StatusBadRequest)
	}

	// Call user store to update the node
	updateParams := store.UpdatePiholeParams{
		Scheme:      updatePiholeBody.Scheme,
		Host:        updatePiholeBody.Host,
		Port:        updatePiholeBody.Port,
		Name:        updatePiholeBody.Name,
		Description: updatePiholeBody.Description,
		Password:    updatePiholeBody.Password,
	}

	updatedNode, err := h.piholeStore.UpdatePiholeNode(id, updateParams)
	if err != nil {
		if strings.Contains(err.Error(), "piholes.host") {
			h.logger.Error().Err(err).Str("host", *updateParams.Host).Int("port", *updateParams.Port).Msg("duplicate host:port")
			http.Error(w, "duplicate host:port", http.StatusConflict)
			return
		}
		if strings.Contains(err.Error(), "piholes.name") {
			h.logger.Error().Err(err).Str("name", *updateParams.Name).Msg("duplicate name")
			http.Error(w, "duplicate name", http.StatusConflict)
			return
		}
		// generic fallback
		h.logger.Error().Err(err).Str("scheme", *updateParams.Scheme).Str("name", *updateParams.Name).Str("host", *updateParams.Host).Int("port", *updateParams.Port).Str("description", *updateParams.Description).Msg("error adding pihole")
		http.Error(w, "failed to update pihole node", http.StatusInternalServerError)
		return
	}

	h.logger.Debug().Int64("id", updatedNode.Id).Str("scheme", updatedNode.Scheme).Str("host", updatedNode.Host).Int("port", updatedNode.Port).Time("created_at", updatedNode.CreatedAt).Str("name", updatedNode.Name).Time("updated_at", updatedNode.UpdatedAt).Msg("updated pihole node")

	// Update client in cluster
	cfg := &pihole.ClientConfig{
		Id:       updatedNode.Id,
		Name:     updatedNode.Name,
		Scheme:   updatedNode.Scheme,
		Host:     updatedNode.Host,
		Port:     updatedNode.Port,
		Password: *updatedNode.Password,
	}
	err = h.cluster.UpdateClient(cfg.Id, cfg)
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
		http.Error(w, "error processing id path parameter", http.StatusBadRequest)
		return
	}
	if id <= 0 {
		h.logger.Error().Msg("invalid id (<= 0)")
		http.Error(w, "invalid id (<= 0)", http.StatusBadRequest)
		return
	}

	found, err := h.piholeStore.RemovePiholeNode(id)
	if err != nil {
		h.logger.Error().Err(err).Int64("id", id).Msg("error removing pihole node from database")
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	if h.cluster.HasClient(id) {
		err = h.cluster.RemoveClient(id)
		if err != nil {
			h.logger.Error().Err(err).Int64("id", id).Msg("removing client from cluster")
			http.Error(w, "internal server error", http.StatusInternalServerError)
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
		http.Error(w, "internal server error", http.StatusInternalServerError)
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
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	if initialized {
		h.logger.Error().Msg("app is already initialized")
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	// Parse request body
	type CreateUserBody struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	var createUserBody CreateUserBody
	err = json.NewDecoder(r.Body).Decode(&createUserBody)
	if err != nil {
		h.logger.Error().Err(err).Msg("invalid JSON body")
		http.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}

	// Validate request params
	if strings.TrimSpace(createUserBody.Username) == "" {
		h.logger.Error().Msg("empty username in body")
		http.Error(w, "empty username in body", http.StatusBadRequest)
		return
	}
	if strings.TrimSpace(createUserBody.Password) == "" {
		h.logger.Error().Msg("empty password in body")
		http.Error(w, "empty password in body", http.StatusBadRequest)
		return
	}

	// Create user
	createUserParams := store.CreateUserParams{
		Username: createUserBody.Username,
		Password: createUserBody.Password,
	}
	user, err := h.userStore.CreateUser(createUserParams)
	if err != nil {
		h.logger.Error().Err(err).Str("username", user.Username).Msg("failed to create user")
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	err = h.initializationStatusStore.SetUserCreated(true)
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
	sessionID := h.sessions.CreateSession(user.Id)
	http.SetCookie(w, h.sessions.Cookie(sessionID))
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
		ctxLogger.Str("cursor", cursor).Int("length", *req.Length)
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
					http.Error(w, "invalid 'from' time", http.StatusBadRequest)
					return
				}
				req.Filters.From = ptrInt64(fromTime.Unix())
			}
			if untilStr != "" {
				untilTime, err := time.Parse(time.RFC3339, untilStr)
				if err != nil {
					http.Error(w, "invalid 'until' time", http.StatusBadRequest)
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
	res, err := h.cluster.FetchQueryLogs(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	for _, nr := range res.Results {
		if nr.Error != "" {
			h.logger.Warn().Str("error", nr.Error).Msg("partial failure fetching logs")
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(res); err != nil {
		h.logger.Error().Err(err).Msg("failed to encode response")
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}

func parseDomainPath(parts []string) (typeParam, kindParam, domainParam *string) {
	switch len(parts) {
	case 0:
		// nothing: get all domains
		return nil, nil, nil

	case 1:
		// 1-part combos:
		p := parts[0]
		switch p {
		case "allow", "deny":
			typeParam = &p
		case "exact", "regex":
			kindParam = &p
		default:
			domainParam = &p
		}

	case 2:
		// 2-part combos:
		p1 := parts[0]
		p2 := parts[1]
		if p1 == "allow" || p1 == "deny" {
			typeParam = &p1
			if p2 == "exact" || p2 == "regex" {
				kindParam = &p2
			} else {
				domainParam = &p2
			}
		} else if p1 == "exact" || p1 == "regex" {
			kindParam = &p1
			domainParam = &p2
		} else {
			// fallback: treat first as domain, second ignored (shouldn't happen in spec)
			domainParam = &p1
		}

	case 3:
		// 3-part combo: /allow|deny/exact|regex/domain
		p1 := parts[0]
		p2 := parts[1]
		p3 := parts[2]
		if p1 == "allow" || p1 == "deny" {
			typeParam = &p1
		}
		if p2 == "exact" || p2 == "regex" {
			kindParam = &p2
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
		ctxLogger.Str("type", *typeParam)
	}
	if kindParam != nil {
		ctxLogger.Str("kind", *kindParam)
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
	results := h.cluster.GetDomainRules(opts)

	for _, nr := range results {
		if nr.Error != "" {
			logger.Warn().Str("error", nr.Error).Msg("partial failure getting domain rule")
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(results); err != nil {
		logger.Error().Err(err).Msg("failed to encode response")
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}

func (h *Handler) AddDomainRule(w http.ResponseWriter, r *http.Request) {
	domainType := chi.URLParam(r, "type")
	domainKind := chi.URLParam(r, "kind")
	logger := h.logger.With().Str("type", domainType).Str("kind", domainKind).Logger()

	// --- Parse JSON body
	var payload pihole.AddDomainPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		logger.Error().Err(err).Msg("invalid JSON body")
		http.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}

	// --- Normalize Domain to []string
	var domains []string
	switch v := payload.Domain.(type) {
	case string:
		domains = []string{v}
	case []interface{}:
		for _, item := range v {
			s, ok := item.(string)
			if !ok {
				logger.Error().Interface("domain", item).Msg("domain list must conain only strings")
				http.Error(w, "domain list must contain only strings", http.StatusBadRequest)
				return
			}
			domains = append(domains, s)
		}
	default:
		logger.Error().Interface("domain", v).Msg("domain must be string or array of strings")
		http.Error(w, "domain must be string or array of strings", http.StatusBadRequest)
		return
	}

	payloadNormalized := payload
	payloadNormalized.Domain = domains

	logger.Debug().Strs("domains", domains).Msg("adding domain rule")

	opts := pihole.AddDomainRuleOptions{
		Type:    domainType,
		Kind:    domainKind,
		Payload: payloadNormalized,
	}
	results := h.cluster.AddDomainRule(opts)

	for _, nr := range results {
		if nr.Error != "" {
			logger.Warn().Str("error", nr.Error).Msg("partial failure adding domain rule")
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(results); err != nil {
		logger.Error().Err(err).Msg("failed to encode response")
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
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
	results := h.cluster.RemoveDomainRule(opts)

	for _, nr := range results {
		if nr.Error != "" {
			logger.Warn().Str("error", nr.Error).Msg("partial failure removing domain rule")
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(results); err != nil {
		logger.Error().Err(err).Msg("failed to encode response")
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}
