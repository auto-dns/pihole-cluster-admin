package api

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"sync"
	"time"

	"github.com/rs/zerolog"
)

type SessionManager struct {
	sessions map[string]string
	mu       sync.RWMutex
	secure   bool
	logger   zerolog.Logger
}

func NewSessionManager(secure bool, logger zerolog.Logger) *SessionManager {
	return &SessionManager{
		sessions: make(map[string]string),
		secure:   secure,
		logger:   logger,
	}
}

func (s *SessionManager) CreateSession(username string) string {
	buf := make([]byte, 32)
	rand.Read(buf)
	sessionID := hex.EncodeToString(buf)

	s.mu.Lock()
	s.sessions[sessionID] = username
	s.mu.Unlock()

	return sessionID
}

func (s *SessionManager) GetUsername(sessionID string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	username, ok := s.sessions[sessionID]
	return username, ok
}

func (s *SessionManager) DestroySession(sessionID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.sessions, sessionID)
}

func (s *SessionManager) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("session_id")
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		if _, ok := s.GetUsername(cookie.Value); !ok {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (s *SessionManager) Cookie(name, value string) *http.Cookie {
	return &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     "/",
		HttpOnly: true,
		Secure:   s.secure,
		SameSite: http.SameSiteStrictMode,
		Expires:  time.Now().Add(24 * time.Hour),
	}
}
