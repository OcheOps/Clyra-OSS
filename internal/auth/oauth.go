package auth

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/oche/diagram/internal/config"
)

// OAuth error sentinels. All callback errors redirect to /login with an
// ?oauth_error= query param — the SPA maps each code to a user-facing toast.
var (
	ErrOAuthStateMismatch     = errors.New("oauth state mismatch")
	ErrOAuthEmailUnverified   = errors.New("oauth email unverified")
	ErrOAuthInviteRequired    = errors.New("invite required")
	ErrOAuthProviderDisabled  = errors.New("oauth provider disabled")
	ErrOAuthUpstreamFailure   = errors.New("oauth upstream failed")
)

const (
	ProviderGoogle = "google"
	ProviderGitHub = "github"

	oauthStateCookie   = "diagram_oauth_state"
	oauthStateMaxAge   = 600 // 10 min — plenty for the redirect round trip
	oauthStateRandSize = 24
)

// OAuthClaims is the provider-agnostic payload passed to FindOrCreateFromOAuth.
// It carries what both Google and GitHub give us after an identity lookup.
type OAuthClaims struct {
	Provider      string // ProviderGoogle | ProviderGitHub
	Subject       string // provider's stable user id (Google "sub", GitHub numeric id as string)
	Email         string
	EmailVerified bool
}

// IssueStateCookie mints a short-lived HMAC-signed state value and returns
// both the cookie (to set on the outbound redirect) and the value to embed
// in the provider's authorize URL. The double-submit check in VerifyState
// catches CSRF on the callback without any server-side state.
func IssueStateCookie(secret string, secure bool, sameSite string) (*http.Cookie, string, error) {
	b := make([]byte, oauthStateRandSize)
	if _, err := rand.Read(b); err != nil {
		return nil, "", fmt.Errorf("state rand: %w", err)
	}
	nonce := hex.EncodeToString(b)
	sig := signState(nonce, secret)
	value := nonce + "." + sig
	cookie := &http.Cookie{
		Name:     oauthStateCookie,
		Value:    value,
		Path:     "/api/v1/auth",
		MaxAge:   oauthStateMaxAge,
		HttpOnly: true,
		Secure:   secure,
		SameSite: parseSameSite(sameSite),
	}
	return cookie, value, nil
}

// VerifyState runs the double-submit check: the state param on the callback
// URL must match the cookie, and the HMAC signature must verify.
func VerifyState(r *http.Request, secret string, param string) error {
	cookie, err := r.Cookie(oauthStateCookie)
	if err != nil {
		return ErrOAuthStateMismatch
	}
	if !hmac.Equal([]byte(cookie.Value), []byte(param)) {
		return ErrOAuthStateMismatch
	}
	idx := strings.LastIndex(param, ".")
	if idx < 1 {
		return ErrOAuthStateMismatch
	}
	nonce, sig := param[:idx], param[idx+1:]
	if !hmac.Equal([]byte(sig), []byte(signState(nonce, secret))) {
		return ErrOAuthStateMismatch
	}
	return nil
}

// ClearStateCookie expires the state cookie — call this on the callback so
// a replay with the same state after success is a no-op.
func ClearStateCookie(secure bool, sameSite string) *http.Cookie {
	return &http.Cookie{
		Name:     oauthStateCookie,
		Value:    "",
		Path:     "/api/v1/auth",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   secure,
		SameSite: parseSameSite(sameSite),
	}
}

func signState(nonce, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(nonce))
	return hex.EncodeToString(mac.Sum(nil))
}

