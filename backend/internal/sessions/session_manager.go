package sessions

import (
	"context"
	"crypto/rand"
	"encoding/hex"
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
	count, err := s.storage.DeleteExpired()
	if err != nil {
		s.logger.Error().Err(err).Msg("purging expired sessions")
	}

	if count > 0 {
		s.logger.Info().Int64("expired_count", count).Msg("purged expired sessions")
	} else {
		s.logger.Debug().Int64("expired_count", count).Msg("purged expired sessions")
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

func truncateSessionID(id string) string {
	if len(id) > 8 {
		return id[:8]
	}
	return id
}
