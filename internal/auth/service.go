package auth

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/mail"
	"time"
	"unicode/utf8"

	"golang.org/x/crypto/bcrypt"

	"github.com/oche/diagram/internal/config"
)

var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrEmailTaken         = errors.New("email already registered")
	ErrInvalidEmail       = errors.New("invalid email address")
	ErrWeakPassword       = errors.New("password must be at least 8 characters")
	ErrInvalidSession     = errors.New("invalid or expired session")
	ErrInvalidInvite      = errors.New("This invite code isn't valid or has already been used.")
	ErrAccountLocked      = errors.New("account temporarily locked: too many failed attempts. Try again in ~15 minutes.")
	ErrResetTokenInvalid  = errors.New("this reset link is invalid or has already been used")
	ErrResetTokenExpired  = errors.New("this reset link has expired — request a new one")
)

// InviteValidator is implemented by the invite repository.
// ValidateAndUse returns the plan attached to the invite so the caller can
// apply it to the newly created user.
type InviteValidator interface {
	ValidateAndUse(ctx context.Context, code string, userID string) (plan string, err error)
}

// LoginLockout lets us plug per-account brute-force protection from the
// abuse package without creating an import cycle.
type LoginLockout interface {
	Check(email string) error
	RecordFailure(email string)
	RecordSuccess(email string)
}

// passwordResetClaimer is the atomic-consume side of the reset store.
// Separate interface so existing PasswordResetStore implementations can add
// it without breaking. Implemented by PasswordResetPostgres.
type passwordResetClaimer interface {
	ClaimByTokenHash(ctx context.Context, hash string) (string, error)
}

// PasswordResetStore abstracts the password_resets table for the auth service.
type PasswordResetStore interface {
	Create(ctx context.Context, rec *PasswordResetRecord) error
	GetByTokenHash(ctx context.Context, hash string) (*PasswordResetRecord, error)
	MarkUsed(ctx context.Context, id string) error
	RecentCountForUser(ctx context.Context, userID string, within time.Duration) (int, error)
}

// PasswordResetRecord mirrors repository.PasswordResetRecord so the auth
// package doesn't have to import the repository package directly.
type PasswordResetRecord struct {
	ID        string
	UserID    string
	TokenHash string
	ExpiresAt time.Time
	UsedAt    *time.Time
	CreatedAt time.Time
}

// PasswordResetMailer sends the "reset your password" email. Satisfied by
// invite.EmailService.
type PasswordResetMailer interface {
	SendPasswordReset(toEmail, token string) error
}

// WelcomeMailer greets newly-registered users. Optional; if unwired,
// signup proceeds without sending mail. Satisfied by invite.EmailService.
type WelcomeMailer interface {
	SendWelcome(toEmail string) error
}

// ActivationRecorder is satisfied by *activation.Recorder. Using an
// interface keeps the auth package from importing activation (and vice
// versa — activation doesn't need to know about users).
type ActivationRecorder interface {
	Record(ctx context.Context, userID string, kind string)
}

// AuditSink is the minimal contract for auth audit logging. Wired via
// WithAudit. Local interface to avoid importing internal/audit (and the
// import cycle that would create with future audit-API endpoints).
type AuditSink interface {
	Log(ctx context.Context, event string, fields map[string]any)
}

type Service struct {
	cfg         *config.Config
	users       UserRepository
	sessions    SessionRepository
	identities  IdentityRepository
	invites     InviteValidator
	lockout     LoginLockout
	resetStore  PasswordResetStore
	resetMailer PasswordResetMailer
	welcome     WelcomeMailer
	activation  ActivationRecorder
	audit       AuditSink
	logger      *slog.Logger
}

func NewService(cfg *config.Config, users UserRepository, sessions SessionRepository, invites InviteValidator, logger *slog.Logger) *Service {
	return &Service{
		cfg:      cfg,
		users:    users,
		sessions: sessions,
		invites:  invites,
		logger:   logger,
	}
}

// WithLoginLockout attaches an optional brute-force guard. Returns the service
// so it chains neatly after NewService.
func (s *Service) WithLoginLockout(l LoginLockout) *Service {
	s.lockout = l
	return s
}

// WithIdentities wires the user_identities repository. Required for the
// OAuth flow — omitting it keeps FindOrCreateFromOAuth inert.
func (s *Service) WithIdentities(r IdentityRepository) *Service {
	s.identities = r
	return s
}

