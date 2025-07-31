package pihole

import (
	"sync"
	"time"

	"github.com/google/uuid"
)

type CursorManager[T any] struct {
	mu      sync.Mutex
	cursors map[string]*CursorState[T] // app cursor -> client-specific cursors
}

type CursorState[T any] struct {
	ExpireAt    time.Time
	NodeCursors map[int64]int // node Id → node cursor
	Options     T
}

func (m *CursorManager[T]) NewCursor(options T, nodeCursors map[int64]int) string {
	id := uuid.NewString()
	m.mu.Lock()
	m.cursors[id] = &CursorState[T]{
		ExpireAt:    time.Now().Add(5 * time.Minute),
		NodeCursors: nodeCursors,
		Options:     options,
	}
	m.mu.Unlock()
	return id
}

func (m *CursorManager[T]) GetCursor(id string) (*CursorState[T], bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	cur, ok := m.cursors[id]
	if ok && time.Now().After(cur.ExpireAt) {
		delete(m.cursors, id)
		return nil, false
	}
	return cur, ok
}
