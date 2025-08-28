package store

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/auto-dns/pihole-cluster-admin/internal/domain"
	"github.com/rs/zerolog"
)

type InitializationStatusStore struct {
	db     *sql.DB
	logger zerolog.Logger
}

const initStatusRowId = 1

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
		WHERE id = ?
	`, initStatusRowId).Scan(&row.UserCreated, &row.PiholeStatus)
	if err != nil {
		return nil, err
	}

	return s.rowToDomainInitStatus(row)
}

func (s *InitializationStatusStore) SetUserCreatedTx(ctx context.Context, q DBTX, userCreated bool) error {
	_, err := q.ExecContext(ctx, `
        UPDATE initialization_status
        SET user_created = ?
        WHERE id = ?
    `, userCreated, initStatusRowId)
	return err
}

func (s *InitializationStatusStore) SetPiholeStatus(piholeStatus domain.PiholeStatus) error {
	piholeStatusStr, err := fromDomainPiholeStatus(piholeStatus)
	if err != nil {
		return err
	}

	_, err = s.db.Exec(`
        UPDATE initialization_status
        SET pihole_status = ?
        WHERE id = ?
    `, piholeStatusStr, initStatusRowId)
	return err
}

func (s *InitializationStatusStore) rowToDomainInitStatus(row initStatusRow) (*domain.InitStatus, error) {
	piholeStatus, err := toDomainPiholeStatus(row.PiholeStatus)
	if err != nil {
		return nil, err
	}

	return &domain.InitStatus{
		UserCreated:  row.UserCreated,
		PiholeStatus: piholeStatus,
	}, nil
}

func toDomainPiholeStatus(s string) (domain.PiholeStatus, error) {
	switch s {
	case string(domain.PiholeUninitialized),
		string(domain.PiholeAdded),
		string(domain.PiholeSkipped):
		return domain.PiholeStatus(s), nil
	default:
		return "", fmt.Errorf("invalid pihole status in DB: %q", s)
	}
}

func fromDomainPiholeStatus(st domain.PiholeStatus) (string, error) {
	if !st.IsValid() {
		return "", fmt.Errorf("invalid pihole status: %q", st)
	}
	return string(st), nil
}
