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
	// -- Authenticated
	FetchQueryLogs(w http.ResponseWriter, r *http.Request)
	GetDomainRules(w http.ResponseWriter, r *http.Request)
	AddDomainRule(w http.ResponseWriter, r *http.Request)
	RemoveDomainRule(w http.ResponseWriter, r *http.Request)
}

type SessionInterface interface {
	CreateSession(username string) string
	GetUsername(sessionID string) (string, bool)
	DestroySession(sessionID string)
	AuthMiddleware(next http.Handler) http.Handler
	Cookie(value string) *http.Cookie
	PurgeExpired()
}
