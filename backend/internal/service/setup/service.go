package setup

import (
	"context"

	"github.com/auto-dns/pihole-cluster-admin/internal/domain"
	"github.com/auto-dns/pihole-cluster-admin/internal/store"
	"github.com/auto-dns/pihole-cluster-admin/internal/transport/httpx"
	"github.com/rs/zerolog"
)

type Service struct {
	initStatusStore initStatusStore
	userStore       userStore
	sessionIssuer   sessionIssuer
	tx              txProvider
	logger          zerolog.Logger
}

func NewService(initStatusStore initStatusStore, userStore userStore, sessionIssuer sessionIssuer, tx txProvider, logger zerolog.Logger) *Service {
	return &Service{
		initStatusStore: initStatusStore,
		userStore:       userStore,
		sessionIssuer:   sessionIssuer,
		tx:              tx,
		logger:          logger,
	}
}

func (s *Service) IsInitialized() (bool, error) {
	return s.userStore.IsInitialized()
}

func (s *Service) CreateUser(ctx context.Context, cmd CreateUserCommand) (*domain.User, string, error) {
	var user *domain.User
	err := s.tx.WithTx(context.Background(), func(ctx context.Context, q store.DBTX) error {
		initialized, err := s.userStore.IsInitializedTx(ctx, q)
		if err != nil {
			return err
		} else if initialized {
			return httpx.NewHttpError(httpx.ErrForbidden, "app is already initialized")
		}

		// Create user
		createUserParams := store.CreateUserParams{
			Username: cmd.Username,
			Password: cmd.Password,
		}
		u, err := s.userStore.CreateUserTx(ctx, q, createUserParams)
		if err != nil {
			return err
		}
		user = u

		err = s.initStatusStore.SetUserCreatedTx(ctx, q, true)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, "", err
	}

	// Create a session and return a cookie
	sessionId, err := s.sessionIssuer.CreateSession(user.Id)
	return user, sessionId, err
}

func (s *Service) GetInitializationStatus() (*domain.InitStatus, error) {
	return s.initStatusStore.GetInitializationStatus()
}

func (s *Service) UpdatePiholeInitializationStatus(cmd UpdatePiholeInitializationStatusCommand) error {
	// Fetch current initialization status from store
	currStatus, err := s.initStatusStore.GetInitializationStatus()
	if err != nil {
		return err
	}

	// Disallow updating to same status as current
	if cmd.Status == currStatus.PiholeStatus {
		return httpx.NewHttpError(httpx.ErrValidation, "new status is same as current status")
	}

	// Handle each inbound status
	switch cmd.Status {
	// Requesting to set uninitialized
	case domain.PiholeUninitialized:
		return httpx.NewHttpError(httpx.ErrValidation, "cannot update status to UNINITIALIZED")
	// Requesting to set added
	case domain.PiholeAdded:
		// Allow setting to "added" from all statuses
	// Requesting to set skipped
	case domain.PiholeSkipped:
		// Disallow setting to "skipped" from "added"
		if currStatus.PiholeStatus == domain.PiholeAdded {
			return httpx.NewHttpError(httpx.ErrValidation, "cannot update status from ADDED to SKIPPED")
		}
	}

	return s.initStatusStore.SetPiholeStatus(cmd.Status)
}
