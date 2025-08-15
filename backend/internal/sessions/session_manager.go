package sessions

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"strings"
	"time"

	"github.com/auto-dns/pihole-cluster-admin/internal/config"
	"github.com/rs/zerolog"
)

type SessionManager struct {
	sessions map[string]Session
	storage  storage
	logger   zerolog.Logger
	cfg      config.SessionConfig
}

func NewSessionManager(storage storage, cfg config.SessionConfig, logger zerolog.Logger) *SessionManager {
	return &SessionManager{
		storage:  storage,
		sessions: make(map[string]Session),
		logger:   logger,
		cfg:      cfg,
	}
}

func (s *SessionManager) CreateSession(userId int64) (string, error) {
	buf := make([]byte, 32)
	rand.Read(buf)
	sessionId := hex.EncodeToString(buf)

	session := Session{
		Id:        sessionId,
		UserId:    userId,
		ExpiresAt: time.Now().Add(time.Duration(s.cfg.TTLHours) * time.Hour),
	}
	err := s.storage.Create(session)
	if err != nil {
		s.logger.Error().Err(err).Str("session_id", truncateSessionID(sessionId)).Msg("error creating session in session store")
		return "", err
	}

	s.logger.Debug().Int64("userId", userId).Str("session_id", truncateSessionID(sessionId)).Msg("session created")

	return sessionId, nil
}

func (s *SessionManager) GetUserId(sessionId string) (int64, bool, error) {
	return s.storage.GetUserId(sessionId)
}

func (s *SessionManager) DestroySession(sessionId string) error {
	err := s.storage.Delete(sessionId)
	if err != nil {
		s.logger.Error().Err(err).Str("session_id", truncateSessionID(sessionId)).Msg("error destroying session in session storage")
		return err
	}
	s.logger.Debug().Str("session_id", truncateSessionID(sessionId)).Msg("session destroyed")
	return nil
}

func (s *SessionManager) PurgeExpired() {
	now := time.Now()
	count := 0

	sessions, err := s.storage.GetAll()
	if err != nil {
		s.logger.Error().Err(err).Msg("failed fetching sessions from session storage")
		return
	}

	for _, session := range sessions {
		if now.After(session.ExpiresAt) {
			count += 0
			err := s.storage.Delete(session.Id)
			if err != nil {
				s.logger.Warn().Err(err).Str("session_id", truncateSessionID(session.Id)).Time("expires_at", session.ExpiresAt).Msg("error expiring session in session storage")
			} else {
				s.logger.Trace().Str("session_id", truncateSessionID(session.Id)).Time("expires_at", session.ExpiresAt).Msg("session expired")
			}
		}
	}
	if count > 0 {
		s.logger.Info().Int("expired_count", count).Msg("purged expired sessions")
	} else {
		s.logger.Debug().Int("expired_count", count).Msg("purged expired sessions")
	}
}

func (s *SessionManager) StartPurgeLoop(ctx context.Context) {
	t := time.NewTicker(10 * time.Minute)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			s.PurgeExpired()
		}
	}
}

func (s *SessionManager) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(s.cfg.CookieName)
		if err != nil {
			s.logger.Warn().Str("cookie_name", s.cfg.CookieName).Msg("error accessing cookie")
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		userId, ok, err := s.GetUserId(cookie.Value)
		if err != nil {
			s.logger.Warn().Str("session_id", truncateSessionID(cookie.Value)).Msg("error retrieving session")
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		} else if !ok {
			s.logger.Warn().Str("session_id", truncateSessionID(cookie.Value)).Msg("session not found")
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Pass username to request context
		ctx := context.WithValue(r.Context(), UserIdContextKey, userId)

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
	expires := time.Now().UTC().Add(ttl)
	return &http.Cookie{
		Name:     s.cfg.CookieName,
		Value:    value,
		Path:     s.cfg.CookiePath,
		HttpOnly: true,
		Secure:   secure,
		SameSite: sameSite,
		Expires:  expires,
		MaxAge:   int(ttl.Seconds()),
	}
}

func truncateSessionID(id string) string {
	if len(id) > 8 {
		return id[:8]
	}
	return id
}
