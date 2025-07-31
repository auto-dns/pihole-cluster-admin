package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/auto-dns/pihole-cluster-admin/internal/pihole"
	"github.com/auto-dns/pihole-cluster-admin/internal/store"
	"github.com/go-chi/chi"
	"github.com/rs/zerolog"
)

func ptrInt64(v int64) *int64 { return &v }

type Handler struct {
	cluster     pihole.ClusterInterface
	sessions    SessionInterface
	piholeStore store.PiholeStoreInterface
	userStore   store.UserStoreInterface
	logger      zerolog.Logger
}

func NewHandler(cluster pihole.ClusterInterface, sessions SessionInterface, piholeStore store.PiholeStoreInterface, userStore store.UserStoreInterface, logger zerolog.Logger) *Handler {
	return &Handler{
		cluster:     cluster,
		sessions:    sessions,
		piholeStore: piholeStore,
		userStore:   userStore,
		logger:      logger,
	}
}

func (h *Handler) AuthMiddleware(next http.Handler) http.Handler {
	return h.sessions.AuthMiddleware(next)
}

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
	valid, err := h.userStore.ValidateUser(creds.Username, creds.Password)
	if err != nil {
		h.logger.Error().Err(err).Str("username", creds.Username).Msg("error validating user")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	if !valid {
		h.logger.Warn().Str("username", creds.Username).Msg("invalid login attempt")
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// Successful login → create session
	h.logger.Info().Str("username", creds.Username).Msg("user login success")
	sessionID := h.sessions.CreateSession(creds.Username)
	http.SetCookie(w, h.sessions.Cookie(sessionID))
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session_id")
	if err == nil {
		sessionId := cookie.Value
		if username, ok := h.sessions.GetUsername(sessionId); ok {
			h.logger.Info().Str("username", username).Msg("user logged out")
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

func (h *Handler) GetInitializationStatus(w http.ResponseWriter, r *http.Request) {
	initialized, err := h.userStore.IsInitialized()
	if err != nil {
		h.logger.Error().Err(err).Msg("failed to get app initialization status")
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(map[string]bool{"initialized": initialized})
}

// Pihole CRUD routes

func (h *Handler) AddPiholeNode(w http.ResponseWriter, r *http.Request) {
	type AddPiholeBody struct {
		Scheme      string `json:"scheme"`
		Host        string `json:"host"`
		Port        int    `json:"port"`
		Description string `json:"description"`
		Password    string `json:"password"`
	}

	// Parse inputs from body
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
	if strings.TrimSpace(addPiholeBody.Password) == "" {
		h.logger.Error().Msg("password must not be empty")
		http.Error(w, "password must not be empty", http.StatusBadRequest)
		return
	}

	var description *string
	if strings.TrimSpace(addPiholeBody.Description) != "" {
		description = &addPiholeBody.Description
	}

	// Call user store to add the node
	piholeNode := store.PiholeNode{
		Scheme:      addPiholeBody.Scheme,
		Host:        addPiholeBody.Host,
		Port:        addPiholeBody.Port,
		Description: description,
		Password:    addPiholeBody.Password,
	}

	h.logger.Debug().Str("scheme", piholeNode.Scheme).Str("host", piholeNode.Host).Int("port", piholeNode.Port).Msg("adding pihole node")

	err := h.piholeStore.AddPiholeNode(piholeNode)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			http.Error(w, "duplicate host", http.StatusConflict)
			return
		}
		h.logger.Error().Err(err).Str("host", piholeNode.Host).Int("port", piholeNode.Port).Msg("failed to add pihole node")
		http.Error(w, "failed to add pihole node", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "pihole node added successfully",
	})
}

func (h *Handler) UpdatePiholeNode(w http.ResponseWriter, r *http.Request) {

}

func (h *Handler) RemovePiholeNode(w http.ResponseWriter, r *http.Request) {

}

func (h *Handler) GetAllPiholeNodes(w http.ResponseWriter, r *http.Request) {

}

// User CRUD routes

func (h *Handler) CreateUser(w http.ResponseWriter, r *http.Request) {

}

func (h *Handler) DeleteUser(w http.ResponseWriter, r *http.Request) {

}

// Application business logic routes

func (h *Handler) FetchQueryLogs(w http.ResponseWriter, r *http.Request) {
	var req pihole.FetchQueryLogRequest
	ctxLogger := h.logger.With()

	cursor := r.URL.Query().Get("cursor")
	if cursor != "" {
		// Cursor request: only cursor and optional length override
		req.CursorID = &cursor
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
