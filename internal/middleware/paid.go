package middleware

import (
	"net/http"

	"github.com/oche/diagram/internal/auth"
	"github.com/oche/diagram/pkg/response"
)

// RequirePaid blocks free-plan users. Used to gate Pro/BYO-only features
// like cloud sync. Expired paid plans fall back to free — same treatment.
func (m *Middleware) RequirePaid(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := auth.UserFromContext(r.Context())
		if user == nil {
			response.Error(w, http.StatusUnauthorized, "authentication required")
			return
		}
		if user.EffectivePlan() == auth.PlanFree {
			response.Error(w, http.StatusPaymentRequired, "this feature requires a Pro or BYO plan")
			return
		}
		next.ServeHTTP(w, r)
	})
}
