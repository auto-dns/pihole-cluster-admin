package store

import (
	"database/sql"
	"strings"

	"github.com/auto-dns/pihole-cluster-admin/internal/domain"
	"github.com/rs/zerolog"
)

type SessionStore struct {
	db     *sql.DB
	logger zerolog.Logger
}

func NewSessionStore(db *sql.DB, logger zerolog.Logger) *SessionStore {
	return &SessionStore{
		db:     db,
		logger: logger,
	}
}

func (s *SessionStore) CreateSession(params CreateSessionParams) (*domain.Session, error) {
	_, err := s.db.Exec(`
		INSERT INTO sessions
		(id, user_id, created_at, expires_at)
		VALUES
		(?, ?, CURRENT_TIMESTAMP, ?)`,
		strings.TrimSpace(params.Id), params.UserId, params.ExpiresAt)
	if err != nil {
		return nil, err
	}

	insertedSession, err := s.GetSession(params.Id)
	if err != nil {
		return nil, err
	}

	return insertedSession, nil
}

func (s *SessionStore) GetAllSessions() ([]*domain.Session, error) {
	rows, err := s.db.Query(`
		SELECT
			id,
			user_id,
			created_at,
			expires_at
		FROM sessions`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []*domain.Session
	for rows.Next() {
		var session sessionRow
		if err := rows.Scan(&session.Id, &session.UserId, &session.CreatedAt, &session.ExpiresAt); err != nil {
			return nil, err
		}
		sessions = append(sessions, rowToDomainSession(session))
	}

	return sessions, rows.Err()
}

func (s *SessionStore) GetSession(id string) (*domain.Session, error) {
	var session sessionRow
	err := s.db.QueryRow(`
		SELECT id, user_id, created_at, expires_at
		FROM sessions WHERE id = ?`, id).Scan(&session.Id, &session.UserId, &session.CreatedAt, &session.ExpiresAt)
	if err != nil {
		return nil, err
	}

	return rowToDomainSession(session), nil
}

func (s *SessionStore) DeleteSession(id string) (found bool, err error) {
	result, err := s.db.Exec(`DELETE FROM sessions WHERE id = ?`, id)

	if err != nil {
		return found, err
	}

	_, err = result.RowsAffected()
	if err != nil {
		return found, err
	}

	return found, nil
}

func rowToDomainSession(row sessionRow) *domain.Session {
	return &domain.Session{
		Id:        row.Id,
		UserId:    row.UserId,
		CreatedAt: row.CreatedAt,
		ExpiresAt: row.ExpiresAt,
	}
}
