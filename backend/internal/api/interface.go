package api

import "net/http"

type HandlerInterface interface {
	Healthcheck(w http.ResponseWriter, r *http.Request)
}
