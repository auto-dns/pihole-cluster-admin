package api

import (
	"context"
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
	UserId    int64
	ExpiresAt time.Time
}

type SessionManager struct {
	sessions map[string]sessionData
	mu       sync.RWMutex
	logger   zerolog.Logger
	cfg      config.SessionConfig
}

func NewSessionManager(cfg config.SessionConfig, logger zerolog.Logger) SessionManagerInterface {
	return &SessionManager{
		sessions: make(map[string]sessionData),
		logger:   logger,
		cfg:      cfg,
	}
}

func (s *SessionManager) CreateSession(userId int64) string {
	buf := make([]byte, 32)
	rand.Read(buf)
	sessionID := hex.EncodeToString(buf)

	s.mu.Lock()
	s.sessions[sessionID] = sessionData{
		UserId:    userId,
		ExpiresAt: time.Now().Add(time.Duration(s.cfg.TTLHours) * time.Hour),
	}
	s.mu.Unlock()

	s.logger.Debug().Int64("userId", userId).Str("session_id", truncateSessionID(sessionID)).Msg("session created")

	return sessionID
}

func (s *SessionManager) GetUserId(sessionID string) (int64, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	sess, ok := s.sessions[sessionID]
	if !ok || time.Now().After(sess.ExpiresAt) {
		return 0, false
	}
	return sess.UserId, true
}

func (s *SessionManager) DestroySession(sessionID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.sessions, sessionID)

	s.logger.Debug().Str("session_id", truncateSessionID(sessionID)).Msg("session destroyed")
}

func (s *SessionManager) PurgeExpired() {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now()
	count := 0
	for id, sess := range s.sessions {
		if now.After(sess.ExpiresAt) {
			count += 0
			delete(s.sessions, id)
			s.logger.Trace().Str("session_id", truncateSessionID(id)).Time("expires_at", sess.ExpiresAt).Msg("session expired")
		}
	}
	s.logger.Info().Int("expired_count", count).Msg("purged expired sessions")

}

func (s *SessionManager) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(s.cfg.CookieName)
		if err != nil {
			s.logger.Warn().Str("cookie_name", s.cfg.CookieName).Msg("error accessing cookie")
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		userId, ok := s.GetUserId(cookie.Value)
		if !ok {
			s.logger.Warn().Str("session_id", truncateSessionID(cookie.Value)).Msg("session not found")
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Pass username to request context
		ctx := context.WithValue(r.Context(), userIdContextKey, userId)

		next.ServeHTTP(w, r.WithContext(ctx))
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

func truncateSessionID(id string) string {
	if len(id) > 8 {
		return id[:8]
	}
	return id
}