// WithPasswordReset wires up the forgot-password flow. Both dependencies are
// required for the flow to work; omitting them makes RequestPasswordReset a no-op.
func (s *Service) WithPasswordReset(store PasswordResetStore, mailer PasswordResetMailer) *Service {
	s.resetStore = store
	s.resetMailer = mailer
	return s
}

// WithWelcomeMailer attaches the welcome-email sender. Optional; without
// it, newly-registered users get no greeting (signup still succeeds).
func (s *Service) WithWelcomeMailer(m WelcomeMailer) *Service {
	s.welcome = m
	return s
}

// sendWelcomeAsync fires a welcome email without blocking signup. Errors
// are logged — a missed welcome is not a signup-failing event.
func (s *Service) sendWelcomeAsync(email string) {
	if s.welcome == nil {
		return
	}
	go func() {
		if err := s.welcome.SendWelcome(email); err != nil {
			s.logger.Warn("welcome email failed", "email", email, "error", err)
		}
	}()
}

// WithAudit attaches the audit recorder. Optional; without it, security
// events (login, password reset) are not persisted to audit_events.
func (s *Service) WithAudit(a AuditSink) *Service {
	s.audit = a
	return s
}

// WithActivation attaches the activation recorder. Optional; without it,
// signup / first-diagram / first-export events aren't recorded.
func (s *Service) WithActivation(r ActivationRecorder) *Service {
	s.activation = r
	return s
}

func (s *Service) Register(ctx context.Context, req RegisterRequest) (*User, *http.Cookie, error) {
	if err := validateEmail(req.Email); err != nil {
		return nil, nil, err
	}
	if err := validatePassword(req.Password); err != nil {
		return nil, nil, err
	}
	openMode := s.cfg != nil && s.cfg.Registration.Mode == config.RegistrationOpen
	if !openMode && req.InviteCode == "" {
		return nil, nil, ErrInvalidInvite
	}

	existing, _ := s.users.GetByEmail(ctx, req.Email)
	if existing != nil {
		return nil, nil, ErrEmailTaken
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), 12)
	if err != nil {
		return nil, nil, fmt.Errorf("hashing password: %w", err)
	}

	user := &User{
		ID:           generateID(),
		Email:        req.Email,
		PasswordHash: string(hash),
		Role:         RoleUser,
		Plan:         PlanFree,
		CreatedAt:    time.Now().UTC(),
		UpdatedAt:    time.Now().UTC(),
	}

	if err := s.users.Create(ctx, user); err != nil {
		return nil, nil, fmt.Errorf("creating user: %w", err)
	}

	// Validate and consume the invite code. Audit finding #2: if we fail
	// here we must delete the user row we just created, otherwise the
	// caller retries with a valid code and hits "email taken" forever.
	//
	// The invite claim itself is atomic (single UPDATE...RETURNING in the
	// invite repo — audit #3) so two concurrent registrations can't both
	// win the same code.
	//
	// Open registration mode skips the invite check entirely — new users
	// land on the free plan.
	if s.invites != nil && req.InviteCode != "" {
		invitePlan, err := s.invites.ValidateAndUse(ctx, req.InviteCode, user.ID)
		if err != nil {
			// Compensate — drop the orphan user so the email is free again.
			// Best-effort: if this fails the user is stranded, but the audit
			// log below flags it for manual cleanup.
			if delErr := s.users.Delete(ctx, user.ID); delErr != nil {
				s.logger.Error("orphan user cleanup failed after invite reject",
					"user_id", user.ID, "email", user.Email, "error", delErr)
			}
			return nil, nil, ErrInvalidInvite
		}
		if plan := NormalizePlan(invitePlan); plan != PlanFree {
			if err := s.users.UpdatePlan(ctx, user.ID, plan); err != nil {
				s.logger.Error("failed to apply invite plan", "user_id", user.ID, "plan", plan, "error", err)
			} else {
				user.Plan = plan
			}
		}
	}

	s.logger.Info("user registered", "user_id", user.ID, "email", user.Email)
	if s.activation != nil {
		s.activation.Record(ctx, user.ID, "signup")
	}
	s.sendWelcomeAsync(user.Email)

	cookie, err := s.createSession(ctx, user)
	if err != nil {
		return nil, nil, err
	}

	return user, cookie, nil
}

