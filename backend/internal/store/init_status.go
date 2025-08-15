package store

import (
	"database/sql"
	"fmt"

	"github.com/auto-dns/pihole-cluster-admin/internal/domain"
	"github.com/rs/zerolog"
)

type InitializationStatusStore struct {
	db     *sql.DB
	logger zerolog.Logger
}

func NewInitializationStore(db *sql.DB, logger zerolog.Logger) *InitializationStatusStore {
	return &InitializationStatusStore{
		db:     db,
		logger: logger,
	}
}

func (s *InitializationStatusStore) GetInitializationStatus() (*domain.InitStatus, error) {
	var row initStatusRow
	err := s.db.QueryRow(`
		SELECT
			user_created,
			pihole_status
		FROM initialization_status
		WHERE id = 1
	`).Scan(&row.UserCreated, &row.PiholeStatus)

	if err != nil {
		return nil, err
	}

	return s.rowToDomainInitStatus(row), nil
}

func (s *InitializationStatusStore) SetUserCreated(userCreated bool) error {
	_, err := s.db.Exec(`
        UPDATE initialization_status
        SET user_created = ?
        WHERE id = 1
    `, userCreated)
	return err
}

func (s *InitializationStatusStore) SetPiholeStatus(piholeStatus domain.PiholeStatus) error {
	if !piholeStatus.IsValid() {
		return fmt.Errorf("invalid pihole status %q", piholeStatus)
	}

	_, err := s.db.Exec(`
        UPDATE initialization_status
        SET pihole_status = ?
        WHERE id = 1
    `, piholeStatus)
	return err
}

func (s *InitializationStatusStore) rowToDomainInitStatus(row initStatusRow) *domain.InitStatus {
	return &domain.InitStatus{
		UserCreated:  row.UserCreated,
		PiholeStatus: row.PiholeStatus,
	}
}
