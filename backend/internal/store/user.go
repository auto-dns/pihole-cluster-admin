package store

import (
	"database/sql"

	"github.com/auto-dns/pihole-cluster-admin/internal/database"
	"golang.org/x/crypto/bcrypt"
)

type UserStore struct {
	db *database.Database
}

func NewUserStore(db *database.Database) *UserStore {
	return &UserStore{db: db}
}

// CreateUser inserts a new user with a hashed password.
func (s *UserStore) CreateUser(username, password string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	_, err = s.db.DB.Exec(`INSERT INTO users (username, password_hash) VALUES (?, ?)`, username, string(hash))
	return err
}

// ValidateUser checks username and password.
func (s *UserStore) ValidateUser(username, password string) (bool, error) {
	var hash string
	err := s.db.DB.QueryRow(`SELECT password_hash FROM users WHERE username = ?`, username).Scan(&hash)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	if bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) != nil {
		return false, nil
	}
	return true, nil
}

// IsInitialized returns true if at least one user exists.
func (s *UserStore) IsInitialized() (bool, error) {
	var count int
	err := s.db.DB.QueryRow(`SELECT COUNT(*) FROM users`).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
