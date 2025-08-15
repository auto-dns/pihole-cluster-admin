package store

import (
	"database/sql"
	"strings"

	"github.com/auto-dns/pihole-cluster-admin/internal/domain"
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

func (s *UserStore) getUserRow(id int64) (userRow, error) {
	var row userRow
	err := s.db.QueryRow(`
		SELECT id, username, password_hash, created_at, updated_at
		FROM users WHERE id = ?`, id).Scan(
		&row.Id, &row.Username, &row.PasswordHash, &row.CreatedAt, &row.UpdatedAt)
	return row, err
}

func (s *UserStore) CreateUser(params CreateUserParams) (*domain.User, error) {
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(strings.TrimSpace(params.Password)), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	result, err := s.db.Exec(`INSERT INTO users (username, password_hash) VALUES (?, ?)`, strings.ToLower(strings.TrimSpace(params.Username)), string(passwordHash))
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	insertedUser, err := s.getUserRow(id)
	if err != nil {
		return nil, err
	}

	return rowToDomainUser(insertedUser), err
}

func (s *UserStore) GetUser(id int64) (*domain.User, error) {
	row, err := s.getUserRow(id)
	return rowToDomainUser(row), err
}

type WrongPasswordError struct {
	message string
}

func (e *WrongPasswordError) Error() string {
	return e.message
}

func (s *UserStore) ValidateUser(username, password string) (*domain.User, error) {
	var row userRow
	err := s.db.QueryRow(`SELECT
			id,	
			username,
			password_hash,
			created_at,
			updated_at
		FROM users
		WHERE username = ?`, strings.ToLower(username)).Scan(&row.Id, &row.Username, &row.PasswordHash, &row.CreatedAt, &row.UpdatedAt)
	if err == sql.ErrNoRows {
		s.logger.Debug().Err(err).Msg("user not found")
		return nil, err
	} else if err != nil {
		s.logger.Error().Err(err).Msg("error getting user from database")
		return nil, err
	}

	if bcrypt.CompareHashAndPassword([]byte(row.PasswordHash), []byte(password)) != nil {
		err := &WrongPasswordError{"wrong password"}
		s.logger.Debug().Err(err).Msg("wrong password")
		return nil, err
	}
	return rowToDomainUser(row), nil
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

func rowToDomainUser(row userRow) *domain.User {
	return &domain.User{
		Id:        row.Id,
		Username:  row.Username,
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
	}
}
