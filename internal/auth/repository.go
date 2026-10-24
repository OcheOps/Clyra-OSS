package auth

import (
	"context"
	"time"
)

// UserRepository defines the data access contract for user operations.
type UserRepository interface {
	Create(ctx context.Context, user *User) error
	GetByEmail(ctx context.Context, email string) (*User, error)
	GetByID(ctx context.Context, id string) (*User, error)
	UpdatePlan(ctx context.Context, userID string, plan Plan) error
	UpgradePlan(ctx context.Context, userID string, plan Plan, expiry time.Time) error
	UpdatePasswordHash(ctx context.Context, userID, hash string) error
	// Delete removes a user row by ID. Used as compensation when a
	// multi-step register flow fails after Create (invite claim rejected,
	// plan assignment failed, etc.) so a bad-invite attempt doesn't strand
	// the user's email as taken forever.
	Delete(ctx context.Context, userID string) error
}

// Identity represents a single OAuth provider linkage for a user. A user can
// have zero (password-only), one, or many — e.g. both Google + GitHub linked.
type Identity struct {
	ID          string
	UserID      string
	Provider    string // "google" | "github"
	ProviderSub string
	Email       string
	CreatedAt   time.Time
}

// IdentityRepository handles the user_identities table. Kept in the auth
// package so Service can call it without a repository import cycle.
type IdentityRepository interface {
	GetByProviderSub(ctx context.Context, provider, sub string) (*Identity, error)
	Link(ctx context.Context, id *Identity) error
	ListForUser(ctx context.Context, userID string) ([]Identity, error)
}

// SessionRepository defines the data access contract for session operations.
type SessionRepository interface {
	Create(ctx context.Context, session *Session) error
	GetByID(ctx context.Context, id string) (*Session, error)
	Delete(ctx context.Context, id string) error
	DeleteByUserID(ctx context.Context, userID string) error
}