func (s *Service) Login(ctx context.Context, req LoginRequest) (*User, *http.Cookie, error) {
	if s.lockout != nil {
		if err := s.lockout.Check(req.Email); err != nil {
			return nil, nil, ErrAccountLocked
		}
	}

	user, err := s.users.GetByEmail(ctx, req.Email)
	if err != nil {
		if s.lockout != nil {
			s.lockout.RecordFailure(req.Email)
		}
		return nil, nil, ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		if s.lockout != nil {
			s.lockout.RecordFailure(req.Email)
		}
		return nil, nil, ErrInvalidCredentials
	}

	if s.lockout != nil {
		s.lockout.RecordSuccess(req.Email)
	}
	s.logger.Info("user logged in", "user_id", user.ID)
	if s.audit != nil {
		s.audit.Log(ctx, "auth.login_success", map[string]any{
			"actor_user_id": user.ID,
			"target_id":     user.ID,
			"method":        "password",
		})
	}

	cookie, err := s.createSession(ctx, user)
	if err != nil {
		return nil, nil, err
	}

	return user, cookie, nil
}

func (s *Service) Logout(ctx context.Context, r *http.Request) []*http.Cookie {
	cookie, err := r.Cookie(s.cfg.Session.CookieName)
	if err == nil {
		sessionID := s.verifySessionToken(cookie.Value)
		if sessionID != "" {
			s.sessions.Delete(ctx, sessionID)
		}
	}

	// Expire both the session and CSRF cookies.
	return []*http.Cookie{
		{
			Name:     s.cfg.Session.CookieName,
			Value:    "",
			Path:     "/",
			MaxAge:   -1,
			HttpOnly: s.cfg.Session.HttpOnly,
			Secure:   s.cfg.Session.Secure,
			SameSite: parseSameSite(s.cfg.Session.SameSite),
		},
		{
			Name:     s.cfg.Session.CookieName + "_csrf",
			Value:    "",
			Path:     "/",
			MaxAge:   -1,
			HttpOnly: false,
			Secure:   s.cfg.Session.Secure,
			SameSite: parseSameSite(s.cfg.Session.SameSite),
		},
	}
}

// passwordResetTTL bounds how long a reset link works for. Kept small because
// email delivery is usually fast and a long window widens the blast radius
// of a compromised inbox.
const passwordResetTTL = time.Hour

// passwordResetThrottleWindow / passwordResetThrottleMax cap how many reset
// tokens a single account can mint — silently — so an attacker who floods
// the endpoint for a known email can't burn our Resend quota.
const (
	passwordResetThrottleWindow = 15 * time.Minute
	passwordResetThrottleMax    = 5
)

// RequestPasswordReset starts the forgot-password flow.
//
// To prevent email enumeration we ALWAYS return nil: the HTTP handler
// responds identically whether or not `email` matches a real account.
// If the email doesn't exist (or the flow isn't configured), we just no-op.
func (s *Service) RequestPasswordReset(ctx context.Context, email string) error {
	if s.resetStore == nil || s.resetMailer == nil {
		// Feature wasn't wired up in main.go — silently succeed so the client
		// still gets the indistinguishable response.
		return nil
	}

	// Normalise (users table already stores lower-cased emails on insert).
	user, err := s.users.GetByEmail(ctx, email)
	if err != nil || user == nil {
		// Unknown email — log at Debug so it's visible for forensics but
		// not returned to the caller.
		s.logger.Debug("password reset requested for unknown email", "email_hash", hashForLog(email))
		return nil
	}

	// Throttle per-user requests so a known email can't be spammed.
	count, err := s.resetStore.RecentCountForUser(ctx, user.ID, passwordResetThrottleWindow)
	if err == nil && count >= passwordResetThrottleMax {
		s.logger.Warn("password reset throttled", "user_id", user.ID, "count", count)
		return nil
	}

	token, hash, err := generatePasswordResetToken()
	if err != nil {
		s.logger.Error("password reset token gen failed", "error", err)
		return nil
	}

	rec := &PasswordResetRecord{
		ID:        generateID(),
		UserID:    user.ID,
		TokenHash: hash,
		ExpiresAt: time.Now().UTC().Add(passwordResetTTL),
		CreatedAt: time.Now().UTC(),
	}
	if err := s.resetStore.Create(ctx, rec); err != nil {
		s.logger.Error("password reset insert failed", "error", err)
		return nil
	}

	// Fire-and-forget the email — a failing send shouldn't leak status to the client.
	if err := s.resetMailer.SendPasswordReset(user.Email, token); err != nil {
		s.logger.Error("password reset email failed", "user_id", user.ID, "error", err)
	} else {
		s.logger.Info("password reset email queued", "user_id", user.ID)
	}
	if s.audit != nil {
		s.audit.Log(ctx, "auth.password_reset_requested", map[string]any{
			"actor_user_id": user.ID,
			"target_id":     user.ID,
		})
	}
	return nil
}

