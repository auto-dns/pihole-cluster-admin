package store

import (
	"database/sql"
	"errors"
	"strings"

	"github.com/auto-dns/pihole-cluster-admin/internal/crypto"
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

func (s *PiholeStore) AddPiholeNode(params AddPiholeParams) (*PiholeNode, error) {
	plaintextPassword := strings.TrimSpace(params.Password)
	encryptedPassword, err := crypto.EncryptPassword(s.encryptionKey, plaintextPassword)
	if err != nil {
		s.logger.Error().Err(err).Msg("error encrypting password")
		return nil, err
	}

	result, err := s.db.Exec(`
        INSERT INTO piholes
		(scheme, host, port, name, description, password_enc, created_at, updated_at)
        VALUES
		(?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`,
		strings.TrimSpace(params.Scheme), strings.TrimSpace(params.Host), params.Port, strings.TrimSpace(params.Name), strings.TrimSpace(params.Description), encryptedPassword)

	if err != nil {
		s.logger.Error().Err(err).Msg("error adding pihole entry to database")
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		s.logger.Error().Err(err).Msg("error getting last insert id")
		return nil, err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		s.logger.Error().Err(err).Msg("error getting rows_affected")
		return nil, err
	}

	s.logger.Debug().Int64("rows_affected", rowsAffected).Int64("last_insert_id", id).Msg("pihole added to database")

	insertedNode, err := s.getPiholeNodeWithPassword(id)
	if err != nil {
		s.logger.Error().Err(err).Int64("id", id).Msg("error retrieving added pihole")
		return nil, err
	}

	return insertedNode, nil
}

func (s *PiholeStore) UpdatePiholeNode(id int64, params UpdatePiholeParams) (*PiholeNode, error) {
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
			s.logger.Error().Err(err).Msg("error encrypting password")
			return nil, err
		}
		updateParts = append(updateParts, "password_enc = ?")
		args = append(args, encryptedPassword)
	}

	if len(args) == 0 {
		err := errors.New("no update fields provided")
		s.logger.Error().Err(err).Msg("bad input")
		return nil, err
	}

	updateClause := strings.Join(updateParts, ", ")

	query := "UPDATE piholes SET " + updateClause + " WHERE id = ?"
	args = append(args, id)

	result, err := s.db.Exec(query, args...)

	if err != nil {
		s.logger.Error().Err(err).Msg("error adding pihole entry to database")
		return nil, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		s.logger.Error().Err(err).Msg("error getting rows_affected")
		return nil, err
	}

	s.logger.Debug().Int64("rows_affected", rowsAffected).Int64("id", id).Msg("pihole entry updated in database")

	insertedNode, err := s.getPiholeNodeWithPassword(id)
	if err != nil {
		s.logger.Error().Err(err).Int64("id", id).Msg("error retrieving added pihole")
		return nil, err
	}

	return insertedNode, nil
}

func (s *PiholeStore) getPiholeNodeWithPassword(id int64) (*PiholeNode, error) {
	var node PiholeNode
	var encryptedPassword string
	err := s.db.QueryRow(`
        SELECT id, scheme, host, port, name, description, password_enc, created_at, updated_at
        FROM piholes WHERE id = ?`, id).Scan(
		&node.Id, &node.Scheme, &node.Host, &node.Port, &node.Name, &node.Description, &encryptedPassword, &node.CreatedAt, &node.UpdatedAt)
	if err != nil {
		return nil, err
	}

	password, err := crypto.DecryptPassword(s.encryptionKey, encryptedPassword)
	if err != nil {
		s.logger.Error().Err(err).Int64("id", node.Id).Msg("decrypting password")
		return nil, err
	}
	node.Password = &password

	return &node, nil
}

func (s *PiholeStore) GetPiholeNodeWithPassword(id int64) (*PiholeNode, error) {
	var node PiholeNode
	var encryptedPassword string
	err := s.db.QueryRow(`
        SELECT
			id,
			scheme,
			host,
			port,
			name,
			description,
			password_enc,
			created_at,
			updated_at
        FROM piholes
		WHERE id = ?`, id).Scan(
		&node.Id, &node.Scheme, &node.Host, &node.Port, &node.Name, &node.Description, &encryptedPassword, &node.CreatedAt, &node.UpdatedAt)
	if err != nil {
		s.logger.Error().Err(err).Int64("id", id).Msg("error getting pihole node from database")
		return nil, err
	}

	password, err := crypto.DecryptPassword(s.encryptionKey, encryptedPassword)
	if err != nil {
		s.logger.Error().Err(err).Int64("id", node.Id).Msg("decrypting password")
		return nil, err
	}
	node.Password = &password

	return &node, nil
}

func (s *PiholeStore) GetAllPiholeNodes() ([]*PiholeNode, error) {
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

	var nodes []*PiholeNode
	for rows.Next() {
		var n PiholeNode
		if err := rows.Scan(&n.Id, &n.Scheme, &n.Host, &n.Port, &n.Name, &n.Description, &n.CreatedAt, &n.UpdatedAt); err != nil {
			s.logger.Error().Err(err).Msg("error scanning row")
			return nil, err
		}

		nodes = append(nodes, &n)
	}

	err = rows.Err()
	if err != nil {
		s.logger.Error().Err(err).Msg("error getting rows")
		return nil, err
	}

	s.logger.Debug().Int("count", len(nodes)).Msg("fetched pihole nodes from the database")

	return nodes, nil
}

func (s *PiholeStore) GetAllPiholeNodesWithPasswords() ([]*PiholeNode, error) {
	rows, err := s.db.Query(`
		SELECT
			id,
			scheme,
			host,
			port,
			name,
			description,
			password_enc,
			created_at,
			updated_at
		FROM piholes`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var nodes []*PiholeNode
	for rows.Next() {
		var n PiholeNode
		var encPwd string
		if err := rows.Scan(&n.Id, &n.Scheme, &n.Host, &n.Port, &n.Name, &n.Description, &encPwd, &n.CreatedAt, &n.UpdatedAt); err != nil {
			s.logger.Error().Err(err).Msg("scanning row")
			return nil, err
		}

		// decrypt
		pwd, err := crypto.DecryptPassword(s.encryptionKey, encPwd)
		if err != nil {
			s.logger.Error().Err(err).Int64("id", n.Id).Msg("decrypting password")
			return nil, err
		}
		n.Password = &pwd

		nodes = append(nodes, &n)
	}

	err = rows.Err()
	if err != nil {
		s.logger.Error().Err(err).Msg("error getting rows")
		return nil, err
	}

	s.logger.Debug().Int("count", len(nodes)).Msg("fetched pihole nodes from the database")

	return nodes, nil
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
