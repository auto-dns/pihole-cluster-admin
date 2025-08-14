package store

import (
	"database/sql"
	"strings"

	"github.com/rs/zerolog"
)

type SessionStore struct {
	db     *sql.DB
	logger zerolog.Logger
}

func NewSessionStore(db *sql.DB, logger zerolog.Logger) SessionStoreInterface {
	return &SessionStore{
		db:     db,
		logger: logger,
	}
}

func (s *SessionStore) CreateSession(params CreateSessionParams) (*Session, error) {
	_, err := s.db.Exec(`
		INSERT INTO sessions
		(id, user_id, created_at, expires_at)
		VALUES
		(?, ?, CURRENT_TIMESTAMP, ?)`,
		strings.TrimSpace(params.Id), params.UserId, params.ExpiresAt)
	if err != nil {
		s.logger.Error().Err(err).Msg("error adding session entry to database")
		return nil, err
	}

	insertedSession, err := s.GetSession(params.Id)
	if err != nil {
		s.logger.Error().Err(err).Msg("error retrieving added session")
		return nil, err
	}

	return insertedSession, nil
}

func (s *SessionStore) GetAllSessions() ([]*Session, error) {
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

	var sessions []*Session
	for rows.Next() {
		var session Session
		if err := rows.Scan(&session.Id, &session.UserId, &session.CreatedAt, &session.ExpiresAt); err != nil {
			s.logger.Error().Err(err).Msg("scanning row")
			return nil, err
		}
		sessions = append(sessions, &session)
	}

	err = rows.Err()
	if err != nil {
		s.logger.Error().Err(err).Msg("getting rows")
	}

	s.logger.Debug().Int("count", len(sessions)).Msg("fetched sessions from the database")

	return sessions, nil
}

func (s *SessionStore) GetSession(id string) (*Session, error) {
	var session Session
	err := s.db.QueryRow(`
		SELECT id, user_id, created_at, expires_at
		FROM sessions WHERE id = ?`, id).Scan(&session.Id, &session.UserId, &session.CreatedAt, &session.ExpiresAt)
	if err != nil {
		return nil, err
	}

	return &session, nil
}

func (s *SessionStore) DeleteSession(id string) (found bool, err error) {
	result, err := s.db.Exec(`DELETE FROM sessions WHERE id = ?`, id)

	if err != nil {
		s.logger.Error().Err(err).Msg("error removing session from the database")
		return found, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		s.logger.Error().Err(err).Msg("error getting number of affected rows")
		return found, err
	}

	if rowsAffected > 0 {
		found = true
		s.logger.Debug().Int64("rows_affected", rowsAffected).Msg("removed session from the database")
	} else {
		s.logger.Warn().Int64("rows_affected", rowsAffected).Msg("did not find session for removal from database")
	}

	return found, nil
}
