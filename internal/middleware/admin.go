package middleware

import (
	"net/http"

	"github.com/oche/diagram/internal/auth"
	"github.com/oche/diagram/pkg/response"
)

// RequireAdmin restricts access to admin users only.
func (m *Middleware) RequireAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := auth.UserFromContext(r.Context())
		if user == nil || user.Role != auth.RoleAdmin {
			response.Error(w, http.StatusForbidden, "admin access required")
			return
		}
		next.ServeHTTP(w, r)
	})
}
