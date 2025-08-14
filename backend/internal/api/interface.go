package api

import (
	"context"
	"net/http"
)

type SessionManagerInterface interface {
	CreateSession(userId int64) (string, error)
	GetUserId(sessionID string) (int64, bool, error)
	DestroySession(sessionID string) error
	AuthMiddleware(next http.Handler) http.Handler
	Cookie(value string) *http.Cookie
	StartPurgeLoop(ctx context.Context)
	PurgeExpired()
}

type SessionStorageInterface interface {
	Create(session session) error
	GetAll() ([]session, error)
	GetUserId(sessionId string) (int64, bool, error)
	Delete(sessionId string) error
}
