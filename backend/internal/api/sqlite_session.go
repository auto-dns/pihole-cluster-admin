package api

import (
	"sync"

	"github.com/auto-dns/pihole-cluster-admin/internal/store"
	"github.com/rs/zerolog"
)

type SqliteSessionStore struct {
	sessionStore store.SessionStoreInterface
	mu           sync.RWMutex
	logger       zerolog.Logger
}

func NewSqliteSessionStore(sessionStore store.SessionStoreInterface, logger zerolog.Logger) SessionStorageInterface {
	return &SqliteSessionStore{
		sessionStore: sessionStore,
	}
}

func (m *SqliteSessionStore) Create(session session) error {
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

func (m *SqliteSessionStore) GetAll() ([]session, error) {
	m.mu.RLock()
	dbSessions, err := m.sessionStore.GetAllSessions()
	m.mu.RUnlock()
	if err != nil {
		return nil, err
	}

	sessions := make([]session, 0, len(dbSessions))
	for _, dbSession := range dbSessions {
		if dbSession != nil {
			sessions = append(sessions, session{
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
		return -1, false, err
	}

	return dbSession.UserId, true, nil
}

func (m *SqliteSessionStore) Delete(sessionId string) error {
	m.mu.RLock()
	defer m.mu.RUnlock()
	_, err := m.sessionStore.DeleteSession(sessionId)
	return err
}
