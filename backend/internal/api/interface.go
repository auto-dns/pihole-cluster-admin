package api

import "net/http"

type HandlerInterface interface {
	Healthcheck(w http.ResponseWriter, r *http.Request)
	FetchQueryLogs(w http.ResponseWriter, r *http.Request)
	HandleGetDomainRules(w http.ResponseWriter, r *http.Request)
	HandleAddDomainRule(w http.ResponseWriter, r *http.Request)
	HandleRemoveDomainRule(w http.ResponseWriter, r *http.Request)
}
