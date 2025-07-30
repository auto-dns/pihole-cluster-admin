package store

import (
	"errors"
	"time"

	"github.com/auto-dns/pihole-cluster-admin/internal/crypto"
	"github.com/auto-dns/pihole-cluster-admin/internal/database"
)

type PiholeStore struct {
	db            *database.Database
	encryptionKey string
}

func NewPiholeStore(db *database.Database, encryptionKey string) *PiholeStore {
	return &PiholeStore{db: db, encryptionKey: encryptionKey}
}

type PiholeNode struct {
	ID          int
	Scheme      string
	Host        string
	Port        int
	Description *string
	Password    string // plaintext on input/output
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (s *PiholeStore) AddPiholeNode(node PiholeNode) error {
	if node.Password == "" {
		return errors.New("password required")
	}

	enc, err := crypto.EncryptPassword(s.encryptionKey, node.Password)
	if err != nil {
		return err
	}

	_, err = s.db.DB.Exec(`
        INSERT INTO piholes (scheme, host, port, description, password_enc, created_at, updated_at)
        VALUES (?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`,
		node.Scheme, node.Host, node.Port, node.Description, enc)
	return err
}

func (s *PiholeStore) GetPiholeNode(id int) (*PiholeNode, error) {
	var node PiholeNode
	var encPwd string
	err := s.db.DB.QueryRow(`
        SELECT id, scheme, host, port, description, password_enc, created_at, updated_at
        FROM piholes WHERE id = ?`, id).Scan(
		&node.ID, &node.Scheme, &node.Host, &node.Port, &node.Description, &encPwd, &node.CreatedAt, &node.UpdatedAt)
	if err != nil {
		return nil, err
	}

	// decrypt password
	pwd, err := crypto.DecryptPassword(s.encryptionKey, encPwd)
	if err != nil {
		return nil, err
	}
	node.Password = pwd

	return &node, nil
}

func (s *PiholeStore) UpdatePiholePassword(id int, newPassword string) error {
	enc, err := crypto.EncryptPassword(s.encryptionKey, newPassword)
	if err != nil {
		return err
	}
	_, err = s.db.DB.Exec(`UPDATE piholes SET password_enc = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`, enc, id)
	return err
}
