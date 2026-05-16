package cookies

import (
	"net/http"
	"strings"
	"time"

	"github.com/auto-dns/pihole-cluster-admin/internal/config"
)

type SessionCookieFactory struct {
	cfg config.SessionConfig
}

func NewSessionCookieFactory(cfg config.SessionConfig) *SessionCookieFactory {
	return &SessionCookieFactory{cfg: cfg}
}

func (f *SessionCookieFactory) Name() string {
	return f.cfg.CookieName
}

func (f *SessionCookieFactory) Make(value string) *http.Cookie {
	ttl := time.Duration(f.cfg.TTLHours) * time.Hour
	secure := f.cfg.Secure && !f.cfg.AllowInsecureCookie
	expires := time.Now().UTC().Add(ttl)

	return &http.Cookie{
		Name:     f.cfg.CookieName,
		Value:    value,
		Path:     f.cfg.CookiePath,
		HttpOnly: true,
		Secure:   secure,
		SameSite: parseSameSite(f.cfg.SameSite),
		Expires:  expires,
		MaxAge:   int(ttl.Seconds()),
	}
}

func (f *SessionCookieFactory) MakeCSRF(token string) *http.Cookie {
	ttl := time.Duration(f.cfg.TTLHours) * time.Hour
	secure := f.cfg.Secure && !f.cfg.AllowInsecureCookie
	expires := time.Now().UTC().Add(ttl)

	return &http.Cookie{
		Name:     "csrf_token",
		Value:    token,
		Path:     f.cfg.CookiePath,
		HttpOnly: false, // JS must read this to send it as a header
		Secure:   secure,
		SameSite: parseSameSite(f.cfg.SameSite),
		Expires:  expires,
		MaxAge:   int(ttl.Seconds()),
	}
}

func parseSameSite(val string) http.SameSite {
	switch strings.ToLower(val) {
	case "lax":
		return http.SameSiteLaxMode
	case "none":
		return http.SameSiteNoneMode
	default:
		return http.SameSiteStrictMode
	}
}