// ResetPassword consumes a reset token and replaces the user's password.
// On success all existing sessions for that user are invalidated.
func (s *Service) ResetPassword(ctx context.Context, token, newPassword string) error {
	if s.resetStore == nil {
		return ErrResetTokenInvalid
	}
	if err := validatePassword(newPassword); err != nil {
		return err
	}
	if token == "" {
		return ErrResetTokenInvalid
	}

	// Consume the token atomically BEFORE touching the password. Audit fix:
	// the previous flow ran UpdatePasswordHash first and MarkUsed second,
	// meaning two concurrent /reset-password calls with the same token both
	// changed the password (and one silently failed to mark it used).
	//
	// ClaimByTokenHash is a single UPDATE...RETURNING that only returns a
	// row when the token is unused + unexpired; the second concurrent call
	// finds no row and errors out. Falls back to the legacy two-step path
	// when the store impl doesn't support atomic claim (e.g. old tests).
	var userID string
	if claimer, ok := s.resetStore.(passwordResetClaimer); ok {
		id, err := claimer.ClaimByTokenHash(ctx, hashPasswordResetToken(token))
		if err != nil {
			return fmt.Errorf("claiming reset token: %w", err)
		}
		if id == "" {
			return ErrResetTokenInvalid
		}
		userID = id
	} else {
		rec, err := s.resetStore.GetByTokenHash(ctx, hashPasswordResetToken(token))
		if err != nil {
			return ErrResetTokenInvalid
		}
		if rec == nil || rec.UsedAt != nil {
			return ErrResetTokenInvalid
		}
		if time.Now().After(rec.ExpiresAt) {
			return ErrResetTokenExpired
		}
		if err := s.resetStore.MarkUsed(ctx, rec.ID); err != nil {
			return fmt.Errorf("marking reset used: %w", err)
		}
		userID = rec.UserID
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), 12)
	if err != nil {
		return fmt.Errorf("hashing new password: %w", err)
	}
	if err := s.users.UpdatePasswordHash(ctx, userID, string(hash)); err != nil {
		// Token is already consumed; user needs a fresh reset link. Log the
		// user_id so operators can flag if this correlates with an outage.
		s.logger.Error("password update failed after token claim",
			"user_id", userID, "error", err)
		return err
	}
	// Belt-and-braces: kick every existing session for this user.
	if err := s.sessions.DeleteByUserID(ctx, userID); err != nil {
		s.logger.Warn("failed to invalidate sessions after reset", "user_id", userID, "error", err)
	}
	// Clear any lockout counter — the user just proved control of the email.
	if s.lockout != nil {
		if user, _ := s.users.GetByID(ctx, userID); user != nil {
			s.lockout.RecordSuccess(user.Email)
		}
	}
	s.logger.Info("password reset completed", "user_id", userID)
	if s.audit != nil {
		s.audit.Log(ctx, "auth.password_reset_completed", map[string]any{
			"actor_user_id": userID,
			"target_id":     userID,
		})
	}
	return nil
}

// generatePasswordResetToken returns (urlSafeToken, sha256HexHash, err).
// The URL-safe token goes in the email link; only the hash is persisted.
func generatePasswordResetToken() (string, string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", "", err
	}
	token := hex.EncodeToString(b)
	return token, hashPasswordResetToken(token), nil
}

// hashPasswordResetToken gives us a deterministic lookup key for the DB.
// SHA-256 is the right tool here: fast, indexable, and a token leak through
// a DB read still doesn't let an attacker use the token (they need the preimage).
func hashPasswordResetToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

// hashForLog produces an opaque short tag for sensitive strings we want
// visible in logs without storing the plaintext.
func hashForLog(s string) string {
	sum := sha256.Sum256([]byte(s))
	return hex.EncodeToString(sum[:8])
}

// CSRFCookieName returns the cookie name used by the CSRF double-submit scheme.
// The frontend reads it via document.cookie and echoes it as X-CSRF-Token.
func (s *Service) CSRFCookieName() string {
	return s.cfg.Session.CookieName + "_csrf"
}

