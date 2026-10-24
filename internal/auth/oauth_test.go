package auth

import (
	"context"
	"io"
	"log/slog"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/oche/diagram/internal/config"
)

func TestIssueAndVerifyState_RoundTrip(t *testing.T) {
	cookie, value, err := IssueStateCookie("test-secret", false, "lax")
	if err != nil {
		t.Fatalf("issue: %v", err)
	}
	r := httptest.NewRequest("GET", "/", nil)
	r.AddCookie(cookie)
	if err := VerifyState(r, "test-secret", value); err != nil {
		t.Fatalf("verify: %v", err)
	}
}

func TestVerifyState_Tampered(t *testing.T) {
	cookie, value, _ := IssueStateCookie("test-secret", false, "lax")
	r := httptest.NewRequest("GET", "/", nil)
	r.AddCookie(cookie)
	// Flip the last char of the value → HMAC breaks.
	tampered := value[:len(value)-1] + "0"
	if tampered == value {
		tampered = value[:len(value)-1] + "1"
	}
	if err := VerifyState(r, "test-secret", tampered); err == nil {
		t.Fatal("expected mismatch, got nil")
	}
}

func TestVerifyState_MissingCookie(t *testing.T) {
	r := httptest.NewRequest("GET", "/", nil)
	if err := VerifyState(r, "test-secret", "abc.def"); err == nil {
		t.Fatal("expected mismatch without cookie")
	}
}

// ---- FindOrCreateFromOAuth ----

type fakeUsers struct {
	byEmail map[string]*User
	byID    map[string]*User
}

func newFakeUsers() *fakeUsers {
	return &fakeUsers{byEmail: map[string]*User{}, byID: map[string]*User{}}
}

func (f *fakeUsers) Create(_ context.Context, u *User) error {
	f.byEmail[u.Email] = u
	f.byID[u.ID] = u
	return nil
}
func (f *fakeUsers) GetByEmail(_ context.Context, e string) (*User, error) {
	if u, ok := f.byEmail[e]; ok {
		return u, nil
	}
	return nil, nil
}
func (f *fakeUsers) GetByID(_ context.Context, id string) (*User, error) {
	if u, ok := f.byID[id]; ok {
		return u, nil
	}
	return nil, nil
}
func (f *fakeUsers) UpdatePlan(context.Context, string, Plan) error { return nil }
func (f *fakeUsers) UpgradePlan(context.Context, string, Plan, time.Time) error {
	return nil
}
func (f *fakeUsers) UpdatePasswordHash(context.Context, string, string) error { return nil }
func (f *fakeUsers) Delete(_ context.Context, id string) error {
	if u, ok := f.byID[id]; ok {
		delete(f.byEmail, u.Email)
		delete(f.byID, id)
	}
	return nil
}

type fakeSessions struct{}

func (fakeSessions) Create(context.Context, *Session) error          { return nil }
func (fakeSessions) GetByID(context.Context, string) (*Session, error) { return nil, nil }
func (fakeSessions) Delete(context.Context, string) error            { return nil }
func (fakeSessions) DeleteByUserID(context.Context, string) error    { return nil }

type fakeIdentities struct {
	byKey map[string]*Identity
	all   []Identity
}

func newFakeIdentities() *fakeIdentities {
	return &fakeIdentities{byKey: map[string]*Identity{}}
}

func (f *fakeIdentities) GetByProviderSub(_ context.Context, provider, sub string) (*Identity, error) {
	return f.byKey[provider+":"+sub], nil
}
func (f *fakeIdentities) Link(_ context.Context, id *Identity) error {
	f.byKey[id.Provider+":"+id.ProviderSub] = id
	f.all = append(f.all, *id)
	return nil
}
func (f *fakeIdentities) ListForUser(context.Context, string) ([]Identity, error) {
	return f.all, nil
}

func newTestService(mode string) (*Service, *fakeUsers, *fakeIdentities) {
	users := newFakeUsers()
	ids := newFakeIdentities()
	cfg := &config.Config{
		Session: config.SessionConfig{
			Secret:     "x",
			CookieName: "s",
			MaxAge:     60,
		},
		Registration: config.RegistrationConfig{Mode: mode},
	}
	svc := NewService(cfg, users, fakeSessions{}, nil, slog.New(slog.NewTextHandler(io.Discard, nil))).
		WithIdentities(ids)
	return svc, users, ids
}

func TestFindOrCreateFromOAuth_InviteModeRejectsNewUser(t *testing.T) {
	svc, _, _ := newTestService(config.RegistrationInvite)
	_, _, err := svc.FindOrCreateFromOAuth(context.Background(), OAuthClaims{
		Provider:      ProviderGoogle,
		Subject:       "g-123",
		Email:         "new@example.com",
		EmailVerified: true,
	})
	if err != ErrOAuthInviteRequired {
		t.Fatalf("want ErrOAuthInviteRequired, got %v", err)
	}
}

func TestFindOrCreateFromOAuth_OpenModeCreatesUser(t *testing.T) {
	svc, users, ids := newTestService(config.RegistrationOpen)
	u, cookies, err := svc.FindOrCreateFromOAuth(context.Background(), OAuthClaims{
		Provider:      ProviderGitHub,
		Subject:       "gh-999",
		Email:         "fresh@example.com",
		EmailVerified: true,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if u.Email != "fresh@example.com" || u.Plan != PlanFree {
		t.Fatalf("wrong user: %+v", u)
	}
	if len(cookies) != 2 {
		t.Fatalf("expected session + csrf cookies, got %d", len(cookies))
	}
	if _, ok := users.byEmail["fresh@example.com"]; !ok {
		t.Fatal("user not persisted")
	}
	if ids.byKey["github:gh-999"] == nil {
		t.Fatal("identity not linked")
	}
}

func TestFindOrCreateFromOAuth_LinksExistingEmail(t *testing.T) {
	svc, users, ids := newTestService(config.RegistrationInvite)
	// Pre-seed a password-only user.
	users.Create(context.Background(), &User{
		ID:           "u1",
		Email:        "existing@example.com",
		PasswordHash: "hash",
		Role:         RoleUser,
		Plan:         PlanFree,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	})
	u, _, err := svc.FindOrCreateFromOAuth(context.Background(), OAuthClaims{
		Provider:      ProviderGoogle,
		Subject:       "g-7",
		Email:         "existing@example.com",
		EmailVerified: true,
	})
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if u.ID != "u1" {
		t.Fatalf("want linked to u1, got %s", u.ID)
	}
	if ids.byKey["google:g-7"] == nil {
		t.Fatal("link not written")
	}
}

func TestFindOrCreateFromOAuth_RejectsUnverified(t *testing.T) {
	svc, _, _ := newTestService(config.RegistrationOpen)
	_, _, err := svc.FindOrCreateFromOAuth(context.Background(), OAuthClaims{
		Provider:      ProviderGoogle,
		Subject:       "g-1",
		Email:         "x@example.com",
		EmailVerified: false,
	})
	if err != ErrOAuthEmailUnverified {
		t.Fatalf("want ErrOAuthEmailUnverified, got %v", err)
	}
}
