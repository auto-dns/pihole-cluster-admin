package sessions

import (
	"sync"
	"time"

	"github.com/auto-dns/pihole-cluster-admin/internal/store"
	"github.com/rs/zerolog"
)

type SqliteSessionStore struct {
	sessionStore store.SessionStoreInterface
	mu           sync.RWMutex
	logger       zerolog.Logger
}

func NewSqliteSessionStore(sessionStore store.SessionStoreInterface, logger zerolog.Logger) *SqliteSessionStore {
	return &SqliteSessionStore{
		sessionStore: sessionStore,
	}
}

func (m *SqliteSessionStore) Create(session Session) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	params := store.CreateSessionParams{
		Id:        session.Id,
		UserId:    session.UserId,
		ExpiresAt: session.ExpiresAt,
	}
	_, err := m.sessionStore.CreateSession(params)
	return err
}

func (m *SqliteSessionStore) GetAll() ([]Session, error) {
	m.mu.RLock()
	dbSessions, err := m.sessionStore.GetAllSessions()
	m.mu.RUnlock()
	if err != nil {
		return nil, err
	}

	sessions := make([]Session, 0, len(dbSessions))
	for _, dbSession := range dbSessions {
		if dbSession != nil {
			sessions = append(sessions, Session{
				Id:        dbSession.Id,
				UserId:    dbSession.UserId,
				ExpiresAt: dbSession.ExpiresAt,
			})
		}
	}
	return sessions, nil
}

func (m *SqliteSessionStore) GetUserId(sessionId string) (int64, bool, error) {
	m.mu.RLock()
	dbSession, err := m.sessionStore.GetSession(sessionId)
	m.mu.RUnlock()

	if err != nil {
		return 0, false, err
	}

	if dbSession == nil || time.Now().After(dbSession.ExpiresAt) {
		return 0, false, nil
	}

	return dbSession.UserId, true, nil
}

func (m *SqliteSessionStore) Delete(sessionId string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	_, err := m.sessionStore.DeleteSession(sessionId)
	return err
}
