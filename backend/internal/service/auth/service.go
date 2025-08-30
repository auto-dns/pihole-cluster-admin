package auth

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/auto-dns/pihole-cluster-admin/internal/domain"
	"github.com/auto-dns/pihole-cluster-admin/internal/errs"
	"github.com/auto-dns/pihole-cluster-admin/internal/store"
	"github.com/rs/zerolog"
)

type Service struct {
	userStore     userStore
	sessionIssuer sessionIssuer
	logger        zerolog.Logger
}

func NewService(userStore userStore, sessionIssuer sessionIssuer, logger zerolog.Logger) *Service {
	return &Service{
		userStore:     userStore,
		sessionIssuer: sessionIssuer,
		logger:        logger,
	}
}

func (s *Service) Login(cmd LoginCommand) (*domain.User, string, error) {
	// Validate against the database
	user, err := s.userStore.ValidateUser(cmd.Username, cmd.Password)
	var wrongPasswordErr *store.WrongPasswordError
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return nil, "", errs.Invalid("invalid credentials", err)
	case errors.As(err, &wrongPasswordErr):
		return nil, "", errs.Invalid("invalid credentials", err)
	case err != nil:
		return nil, "", errs.New(errs.KindUnknown, "unhandled error", err)
	}

	// Successful login → create session
	sessionId, err := s.sessionIssuer.CreateSession(user.Id)
	if err != nil {
		return nil, "", errs.Unauthorized("unhandled error creating session", fmt.Errorf("creating session: %w", err))
	}

	return user, sessionId, nil
}

func (s *Service) Logout(sessionId string) error {
	userId, ok, err := s.sessionIssuer.GetUserId(sessionId)
	if err != nil {
		s.logger.Error().Err(err).Int64("userId", userId).Msg("error getting user session")
	} else if ok {
		s.logger.Info().Int64("userId", userId).Msg("user logged out")
	} else {
		s.logger.Warn().Msg("user attempted logout, but no username was found in the session")
	}
	return s.sessionIssuer.DestroySession(sessionId)
}

func (s *Service) GetUser(id int64) (*domain.User, error) {
	return s.userStore.GetUser(id)
}
