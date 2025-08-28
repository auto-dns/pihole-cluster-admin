package user

import (
	"errors"
	"strings"

	"github.com/auto-dns/pihole-cluster-admin/internal/crypto"
	"github.com/auto-dns/pihole-cluster-admin/internal/domain"
	"github.com/auto-dns/pihole-cluster-admin/internal/store"
	"github.com/auto-dns/pihole-cluster-admin/internal/transport/httpx"
	"github.com/rs/zerolog"
)

type Service struct {
	userStore userStore
	logger    zerolog.Logger
}

func NewService(userStore userStore, logger zerolog.Logger) *Service {
	return &Service{
		userStore: userStore,
		logger:    logger,
	}
}

func (s *Service) Patch(id int64, cmd PatchUserCommand) (*domain.User, error) {
	currentUser, err := s.userStore.GetUser(id)
	if err != nil {
		return nil, err
	}
	if currentUser == nil {
		return nil, httpx.NewHttpError(httpx.ErrInternalService, "error getting user")
	}

	// Validate content
	if cmd.Username != nil {
		if strings.TrimSpace(*cmd.Username) == strings.TrimSpace(currentUser.Username) {
			return nil, httpx.NewHttpError(httpx.ErrValidation, "username empty")
		}
	}

	// Call user store to update the node
	updateParams := store.UpdateUserParams{
		Username: cmd.Username,
	}

	updatedNode, err := s.userStore.UpdateUser(id, updateParams)

	return updatedNode, err
}

func (s *Service) UpdatePassword(id int64, cmd UpdatePasswordCommand) (*domain.User, error) {
	currentUserAuth, err := s.userStore.GetUserAuth(id)
	if err != nil {
		return nil, err
	} else if currentUserAuth == nil {
		return nil, errors.New("error fetching password hash")
	}

	// Validate content
	if crypto.CompareHashAndPassword(currentUserAuth.PasswordHash, cmd.CurrentPassword) != nil {
		return nil, httpx.NewHttpError(httpx.ErrUnauthorized, "current password incorrect")
	}

	if crypto.CompareHashAndPassword(currentUserAuth.PasswordHash, cmd.NewPassword) == nil {
		return nil, httpx.NewHttpError(httpx.ErrValidation, "new password matches current password")
	}

	// Call user store to do the update
	updateParams := store.UpdateUserParams{
		Password: &cmd.NewPassword,
	}
	return s.userStore.UpdateUser(id, updateParams)
}
