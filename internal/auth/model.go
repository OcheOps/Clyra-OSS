package auth

import "time"

type Role string

const (
	RoleUser  Role = "user"
	RoleAdmin Role = "admin"
)

// Plan determines per-user daily usage limits. See config.AbuseConfig.PlanLimits.
type Plan string

const (
	PlanFree Plan = "free"
	PlanPro  Plan = "pro"
	PlanBYO  Plan = "byo"
)

// NormalizePlan returns a known plan or falls back to PlanFree for unknown values.
func NormalizePlan(s string) Plan {
	switch Plan(s) {
	case PlanFree, PlanPro, PlanBYO:
		return Plan(s)
	default:
		return PlanFree
	}
}

type User struct {
	ID           string     `json:"id"`
	Email        string     `json:"email"`
	PasswordHash string     `json:"-"`
	Role         Role       `json:"role"`
	Plan         Plan       `json:"plan"`
	PlanExpiry   *time.Time `json:"plan_expiry,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

// EffectivePlan returns the plan the user should actually be treated as right
// now. A paid plan whose expiry has passed automatically degrades to free.
// A nil expiry means the plan has no time limit.
func (u *User) EffectivePlan() Plan {
	if u == nil {
		return PlanFree
	}
	if u.PlanExpiry != nil && time.Now().After(*u.PlanExpiry) {
		return PlanFree
	}
	return u.Plan
}

// Expired reports whether the user's paid plan has lapsed.
func (u *User) Expired() bool {
	return u != nil && u.PlanExpiry != nil && time.Now().After(*u.PlanExpiry)
}

// Session represents a server-side session tied to a cookie.
type Session struct {
	ID        string
	UserID    string
	ExpiresAt time.Time
	CreatedAt time.Time
}

// SessionUser is the user data stored in context for the duration of a request.
type SessionUser struct {
	UserID     string
	Email      string
	Role       Role
	Plan       Plan
	PlanExpiry *time.Time
}

// EffectivePlan returns the current-effective plan (expired = free).
func (s *SessionUser) EffectivePlan() Plan {
	if s == nil {
		return PlanFree
	}
	if s.PlanExpiry != nil && time.Now().After(*s.PlanExpiry) {
		return PlanFree
	}
	return s.Plan
}

// Request/response types

type RegisterRequest struct {
	Email      string `json:"email"       validate:"required,email"`
	Password   string `json:"password"    validate:"required,min=8"`
	InviteCode string `json:"invite_code"`
}

type LoginRequest struct {
	Email    string `json:"email"    validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type ForgotPasswordRequest struct {
	Email string `json:"email" validate:"required,email"`
}

type ResetPasswordRequest struct {
	Token    string `json:"token"    validate:"required,min=32,max=200"`
	Password string `json:"password" validate:"required,min=8"`
}

type AuthResponse struct {
	User *UserProfile `json:"user"`
}

type UserProfile struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Role  Role   `json:"role"`
	// Plan is the raw value in the users table — persists past expiry so the
	// UI can show "Your Pro plan lapsed on Aug 5" copy. Use EffectivePlan for
	// enforcement decisions; use Plan + PlanExpiry for display context.
	Plan       Plan       `json:"plan"`
	PlanExpiry *time.Time `json:"plan_expiry,omitempty"`
	// EffectivePlan is what the backend actually enforces right now — free if
	// expiry has passed regardless of the raw Plan. Frontend must gate off
	// this, not Plan, otherwise display and 402 responses disagree.
	EffectivePlan Plan `json:"effective_plan"`
}
