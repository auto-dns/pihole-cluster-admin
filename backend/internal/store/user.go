package store

import (
	"database/sql"
	"strings"

	"github.com/rs/zerolog"
	"golang.org/x/crypto/bcrypt"
)

type UserStore struct {
	db     *sql.DB
	logger zerolog.Logger
}

func NewUserStore(db *sql.DB, logger zerolog.Logger) UserStoreInterface {
	return &UserStore{
		db:     db,
		logger: logger,
	}
}

func (s *UserStore) CreateUser(params CreateUserParams) (*User, error) {
	logger := s.logger.With().Str("username", params.Username).Logger()

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(strings.TrimSpace(params.Password)), bcrypt.DefaultCost)
	if err != nil {
		logger.Error().Err(err).Msg("error hashing password")
		return nil, err
	}

	result, err := s.db.Exec(`INSERT INTO users (username, password_hash) VALUES (?, ?)`, strings.ToLower(strings.TrimSpace(params.Username)), string(passwordHash))
	if err != nil {
		s.logger.Error().Err(err).Msg("error inserting uesr into database")
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		logger.Error().Err(err).Msg("error getting last insert Id")
		return nil, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		logger.Error().Err(err).Msg("error getting affected row count")
		return nil, err
	}

	s.logger.Debug().Int64("rows_affected", rowsAffected).Int64("last_insert_id", id).Msg("user added to database")

	insertedUser, err := s.getUser(id)
	if err != nil {
		s.logger.Error().Err(err).Int64("id", id).Msg("error getting inserted user")
		return nil, err
	}

	return insertedUser, err
}

func (s *UserStore) getUser(id int64) (*User, error) {
	var user User
	err := s.db.QueryRow(`
        SELECT id, username, password_hash, created_at, updated_at
        FROM users WHERE id = ?`, id).Scan(
		&user.Id, &user.Username, &user.PasswordHash, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		s.logger.Error().Err(err).Int64("id", id).Msg("error getting user from database")
		return nil, err
	}

	return &user, nil
}

func (s *UserStore) ValidateUser(username, password string) (bool, error) {
	var hash string
	err := s.db.QueryRow(`SELECT password_hash FROM users WHERE username = ?`, strings.ToLower(username)).Scan(&hash)
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
