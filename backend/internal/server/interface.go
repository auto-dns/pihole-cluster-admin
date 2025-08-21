package server

import (
	"net/http"

	"github.com/go-chi/chi"
)

type handler interface {
	// Handler
	AuthMiddleware(next http.Handler) http.Handler
	// Routes
	// -- Unauthenticated
	Healthcheck(w http.ResponseWriter, r *http.Request)
	Login(w http.ResponseWriter, r *http.Request)
	Logout(w http.ResponseWriter, r *http.Request)
	GetIsInitialized(w http.ResponseWriter, r *http.Request)
	// -- Authenticated
	// ---- Setup status
	GetInitializationStatus(w http.ResponseWriter, r *http.Request)
	UpdatePiholeInitializationStatus(w http.ResponseWriter, r *http.Request)
	// ---- Event Streaming
	HandleEvents(w http.ResponseWriter, r *http.Request)
	// ---- Health Status
	GetHealthSummary(w http.ResponseWriter, r *http.Request)
	GetNodeHealth(w http.ResponseWriter, r *http.Request)
	// ---- User
	GetSessionUser(w http.ResponseWriter, r *http.Request)
	PatchUser(w http.ResponseWriter, r *http.Request)
	UpdateUserPassword(w http.ResponseWriter, r *http.Request)
	// ---- Pihole CRUD
	AddPiholeNode(w http.ResponseWriter, r *http.Request)
	UpdatePiholeNode(w http.ResponseWriter, r *http.Request)
	RemovePiholeNode(w http.ResponseWriter, r *http.Request)
	GetAllPiholeNodes(w http.ResponseWriter, r *http.Request)
	TestExistingPiholeConnection(w http.ResponseWriter, r *http.Request)
	TestPiholeInstanceConnection(w http.ResponseWriter, r *http.Request)
	// ---- User CRUD
	CreateUser(w http.ResponseWriter, r *http.Request)
	// ---- Application business logic
	FetchQueryLogs(w http.ResponseWriter, r *http.Request)
	GetDomainRules(w http.ResponseWriter, r *http.Request)
	AddDomainRule(w http.ResponseWriter, r *http.Request)
	RemoveDomainRule(w http.ResponseWriter, r *http.Request)
}

type routeRegistrar interface {
	Routes() chi.Router
}
