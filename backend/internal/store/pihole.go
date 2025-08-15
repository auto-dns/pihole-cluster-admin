package store

import (
	"database/sql"
	"errors"
	"strings"

	"github.com/auto-dns/pihole-cluster-admin/internal/crypto"
	"github.com/auto-dns/pihole-cluster-admin/internal/domain"
	"github.com/rs/zerolog"
)

type PiholeStore struct {
	db            *sql.DB
	encryptionKey string
	logger        zerolog.Logger
}

func NewPiholeStore(db *sql.DB, encryptionKey string, logger zerolog.Logger) *PiholeStore {
	return &PiholeStore{
		db:            db,
		encryptionKey: encryptionKey,
		logger:        logger,
	}
}

func (s *PiholeStore) getPiholeRow(id int64) (piholeRow, error) {
	var row piholeRow
	err := s.db.QueryRow(`
		SELECT id, scheme, host, port, name, description, password_enc, created_at, updated_at
		FROM piholes WHERE id = ?`, id).Scan(
		&row.Id, &row.Scheme, &row.Host, &row.Port, &row.Name, &row.Description, &row.PasswordEnc, &row.CreatedAt, &row.UpdatedAt)
	return row, err
}

func (s *PiholeStore) AddPiholeNode(params AddPiholeParams) (*domain.PiholeNode, error) {
	plaintextPassword := strings.TrimSpace(params.Password)
	encryptedPassword, err := crypto.EncryptPassword(s.encryptionKey, plaintextPassword)
	if err != nil {
		return nil, err
	}

	result, err := s.db.Exec(`
        INSERT INTO piholes
		(scheme, host, port, name, description, password_enc, created_at, updated_at)
        VALUES
		(?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`,
		strings.TrimSpace(params.Scheme), strings.TrimSpace(params.Host), params.Port, strings.TrimSpace(params.Name), strings.TrimSpace(params.Description), encryptedPassword)

	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}
	insertedNode, err := s.getPiholeRow(id)
	if err != nil {
		return nil, err
	}

	return rowToDomainNode(insertedNode), nil
}

func (s *PiholeStore) UpdatePiholeNode(id int64, params UpdatePiholeParams) (*domain.PiholeNode, error) {
	var updateParts []string
	var args []any
	if params.Scheme != nil {
		updateParts = append(updateParts, "scheme = ?")
		args = append(args, *params.Scheme)
	}
	if params.Host != nil {
		updateParts = append(updateParts, "host = ?")
		args = append(args, *params.Host)
	}
	if params.Port != nil {
		updateParts = append(updateParts, "port = ?")
		args = append(args, *params.Port)
	}
	if params.Name != nil {
		updateParts = append(updateParts, "name = ?")
		args = append(args, *params.Name)
	}
	if params.Description != nil {
		updateParts = append(updateParts, "description = ?")
		args = append(args, *params.Description)
	}
	if params.Password != nil {
		plaintextPassword := strings.TrimSpace(*params.Password)
		encryptedPassword, err := crypto.EncryptPassword(s.encryptionKey, plaintextPassword)
		if err != nil {
			return nil, err
		}
		updateParts = append(updateParts, "password_enc = ?")
		args = append(args, encryptedPassword)
	}

	if len(args) == 0 {
		err := errors.New("no update fields provided")
		return nil, err
	}

	updateClause := strings.Join(updateParts, ", ")

	query := "UPDATE piholes SET " + updateClause + " WHERE id = ?"
	args = append(args, id)

	_, err := s.db.Exec(query, args...)

	if err != nil {
		return nil, err
	}

	insertedNode, err := s.getPiholeRow(id)
	if err != nil {
		return nil, err
	}

	return rowToDomainNode(insertedNode), err
}

func (s *PiholeStore) GetPiholeNode(id int64) (*domain.PiholeNode, error) {
	row, err := s.getPiholeRow(id)
	if err != nil {
		return nil, err
	}
	return rowToDomainNode(row), nil
}

func (s *PiholeStore) GetPiholeNodeSecret(id int64) (*domain.PiholeNodeSecret, error) {
	row, err := s.getPiholeRow(id)
	if err != nil {
		return nil, err
	}
	password, err := crypto.DecryptPassword(s.encryptionKey, row.PasswordEnc)
	if err != nil {
		return nil, err
	}
	return &domain.PiholeNodeSecret{NodeId: row.Id, Password: password}, nil
}

func (s *PiholeStore) GetAllPiholeNodes() ([]*domain.PiholeNode, error) {
	rows, err := s.db.Query(`
		SELECT
			id,
			scheme,
			host,
			port,
			name,
			description,
			created_at,
			updated_at
		FROM piholes`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var nodes []*domain.PiholeNode
	for rows.Next() {
		var r piholeRow
		if err := rows.Scan(&r.Id, &r.Scheme, &r.Host, &r.Port, &r.Name, &r.Description, &r.CreatedAt, &r.UpdatedAt); err != nil {
			return nil, err
		}

		nodes = append(nodes, rowToDomainNode(r))
	}

	return nodes, rows.Err()
}

func (s *PiholeStore) RemovePiholeNode(id int64) (found bool, err error) {
	logger := s.logger.With().Int64("id", id).Logger()

	result, err := s.db.Exec(`DELETE FROM piholes WHERE id = ?`, id)

	if err != nil {
		logger.Error().Err(err).Msg("error removing pihole node from the database")
		return found, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		logger.Error().Err(err).Msg("error getting number of affected rows")
		return found, err
	}

	if rowsAffected > 0 {
		found = true
		logger.Debug().Int64("rows_affected", rowsAffected).Msg("removed pihole node from the database")
	} else {
		logger.Warn().Int64("rows_affected", rowsAffected).Msg("did not find pihole node for removal from database")
	}

	return found, nil
}

func rowToDomainNode(row piholeRow) *domain.PiholeNode {
	return &domain.PiholeNode{
		Id:          row.Id,
		Scheme:      row.Scheme,
		Host:        row.Host,
		Port:        row.Port,
		Name:        row.Name,
		Description: row.Description,
		CreatedAt:   row.CreatedAt,
		UpdatedAt:   row.UpdatedAt,
	}
}
