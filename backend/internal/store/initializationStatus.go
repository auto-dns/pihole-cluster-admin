package store

import (
	"database/sql"

	"github.com/rs/zerolog"
)

type InitializationStatusStore struct {
	db     *sql.DB
	logger zerolog.Logger
}

func NewInitializationStore(db *sql.DB, logger zerolog.Logger) InitializationStatusStoreInterface {
	return &InitializationStatusStore{
		db:     db,
		logger: logger,
	}
}

func (s *InitializationStatusStore) GetInitializationStatus() (*InitializationStatus, error) {
	var status InitializationStatus
	var piholeStatusStr string
	err := s.db.QueryRow(`
		SELECT
			user_created,
			pihole_status
		FROM initialization_status
		WHERE id = 1
	`).Scan(&status.UserCreated, &piholeStatusStr)

	if err != nil {
		s.logger.Error().Err(err).Msg("getting initialization status from database")
		return nil, err
	}

	switch PiholeStatus(piholeStatusStr) {
	case PiholeUninitialized, PiholeAdded, PiholeSkipped:
		status.PiholeStatus = PiholeStatus(piholeStatusStr)
	default:
		s.logger.Warn().Str("pihole_status", piholeStatusStr).Msg("Unknown pihole status in DB, defaulting to UNINITIALIZED")
		status.PiholeStatus = PiholeUninitialized
	}

	return &status, nil
}

func (s *InitializationStatusStore) SetUserCreated(userCreated bool) error {
	_, err := s.db.Exec(`
        UPDATE initialization_status
        SET user_created = ?
        WHERE id = 1
    `, userCreated)
	return err
}

func (s *InitializationStatusStore) SetPiholeStatus(piholeStatus PiholeStatus) error {
	_, err := s.db.Exec(`
        UPDATE initialization_status
        SET pihole_status = ?
        WHERE id = 1
    `, piholeStatus)
	return err
}
