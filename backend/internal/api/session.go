package api

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/auto-dns/pihole-cluster-admin/internal/config"
	"github.com/rs/zerolog"
)

type sessionData struct {
	Username  string
	ExpiresAt time.Time
}

type SessionManager struct {
	sessions map[string]sessionData
	mu       sync.RWMutex
	logger   zerolog.Logger
	cfg      config.SessionConfig
}

func NewSessionManager(cfg config.SessionConfig, logger zerolog.Logger) *SessionManager {
	return &SessionManager{
		sessions: make(map[string]sessionData),
		logger:   logger,
		cfg:      cfg,
	}
}

func (s *SessionManager) CreateSession(username string) string {
	buf := make([]byte, 32)
	rand.Read(buf)
	sessionID := hex.EncodeToString(buf)

	s.mu.Lock()
	s.sessions[sessionID] = sessionData{
		Username:  username,
		ExpiresAt: time.Now().Add(time.Duration(s.cfg.TTLHours) * time.Hour),
	}
	s.mu.Unlock()

	return sessionID
}

func (s *SessionManager) GetUsername(sessionID string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	sess, ok := s.sessions[sessionID]
	if !ok || time.Now().After(sess.ExpiresAt) {
		return "", false
	}
	return sess.Username, true
}

func (s *SessionManager) DestroySession(sessionID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.sessions, sessionID)
}

func (s *SessionManager) PurgeExpired() {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now()
	for id, sess := range s.sessions {
		if now.After(sess.ExpiresAt) {
			delete(s.sessions, id)
		}
	}
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

func parseSameSite(val string) http.SameSite {
	switch strings.ToLower(val) {
	case "lax":
		return http.SameSiteLaxMode
	case "none":
		return http.SameSiteNoneMode
	default:
		return http.SameSiteStrictMode
	}
}

func (s *SessionManager) Cookie(value string) *http.Cookie {
	ttl := time.Duration(s.cfg.TTLHours) * time.Hour
	secure := s.cfg.Secure && !s.cfg.AllowInsecureCookie
	sameSite := parseSameSite(s.cfg.SameSite)
	return &http.Cookie{
		Name:     s.cfg.CookieName,
		Value:    value,
		Path:     s.cfg.CookiePath,
		HttpOnly: true,
		Secure:   secure,
		SameSite: sameSite,
		Expires:  time.Now().Add(ttl),
	}
}
