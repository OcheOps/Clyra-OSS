package auth

import "context"

type contextKey string

const sessionUserKey contextKey = "session_user"

// NewContext stores the session user in the context.
func NewContext(ctx context.Context, user *SessionUser) context.Context {
	return context.WithValue(ctx, sessionUserKey, user)
}

// UserFromContext retrieves the session user from the context.
// Returns nil if no user is present.
func UserFromContext(ctx context.Context) *SessionUser {
	user, _ := ctx.Value(sessionUserKey).(*SessionUser)
	return user
}
