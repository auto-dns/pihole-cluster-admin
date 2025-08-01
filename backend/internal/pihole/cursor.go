package pihole

import (
	"sync"
	"time"

	"github.com/google/uuid"
)

// Manager cursors for an entire cluster
type CursorManager[T any] struct {
	mu                   sync.RWMutex
	ttlHours             int
	searchStatesByCursor map[string]SearchStateInterface[T] // app cursor -> client-specific cursors
}

func NewCursorManager[T any](ttlHours int) CursorManagerInterface[T] {
	return &CursorManager[T]{
		searchStatesByCursor: make(map[string]SearchStateInterface[T]),
	}
}

func (m *CursorManager[T]) CreateCursor(requestParams T, piholeCursors map[int64]int) string {
	cursor := uuid.NewString()
	m.mu.Lock()
	defer m.mu.Unlock()
	m.searchStatesByCursor[cursor] = NewSearchState[T](m.ttlHours, requestParams, piholeCursors)
	return cursor
}

func (m *CursorManager[T]) GetSearchState(cursor string) (SearchStateInterface[T], bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	searchState, ok := m.searchStatesByCursor[cursor]
	if ok && time.Now().After(searchState.Expiration()) {
		delete(m.searchStatesByCursor, cursor)
		return nil, false
	}
	return searchState, ok
}

func (m *CursorManager[T]) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.searchStatesByCursor = make(map[string]SearchStateInterface[T])
}

// State of the search request

type SearchState[T any] struct {
	expireAt      time.Time
	piholeCursors map[int64]int // node Id → node cursor
	requestParams T
}

func NewSearchState[T any](ttlHours int, requestParams T, piholeCursors map[int64]int) SearchStateInterface[T] {
	return &SearchState[T]{
		expireAt:      time.Now().Add(time.Duration(ttlHours) * time.Hour),
		piholeCursors: piholeCursors,
		requestParams: requestParams,
	}
}

func (s *SearchState[T]) Expiration() time.Time {
	return s.expireAt
}

func (s *SearchState[T]) GetRequestParams() T {
	return s.requestParams
}

func (s *SearchState[T]) GetPiholeCursor(id int64) (int, bool) {
	cursor, ok := s.piholeCursors[id]
	return cursor, ok
}
