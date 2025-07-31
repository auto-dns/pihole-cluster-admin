package store

import (
	"database/sql"

	"github.com/rs/zerolog"
	"golang.org/x/crypto/bcrypt"
)

type UserStore struct {
	db     *sql.DB
	logger zerolog.Logger
}

func NewUserStore(db *sql.DB, logger zerolog.Logger) *UserStore {
	return &UserStore{
		db:     db,
		logger: logger,
	}
}

func (s *UserStore) CreateUser(username, password string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	_, err = s.db.Exec(`INSERT INTO users (username, password_hash) VALUES (?, ?)`, username, string(hash))
	return err
}

func (s *UserStore) DeleteUser(username string) error {
	_, err := s.db.Exec(`DELETE FROM users WHERE username = ?`, username)
	return err
}

func (s *UserStore) ValidateUser(username, password string) (bool, error) {
	var hash string
	err := s.db.QueryRow(`SELECT password_hash FROM users WHERE username = ?`, username).Scan(&hash)
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
	err := s.db.QueryRow(`SELECT COUNT(*) FROM users`).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