// IssueCSRFCookie mints a fresh CSRF token cookie. Call this after createSession
// in the handler layer so login/register responses get both cookies.
func (s *Service) IssueCSRFCookie() *http.Cookie {
	b := make([]byte, 24)
	_, _ = rand.Read(b)
	token := hex.EncodeToString(b)
	return &http.Cookie{
		Name:     s.CSRFCookieName(),
		Value:    token,
		Path:     "/",
		MaxAge:   s.cfg.Session.MaxAge,
		HttpOnly: false, // JS must read this
		Secure:   s.cfg.Session.Secure,
		SameSite: parseSameSite(s.cfg.Session.SameSite),
	}
}

// ValidateSession checks the session cookie and returns the associated user info.
func (s *Service) ValidateSession(ctx context.Context, r *http.Request) (*SessionUser, error) {
	cookie, err := r.Cookie(s.cfg.Session.CookieName)
	if err != nil {
		return nil, ErrInvalidSession
	}

	sessionID := s.verifySessionToken(cookie.Value)
	if sessionID == "" {
		return nil, ErrInvalidSession
	}

	session, err := s.sessions.GetByID(ctx, sessionID)
	if err != nil {
		return nil, ErrInvalidSession
	}

	if time.Now().After(session.ExpiresAt) {
		s.sessions.Delete(ctx, session.ID)
		return nil, ErrInvalidSession
	}

	user, err := s.users.GetByID(ctx, session.UserID)
	if err != nil {
		return nil, ErrInvalidSession
	}

	return &SessionUser{
		UserID:     user.ID,
		Email:      user.Email,
		Role:       user.Role,
		Plan:       user.Plan,
		PlanExpiry: user.PlanExpiry,
	}, nil
}

func (s *Service) createSession(ctx context.Context, user *User) (*http.Cookie, error) {
	sessionID := generateID()
	now := time.Now().UTC()
	maxAge := time.Duration(s.cfg.Session.MaxAge) * time.Second

	session := &Session{
		ID:        sessionID,
		UserID:    user.ID,
		ExpiresAt: now.Add(maxAge),
		CreatedAt: now,
	}

	if err := s.sessions.Create(ctx, session); err != nil {
		return nil, fmt.Errorf("creating session: %w", err)
	}

	// Token = sessionID.HMAC(sessionID) — prevents ID guessing
	token := s.signSessionToken(sessionID)

	cookie := &http.Cookie{
		Name:     s.cfg.Session.CookieName,
		Value:    token,
		Path:     "/",
		MaxAge:   s.cfg.Session.MaxAge,
		HttpOnly: s.cfg.Session.HttpOnly,
		Secure:   s.cfg.Session.Secure,
		SameSite: parseSameSite(s.cfg.Session.SameSite),
	}

	return cookie, nil
}

// signSessionToken creates a tamper-proof token: sessionID.hmac
func (s *Service) signSessionToken(sessionID string) string {
	mac := hmac.New(sha256.New, []byte(s.cfg.Session.Secret))
	mac.Write([]byte(sessionID))
	sig := hex.EncodeToString(mac.Sum(nil))
	return sessionID + "." + sig
}

// verifySessionToken validates the HMAC and returns the sessionID, or "" if invalid.
func (s *Service) verifySessionToken(token string) string {
	// Token format: sessionID.hmacSignature
	dotIdx := -1
	for i := len(token) - 1; i >= 0; i-- {
		if token[i] == '.' {
			dotIdx = i
			break
		}
	}
	if dotIdx < 1 {
		return ""
	}

	sessionID := token[:dotIdx]
	providedSig := token[dotIdx+1:]

	mac := hmac.New(sha256.New, []byte(s.cfg.Session.Secret))
	mac.Write([]byte(sessionID))
	expectedSig := hex.EncodeToString(mac.Sum(nil))

	if !hmac.Equal([]byte(providedSig), []byte(expectedSig)) {
		return ""
	}

	return sessionID
}

func parseSameSite(s string) http.SameSite {
	switch s {
	case "strict":
		return http.SameSiteStrictMode
	case "none":
		return http.SameSiteNoneMode
	default:
		return http.SameSiteLaxMode
	}
}

func validateEmail(email string) error {
	if _, err := mail.ParseAddress(email); err != nil {
		return ErrInvalidEmail
	}
	return nil
}

func validatePassword(password string) error {
	if utf8.RuneCountInString(password) < 8 {
		return ErrWeakPassword
	}
	return nil
}

func generateID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}
