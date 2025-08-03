package api

import "net/http"

type HandlerInterface interface {
	// Handler
	AuthMiddleware(next http.Handler) http.Handler
	// Routes
	// -- Unauthenticated
	Healthcheck(w http.ResponseWriter, r *http.Request)
	Login(w http.ResponseWriter, r *http.Request)
	Logout(w http.ResponseWriter, r *http.Request)
	GetInitializationStatus(w http.ResponseWriter, r *http.Request)
	// -- Authenticated
	// ---- User
	GetSessionUser(w http.ResponseWriter, r *http.Request)
	// ---- Pihole CRUD
	AddPiholeNode(w http.ResponseWriter, r *http.Request)
	UpdatePiholeNode(w http.ResponseWriter, r *http.Request)
	RemovePiholeNode(w http.ResponseWriter, r *http.Request)
	GetAllPiholeNodes(w http.ResponseWriter, r *http.Request)
	// ---- User CRUD
	CreateUser(w http.ResponseWriter, r *http.Request)
	// ---- Application business logic
	FetchQueryLogs(w http.ResponseWriter, r *http.Request)
	GetDomainRules(w http.ResponseWriter, r *http.Request)
	AddDomainRule(w http.ResponseWriter, r *http.Request)
	RemoveDomainRule(w http.ResponseWriter, r *http.Request)
}

type SessionManagerInterface interface {
	CreateSession(userId int64) string
	GetUserId(sessionID string) (int64, bool)
	DestroySession(sessionID string)
	AuthMiddleware(next http.Handler) http.Handler
	Cookie(value string) *http.Cookie
	PurgeExpired()
}

type ContextKey string

const userIdContextKey ContextKey = "userId"
