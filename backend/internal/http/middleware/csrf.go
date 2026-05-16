package middleware

import (
	"net/http"
)

const (
	CSRFHeaderName = "X-CSRF-Token"
	CSRFCookieName = "csrf_token"
)

// CSRF implements the double-submit cookie pattern. On mutating requests it
// verifies that the X-CSRF-Token header matches the csrf_token cookie value.
// Safe methods (GET, HEAD, OPTIONS) are unconditionally allowed.
func CSRF(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet, http.MethodHead, http.MethodOptions:
			next.ServeHTTP(w, r)
			return
		}

		cookie, err := r.Cookie(CSRFCookieName)
		if err != nil || cookie.Value == "" {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		if r.Header.Get(CSRFHeaderName) != cookie.Value {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}