// FindOrCreateFromOAuth is the one place where account linking happens.
// Rules (keep in sync with brand.md / privacy page):
//   1. Known (provider, sub) → log that user in.
//   2. Unknown (provider, sub) but email already registered (any method) →
//      auto-link the new identity to that existing account. Safe because
//      the provider has verified the email; we don't trust unverified ones.
//   3. Unknown email → create a fresh user on the free plan, but only if
//      REGISTRATION_MODE=open. In invite mode we bounce to /login with
//      invite_required so new-user enrollment stays behind the gate.
//
// Returns the resolved user and the cookies (session + CSRF) for the handler
// to set before the final redirect to /dashboard.
func (s *Service) FindOrCreateFromOAuth(ctx context.Context, claims OAuthClaims) (*User, []*http.Cookie, error) {
	if s.identities == nil {
		return nil, nil, fmt.Errorf("identity repository not wired")
	}
	if !claims.EmailVerified {
		return nil, nil, ErrOAuthEmailUnverified
	}
	claims.Email = strings.ToLower(strings.TrimSpace(claims.Email))
	if claims.Email == "" || claims.Subject == "" {
		return nil, nil, ErrOAuthUpstreamFailure
	}

	// Path 1 — known identity.
	if existing, err := s.identities.GetByProviderSub(ctx, claims.Provider, claims.Subject); err == nil && existing != nil {
		user, err := s.users.GetByID(ctx, existing.UserID)
		if err != nil {
			return nil, nil, fmt.Errorf("loading linked user: %w", err)
		}
		cookies, err := s.issueAuthCookies(ctx, user)
		if err != nil {
			return nil, nil, err
		}
		s.logger.Info("oauth login", "provider", claims.Provider, "user_id", user.ID)
		return user, cookies, nil
	}

	// Path 2 — email already registered; auto-link.
	if existing, _ := s.users.GetByEmail(ctx, claims.Email); existing != nil {
		if err := s.identities.Link(ctx, &Identity{
			ID:          generateID(),
			UserID:      existing.ID,
			Provider:    claims.Provider,
			ProviderSub: claims.Subject,
			Email:       claims.Email,
			CreatedAt:   time.Now().UTC(),
		}); err != nil {
			return nil, nil, fmt.Errorf("linking identity: %w", err)
		}
		cookies, err := s.issueAuthCookies(ctx, existing)
		if err != nil {
			return nil, nil, err
		}
		s.logger.Info("oauth linked existing account", "provider", claims.Provider, "user_id", existing.ID)
		return existing, cookies, nil
	}

	// Path 3 — brand new user. Gated on REGISTRATION_MODE.
	if s.cfg == nil || s.cfg.Registration.Mode != config.RegistrationOpen {
		return nil, nil, ErrOAuthInviteRequired
	}

	now := time.Now().UTC()
	newUser := &User{
		ID:        generateID(),
		Email:     claims.Email,
		Role:      RoleUser,
		Plan:      PlanFree,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := s.users.Create(ctx, newUser); err != nil {
		return nil, nil, fmt.Errorf("creating oauth user: %w", err)
	}
	if err := s.identities.Link(ctx, &Identity{
		ID:          generateID(),
		UserID:      newUser.ID,
		Provider:    claims.Provider,
		ProviderSub: claims.Subject,
		Email:       claims.Email,
		CreatedAt:   now,
	}); err != nil {
		// The user row survives — a retry will hit path 2 and link cleanly.
		return nil, nil, fmt.Errorf("linking new identity: %w", err)
	}
	cookies, err := s.issueAuthCookies(ctx, newUser)
	if err != nil {
		return nil, nil, err
	}
	s.logger.Info("oauth new user", "provider", claims.Provider, "user_id", newUser.ID, "email", newUser.Email)
	if s.activation != nil {
		s.activation.Record(ctx, newUser.ID, "signup")
	}
	s.sendWelcomeAsync(newUser.Email)
	return newUser, cookies, nil
}

// issueAuthCookies creates the session + CSRF cookie pair in one call so
// OAuth callbacks land the user in the exact same cookie state as a
// password login.
func (s *Service) issueAuthCookies(ctx context.Context, user *User) ([]*http.Cookie, error) {
	sessionCookie, err := s.createSession(ctx, user)
	if err != nil {
		return nil, err
	}
	return []*http.Cookie{sessionCookie, s.IssueCSRFCookie()}, nil
}
