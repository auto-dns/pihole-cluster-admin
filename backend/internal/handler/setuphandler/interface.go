package setuphandler

import (
	"net/http"

	"github.com/auto-dns/pihole-cluster-admin/internal/domain"
	"github.com/auto-dns/pihole-cluster-admin/internal/service/setupservice"
)

type service interface {
	IsInitialized() (bool, error)
	CreateUser(setupservice.CreateUserParams) (*domain.User, string, error)
	GetInitializationStatus() (*domain.InitStatus, error)
	UpdatePiholeInitializationStatus(params setupservice.UpdatePiholeInitializationStatusParams) error
}

type httpCookieFactory interface {
	Cookie(value string) *http.Cookie
}
