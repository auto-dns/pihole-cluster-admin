package setup

import (
	"net/http"

	"github.com/auto-dns/pihole-cluster-admin/internal/domain"
	setup_s "github.com/auto-dns/pihole-cluster-admin/internal/service/setup"
)

type service interface {
	IsInitialized() (bool, error)
	CreateUser(setup_s.CreateUserParams) (*domain.User, string, error)
	GetInitializationStatus() (*domain.InitStatus, error)
	UpdatePiholeInitializationStatus(params setup_s.UpdatePiholeInitializationStatusParams) error
}

type httpCookieFactory interface {
	Cookie(value string) *http.Cookie
}
