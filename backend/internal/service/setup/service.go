package setup

import (
	"github.com/auto-dns/pihole-cluster-admin/internal/domain"
	"github.com/auto-dns/pihole-cluster-admin/internal/store"
	"github.com/auto-dns/pihole-cluster-admin/internal/transport/httpx"
	"github.com/rs/zerolog"
)

type Service struct {
	initStatusStore initStatusStore
	userStore       userStore
	sessionIssuer   sessionIssuer
	logger          zerolog.Logger
}

func NewService(initStatusStore initStatusStore, userStore userStore, sessionIssuer sessionIssuer, logger zerolog.Logger) *Service {
	return &Service{
		initStatusStore: initStatusStore,
		userStore:       userStore,
		sessionIssuer:   sessionIssuer,
		logger:          logger,
	}
}

func (s *Service) IsInitialized() (bool, error) {
	return s.userStore.IsInitialized()
}

func (s *Service) CreateUser(params CreateUserParams) (*domain.User, string, error) {
	// Verify app not initialized
	initialized, err := s.userStore.IsInitialized()
	if err != nil {
		return nil, "", err
	}

	if initialized {
		return nil, "", httpx.NewHttpError(httpx.ErrForbidden, "app is already initialized")
	}

	// Create user
	createUserParams := store.CreateUserParams{
		Username: params.Username,
		Password: params.Password,
	}
	user, err := s.userStore.CreateUser(createUserParams)
	if err != nil {
		return nil, "", err
	}

	err = s.initStatusStore.SetUserCreated(true)
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

func (s *Service) UpdatePiholeInitializationStatus(params UpdatePiholeInitializationStatusParams) error {
	// Fetch current initialization status from store
	currStatus, err := s.initStatusStore.GetInitializationStatus()
	if err != nil {
		return err
	}

	// Disallow updating to same status as current
	if params.Status == currStatus.PiholeStatus {
		return httpx.NewHttpError(httpx.ErrValidation, "new status is same as current status")
	}

	// Handle each inbound status
	switch params.Status {
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

	return s.initStatusStore.SetPiholeStatus(params.Status)
}
