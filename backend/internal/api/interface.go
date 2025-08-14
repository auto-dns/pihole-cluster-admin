package api

import (
	"net/http"
)

type sessionAuth interface {
	AuthMiddleware(next http.Handler) http.Handler
}

type sessionIssuer interface {
	CreateSession(userId int64) (string, error)
	GetUserId(sessionID string) (int64, bool, error)
	DestroySession(sessionID string) error
	Cookie(value string) *http.Cookie
}

type sessionDeps interface {
	sessionAuth
	sessionIssuer
}
