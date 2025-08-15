package sessions

import (
	"sync"
	"time"
)

type MemorySessionStore struct {
	sessions map[string]Session
	mu       sync.RWMutex
}

func NewMemorySessionStore() *MemorySessionStore {
	return &MemorySessionStore{
		sessions: make(map[string]Session),
	}
}

func (m *MemorySessionStore) Create(session Session) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sessions[session.Id] = session
	return nil
}

func (m *MemorySessionStore) GetAll() ([]Session, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	sessions := make([]Session, 0, len(m.sessions))
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
