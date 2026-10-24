package abuse

import (
	"log/slog"
	"net/http"

	"github.com/oche/diagram/internal/auth"
	"github.com/oche/diagram/pkg/response"
)

// Middleware holds the abuse prevention services used by HTTP middleware.
type Middleware struct {
	userRateLimiter UserRateLimiter
	quotaService    *QuotaService
	promptValidator *PromptValidator
	logger          *slog.Logger
}

func NewMiddleware(
	userRateLimiter UserRateLimiter,
	quotaService *QuotaService,
	promptValidator *PromptValidator,
	logger *slog.Logger,
) *Middleware {
	return &Middleware{
		userRateLimiter: userRateLimiter,
		quotaService:    quotaService,
		promptValidator: promptValidator,
		logger:          logger,
	}
}

// UserRateLimit enforces per-user rate limiting on generation endpoints.
// Must be applied AFTER authentication middleware (needs user in context).
func (m *Middleware) UserRateLimit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := auth.UserFromContext(r.Context())
		if user == nil {
			// No user in context — let the auth middleware handle it
			next.ServeHTTP(w, r)
			return
		}

		if !m.userRateLimiter.Allow(r.Context(), user.UserID) {
			if m.logger != nil {
				m.logger.Warn("user rate limit exceeded",
					"user_id", user.UserID,
					"email", user.Email,
				)
			}
			response.Error(w, http.StatusTooManyRequests, "rate_limit_exceeded")
			return
		}

		next.ServeHTTP(w, r)
	})
}

// UsageQuota enforces daily usage limits on generation endpoints.
// Must be applied AFTER authentication middleware.
// BYO providers (user-supplied API key) bypass this check.
func (m *Middleware) UsageQuota(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := auth.UserFromContext(r.Context())
		if user == nil {
			next.ServeHTTP(w, r)
			return
		}

		effectivePlan := string(user.EffectivePlan())
		if err := m.quotaService.Check(r.Context(), user.UserID, effectivePlan); err != nil {
			if m.logger != nil {
				m.logger.Warn("daily limit reached",
					"user_id", user.UserID,
					"email", user.Email,
					"plan", user.Plan,
					"effective_plan", effectivePlan,
				)
			}
			response.Error(w, http.StatusTooManyRequests, "daily_limit_reached")
			return
		}

		next.ServeHTTP(w, r)
	})
}
