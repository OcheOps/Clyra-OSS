// Package account exposes read-only user account endpoints.
package account

import (
	"net/http"

	"github.com/oche/diagram/internal/abuse"
	"github.com/oche/diagram/internal/auth"
	"github.com/oche/diagram/pkg/response"
)

type Handler struct {
	quota *abuse.QuotaService
}

func NewHandler(quota *abuse.QuotaService) *Handler {
	return &Handler{quota: quota}
}

// Usage returns the current plan + remaining-today for the authenticated user.
//
// Response:
//
//	{
//	  "plan": "pro",
//	  "effective_plan": "free",      // pro expired → treated as free
//	  "plan_expiry": "2026-05-01...",
//	  "daily_limit": 10,
//	  "daily_remaining": 7
//	}
func (h *Handler) Usage(w http.ResponseWriter, r *http.Request) {
	user := auth.UserFromContext(r.Context())
	if user == nil {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	effective := user.EffectivePlan()
	limit := h.quota.LimitFor(string(effective))
	remaining, _ := h.quota.Remaining(r.Context(), user.UserID, string(effective))

	used := limit - remaining
	if used < 0 {
		used = 0
	}

	response.JSON(w, http.StatusOK, map[string]any{
		"plan":            user.Plan,
		"effective_plan":  effective,
		"plan_expiry":     user.PlanExpiry,
		"daily_limit":     limit,
		"daily_used":      used,
		"daily_remaining": remaining,
	})
}
