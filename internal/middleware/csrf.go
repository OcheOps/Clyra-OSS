package middleware

import (
	"crypto/subtle"
	"net/http"

	"github.com/oche/diagram/pkg/response"
)

// CSRF enforces the double-submit cookie pattern for authenticated mutating
// requests. The frontend reads the `clyra_csrf` cookie (non-HttpOnly) from JS
// and echoes its value as the X-CSRF-Token header; this middleware confirms
// they match.
//
// Apply inside the authenticated group, AFTER Authenticate — an unauth'd user
// never holds a csrf cookie so we'd 401 anyway.
func (m *Middleware) CSRF(next http.Handler) http.Handler {
	cookieName := m.cfg.Session.CookieName + "_csrf"
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Read-only methods don't need protection.
		switch r.Method {
		case http.MethodGet, http.MethodHead, http.MethodOptions:
			next.ServeHTTP(w, r)
			return
		}

		cookie, err := r.Cookie(cookieName)
		if err != nil || cookie.Value == "" {
			response.Error(w, http.StatusForbidden, "csrf_cookie_missing")
			return
		}
		header := r.Header.Get("X-CSRF-Token")
		if header == "" {
			response.Error(w, http.StatusForbidden, "csrf_header_missing")
			return
		}
		if subtle.ConstantTimeCompare([]byte(header), []byte(cookie.Value)) != 1 {
			response.Error(w, http.StatusForbidden, "csrf_token_mismatch")
			return
		}
		next.ServeHTTP(w, r)
	})
}
