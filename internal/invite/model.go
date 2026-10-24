package invite

import "time"

type InviteCode struct {
	ID        string     `json:"id"`
	Code      string     `json:"code"`
	CreatedBy string     `json:"created_by"`
	UsedBy    *string    `json:"used_by,omitempty"`
	UsedAt    *time.Time `json:"used_at,omitempty"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
	Plan      string     `json:"plan"`
	CreatedAt time.Time  `json:"created_at"`
}

func (i *InviteCode) IsUsed() bool {
	return i.UsedBy != nil
}

func (i *InviteCode) IsExpired() bool {
	if i.ExpiresAt == nil {
		return false
	}
	return time.Now().After(*i.ExpiresAt)
}

func (i *InviteCode) IsValid() bool {
	return !i.IsUsed() && !i.IsExpired()
}

// normalizePlan coerces a user-supplied plan string to a known value.
// Unknown or empty plans fall back to "free".
func normalizePlan(s string) string {
	switch s {
	case "free", "pro", "byo":
		return s
	default:
		return "free"
	}
}
