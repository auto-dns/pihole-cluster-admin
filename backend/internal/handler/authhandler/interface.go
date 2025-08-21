package authhandler

import (
	"net/http"

	"github.com/auto-dns/pihole-cluster-admin/internal/domain"
	"github.com/auto-dns/pihole-cluster-admin/internal/service/authservice"
)

type service interface {
	Login(params authservice.LoginParams) (*domain.User, string, error)
	Logout(sessionId string) error
	GetUser(id int64) (*domain.User, error)
}

type httpCookieFactory interface {
	Cookie(value string) *http.Cookie
	CookieName() string
}
