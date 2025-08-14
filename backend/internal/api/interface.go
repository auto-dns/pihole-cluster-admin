package api

import (
	"net/http"

	"github.com/auto-dns/pihole-cluster-admin/internal/health"
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

type healthService interface {
	NodeHealth() map[int64]health.NodeHealth
	Summary() health.Summary
}
