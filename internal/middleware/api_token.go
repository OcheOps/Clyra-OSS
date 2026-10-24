package middleware

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/oche/diagram/internal/apitoken"
	"github.com/oche/diagram/internal/auth"
	"github.com/oche/diagram/pkg/response"
)

// APITokenAuthenticator is the subset of the api-token subsystem the
// middleware needs. Kept minimal so tests can pass a fake without wiring
// the whole apitoken package.
type APITokenAuthenticator interface {
	GetByHash(ctx context.Context, hash string) (*apitoken.Token, error)
	TouchLastUsed(ctx context.Context, id string) error
}

// UserResolver looks up a full User by ID so we can build the SessionUser
// context downstream handlers already expect (Plan / PlanExpiry drive
// billing gates). Implemented by the existing user repo.
type UserResolver interface {
	GetByID(ctx context.Context, id string) (*auth.User, error)
}

// AuthenticateAPIToken enforces Bearer clyra_pat_* on protected MCP routes.
// Populates the same auth.SessionUser context the session-cookie path uses,
// so downstream code (RequirePaid, quota gates, audit logs) is oblivious to
// whether the caller is a browser or an MCP client.
//
// Contract:
//   - Missing / malformed header  → 401
//   - Token not found or revoked  → 401  (indistinguishable from missing)
//   - Token valid, user vanished  → 401  (defence in depth)
//   - Valid + user present        → context populated, next handler runs,
//     last_used stamped asynchronously
func AuthenticateAPIToken(tokens APITokenAuthenticator, users UserResolver) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			clear := extractBearer(r)
			if clear == "" {
				response.Error(w, http.StatusUnauthorized, "missing bearer token")
				return
			}
			hash, err := apitoken.HashForLookup(clear)
			if err != nil {
				response.Error(w, http.StatusUnauthorized, "invalid token")
				return
			}
			tok, err := tokens.GetByHash(r.Context(), hash)
			if err != nil || tok == nil {
				response.Error(w, http.StatusUnauthorized, "invalid or revoked token")
				return
			}
			user, err := users.GetByID(r.Context(), tok.UserID)
			if err != nil || user == nil {
				response.Error(w, http.StatusUnauthorized, "invalid token")
				return
			}
			su := &auth.SessionUser{
				UserID:     user.ID,
				Email:      user.Email,
				Role:       user.Role,
				Plan:       user.Plan,
				PlanExpiry: user.PlanExpiry,
			}
			ctx := auth.NewContext(r.Context(), su)

			// Stamp last_used asynchronously — a 1-3ms DB write per request
			// isn't worth blocking the handler for, and a failure just
			// means the UI shows stale freshness. Detached context so the
			// request context canceling (client hang up) doesn't abort
			// the update.
			go func(tokenID string) {
				ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
				defer cancel()
				_ = tokens.TouchLastUsed(ctx, tokenID)
			}(tok.ID)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// extractBearer parses "Authorization: Bearer <token>", tolerating any
// case in the scheme and extra whitespace. Returns "" when absent.
func extractBearer(r *http.Request) string {
	h := r.Header.Get("Authorization")
	if h == "" {
		return ""
	}
	const prefix = "bearer "
	if len(h) < len(prefix) || !strings.EqualFold(h[:len(prefix)], prefix) {
		return ""
	}
	return strings.TrimSpace(h[len(prefix):])
}
