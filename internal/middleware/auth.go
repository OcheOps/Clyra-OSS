package middleware

import (
	"net/http"

	"github.com/oche/diagram/internal/auth"
	"github.com/oche/diagram/pkg/response"
)

// OptionalAuthenticate is like Authenticate but never short-circuits with
// 401. If a valid session exists, the user is injected into the context;
// otherwise the request continues anonymously. Used by endpoints that
// behave differently for logged-in vs. logged-out callers (e.g. the public
// support form attaches a user_id when available).
func (m *Middleware) OptionalAuthenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if user, err := m.authService.ValidateSession(r.Context(), r); err == nil {
			ctx := auth.NewContext(r.Context(), user)
			r = r.WithContext(ctx)
		}
		next.ServeHTTP(w, r)
	})
}

// Authenticate validates the session cookie and injects the user into the request context.
//
// Backwards-compat: if a valid session exists but no CSRF cookie is present
// (e.g. a user who was signed in before the CSRF middleware was deployed),
// we mint a fresh CSRF cookie on the response. Their next mutating request
// will succeed without having to log in again.
func (m *Middleware) Authenticate(next http.Handler) http.Handler {
	csrfName := m.cfg.Session.CookieName + "_csrf"
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sessionUser, err := m.authService.ValidateSession(r.Context(), r)
		if err != nil {
			response.Error(w, http.StatusUnauthorized, "invalid or expired session")
			return
		}

		// Re-arm CSRF cookie for legacy sessions that predate the hardening push.
		if _, cookieErr := r.Cookie(csrfName); cookieErr != nil {
			http.SetCookie(w, m.authService.IssueCSRFCookie())
		}

		ctx := auth.NewContext(r.Context(), sessionUser)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
