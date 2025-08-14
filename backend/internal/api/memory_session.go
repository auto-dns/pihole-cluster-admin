package api

import (
	"sync"
	"time"
)

type MemorySessionStore struct {
	sessions map[string]session
	mu       sync.RWMutex
}

func NewMemorySessionStore() SessionStorageInterface {
	return &MemorySessionStore{
		sessions: make(map[string]session),
	}
}

func (m *MemorySessionStore) Create(session session) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sessions[session.Id] = session
	return nil
}

func (m *MemorySessionStore) GetAll() ([]session, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	sessions := make([]session, 0, len(m.sessions))
	for _, session := range m.sessions {
		sessions = append(sessions, session)
	}
	return sessions, nil
}

func (m *MemorySessionStore) GetUserId(sessionId string) (int64, bool, error) {
	m.mu.RLock()
	sess, ok := m.sessions[sessionId]
	m.mu.RUnlock()

	if !ok || time.Now().After(sess.ExpiresAt) {
		return 0, false, nil
	}

	return sess.UserId, true, nil
}

func (m *MemorySessionStore) Delete(sessionId string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.sessions, sessionId)
	return nil
}
