package user

import (
	"errors"
	"strings"

	"github.com/auto-dns/pihole-cluster-admin/internal/crypto"
	"github.com/auto-dns/pihole-cluster-admin/internal/domain"
	"github.com/auto-dns/pihole-cluster-admin/internal/errs"
	"github.com/auto-dns/pihole-cluster-admin/internal/store"
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
		return nil, errs.New(errs.KindUnknown, "unknown error", err)
	}

	// Validate content
	if cmd.Username != nil {
		if strings.TrimSpace(*cmd.Username) == strings.TrimSpace(currentUser.Username) {
			return nil, errs.Invalid("username empty", errors.New("username empty"))
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
		return nil, errs.Unauthorized("wrong password", errors.New("current password incorrect"))
	}

	if crypto.CompareHashAndPassword(currentUserAuth.PasswordHash, cmd.NewPassword) == nil {
		return nil, errs.Invalid("new password matches current password", errors.New("new password matches current password"))
	}

	// Call user store to do the update
	updateParams := store.UpdateUserParams{
		Password: &cmd.NewPassword,
	}
	return s.userStore.UpdateUser(id, updateParams)
}
