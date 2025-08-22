package auth

import (
	"net/http"

	"github.com/auto-dns/pihole-cluster-admin/internal/domain"
	auth_s "github.com/auto-dns/pihole-cluster-admin/internal/service/auth"
)

type service interface {
	Login(params auth_s.LoginParams) (*domain.User, string, error)
	Logout(sessionId string) error
	GetUser(id int64) (*domain.User, error)
}

type httpCookieFactory interface {
	Cookie(value string) *http.Cookie
	CookieName() string
}
