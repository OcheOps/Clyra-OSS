package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/oche/diagram/internal/config"
)

// OAuthHandler wires the provider-agnostic HTTP endpoints. Routes are
// registered per provider (`/auth/google/{start,callback}`,
// `/auth/github/{start,callback}`) but share one handler instance so the
// state-cookie / find-or-create logic stays centralized.
type OAuthHandler struct {
	cfg     *config.Config
	service *Service
	client  *http.Client
}

func NewOAuthHandler(cfg *config.Config, service *Service) *OAuthHandler {
	return &OAuthHandler{
		cfg:     cfg,
		service: service,
		// 10s ceiling for token exchange + userinfo. Providers that can't
		// answer in 10s during a foreground login are broken from the
		// user's perspective regardless.
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

// GoogleStart / GoogleCallback — PKCE-less code flow because the client
// secret lives server-side. Scopes: openid email profile.
func (h *OAuthHandler) GoogleStart(w http.ResponseWriter, r *http.Request) {
	p := h.cfg.OAuth.Google
	if !p.Enabled {
		h.redirectError(w, r, "disabled")
		return
	}
	state, err := h.beginFlow(w)
	if err != nil {
		h.redirectError(w, r, "server")
		return
	}
	u, _ := url.Parse("https://accounts.google.com/o/oauth2/v2/auth")
	q := u.Query()
	q.Set("client_id", p.ClientID)
	q.Set("redirect_uri", p.RedirectURL)
	q.Set("response_type", "code")
	q.Set("scope", "openid email profile")
	q.Set("state", state)
	q.Set("access_type", "online")
	q.Set("prompt", "select_account")
	u.RawQuery = q.Encode()
	http.Redirect(w, r, u.String(), http.StatusFound)
}

func (h *OAuthHandler) GoogleCallback(w http.ResponseWriter, r *http.Request) {
	p := h.cfg.OAuth.Google
	if !p.Enabled {
		h.redirectError(w, r, "disabled")
		return
	}
	if code, errParam := r.URL.Query().Get("code"), r.URL.Query().Get("error"); errParam != "" || code == "" {
		h.redirectError(w, r, "denied")
		return
	}
	if err := VerifyState(r, h.cfg.OAuth.StateSecret, r.URL.Query().Get("state")); err != nil {
		h.redirectError(w, r, "state")
		return
	}
	http.SetCookie(w, ClearStateCookie(h.cfg.Session.Secure, h.cfg.Session.SameSite))

	// Exchange code for tokens.
	tok, err := h.exchangeGoogle(r.Context(), p, r.URL.Query().Get("code"))
	if err != nil {
		h.redirectError(w, r, "upstream")
		return
	}
	// Fetch userinfo (simpler than parsing/verifying the ID token locally —
	// we hit googleapis over TLS right after the exchange, so MITM isn't a
	// practical risk here).
	info, err := h.fetchGoogleUser(r.Context(), tok.AccessToken)
	if err != nil {
		h.redirectError(w, r, "upstream")
		return
	}
	h.finishFlow(w, r, OAuthClaims{
		Provider:      ProviderGoogle,
		Subject:       info.Sub,
		Email:         info.Email,
		EmailVerified: info.EmailVerified,
	})
}

// GitHubStart / GitHubCallback — scopes read:user user:email. We deliberately
// skip `repo` so the consent screen isn't scary.
func (h *OAuthHandler) GitHubStart(w http.ResponseWriter, r *http.Request) {
	p := h.cfg.OAuth.GitHub
	if !p.Enabled {
		h.redirectError(w, r, "disabled")
		return
	}
	state, err := h.beginFlow(w)
	if err != nil {
		h.redirectError(w, r, "server")
		return
	}
	u, _ := url.Parse("https://github.com/login/oauth/authorize")
	q := u.Query()
	q.Set("client_id", p.ClientID)
	q.Set("redirect_uri", p.RedirectURL)
	q.Set("scope", "read:user user:email")
	q.Set("state", state)
	q.Set("allow_signup", "true")
	u.RawQuery = q.Encode()
	http.Redirect(w, r, u.String(), http.StatusFound)
}

func (h *OAuthHandler) GitHubCallback(w http.ResponseWriter, r *http.Request) {
	p := h.cfg.OAuth.GitHub
	if !p.Enabled {
		h.redirectError(w, r, "disabled")
		return
	}
	if code, errParam := r.URL.Query().Get("code"), r.URL.Query().Get("error"); errParam != "" || code == "" {
		h.redirectError(w, r, "denied")
		return
	}
	if err := VerifyState(r, h.cfg.OAuth.StateSecret, r.URL.Query().Get("state")); err != nil {
		h.redirectError(w, r, "state")
		return
	}
	http.SetCookie(w, ClearStateCookie(h.cfg.Session.Secure, h.cfg.Session.SameSite))

	accessToken, err := h.exchangeGitHub(r.Context(), p, r.URL.Query().Get("code"))
	if err != nil {
		h.redirectError(w, r, "upstream")
		return
	}
	user, err := h.fetchGitHubUser(r.Context(), accessToken)
	if err != nil {
		h.redirectError(w, r, "upstream")
		return
	}
	email, verified, err := h.fetchGitHubPrimaryEmail(r.Context(), accessToken)
	if err != nil {
		h.redirectError(w, r, "upstream")
		return
	}
	h.finishFlow(w, r, OAuthClaims{
		Provider:      ProviderGitHub,
		Subject:       strconv.FormatInt(user.ID, 10),
		Email:         email,
		EmailVerified: verified,
	})
}

// beginFlow issues the state cookie and returns the value to embed in the
// provider authorize URL.
func (h *OAuthHandler) beginFlow(w http.ResponseWriter) (string, error) {
	cookie, value, err := IssueStateCookie(h.cfg.OAuth.StateSecret, h.cfg.Session.Secure, h.cfg.Session.SameSite)
	if err != nil {
		return "", err
	}
	http.SetCookie(w, cookie)
	return value, nil
}

// finishFlow is the shared tail of both callbacks: resolve the user,
// set session cookies, redirect to the frontend.
func (h *OAuthHandler) finishFlow(w http.ResponseWriter, r *http.Request, claims OAuthClaims) {
	user, cookies, err := h.service.FindOrCreateFromOAuth(r.Context(), claims)
	if err != nil {
		h.redirectError(w, r, mapOAuthError(err))
		return
	}
	for _, c := range cookies {
		http.SetCookie(w, c)
	}
	_ = user
	// Cache-bust the redirect target so a stale Cloudflare-edge cache of
	// /dashboard (older SPA shell without the /auth/me hydration probe)
	// doesn't steal the response. Each callback gets a unique ts param —
	// the SPA ignores it but the CDN sees a fresh URL.
	target := fmt.Sprintf("%s/dashboard?oauth=%d", h.cfg.OAuth.FrontendURL, time.Now().UnixNano())
	http.Redirect(w, r, target, http.StatusFound)
}

func (h *OAuthHandler) redirectError(w http.ResponseWriter, r *http.Request, code string) {
	target := h.cfg.OAuth.FrontendURL + "/login?oauth_error=" + url.QueryEscape(code)
	http.Redirect(w, r, target, http.StatusFound)
}

func mapOAuthError(err error) string {
	switch {
	case errors.Is(err, ErrOAuthEmailUnverified):
		return "unverified"
	case errors.Is(err, ErrOAuthInviteRequired):
		return "invite_required"
	case errors.Is(err, ErrOAuthStateMismatch):
		return "state"
	case errors.Is(err, ErrOAuthProviderDisabled):
		return "disabled"
	case errors.Is(err, ErrOAuthUpstreamFailure):
		return "upstream"
	default:
		return "server"
	}
}

// --- Google helpers ---

type googleTokenResp struct {
	AccessToken string `json:"access_token"`
	IDToken     string `json:"id_token"`
	ExpiresIn   int    `json:"expires_in"`
	TokenType   string `json:"token_type"`
}

type googleUserInfo struct {
	Sub           string `json:"sub"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	Name          string `json:"name"`
	Picture       string `json:"picture"`
}

func (h *OAuthHandler) exchangeGoogle(ctx context.Context, p config.OAuthProvider, code string) (*googleTokenResp, error) {
	form := url.Values{}
	form.Set("code", code)
	form.Set("client_id", p.ClientID)
	form.Set("client_secret", p.ClientSecret)
	form.Set("redirect_uri", p.RedirectURL)
	form.Set("grant_type", "authorization_code")
	req, _ := http.NewRequestWithContext(ctx, http.MethodPost,
		"https://oauth2.googleapis.com/token", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	return decodeJSON[googleTokenResp](h.client, req)
}

func (h *OAuthHandler) fetchGoogleUser(ctx context.Context, accessToken string) (*googleUserInfo, error) {
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet,
		"https://openidconnect.googleapis.com/v1/userinfo", nil)
	req.Header.Set("Authorization", "Bearer "+accessToken)
	return decodeJSON[googleUserInfo](h.client, req)
}

// --- GitHub helpers ---

type githubTokenResp struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	Scope       string `json:"scope"`
	Error       string `json:"error"`
}

type githubUser struct {
	ID    int64  `json:"id"`
	Login string `json:"login"`
	Email string `json:"email"` // often null for private-email users
	Name  string `json:"name"`
}

type githubEmail struct {
	Email    string `json:"email"`
	Primary  bool   `json:"primary"`
	Verified bool   `json:"verified"`
}

func (h *OAuthHandler) exchangeGitHub(ctx context.Context, p config.OAuthProvider, code string) (string, error) {
	form := url.Values{}
	form.Set("client_id", p.ClientID)
	form.Set("client_secret", p.ClientSecret)
	form.Set("code", code)
	form.Set("redirect_uri", p.RedirectURL)
	req, _ := http.NewRequestWithContext(ctx, http.MethodPost,
		"https://github.com/login/oauth/access_token", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")
	out, err := decodeJSON[githubTokenResp](h.client, req)
	if err != nil {
		return "", err
	}
	if out.Error != "" || out.AccessToken == "" {
		return "", fmt.Errorf("github token: %s", out.Error)
	}
	return out.AccessToken, nil
}

func (h *OAuthHandler) fetchGitHubUser(ctx context.Context, accessToken string) (*githubUser, error) {
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.github.com/user", nil)
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/vnd.github+json")
	return decodeJSON[githubUser](h.client, req)
}

// fetchGitHubPrimaryEmail picks the primary-verified email. GitHub allows
// users to keep emails private on /user, so /user/emails is the reliable path.
func (h *OAuthHandler) fetchGitHubPrimaryEmail(ctx context.Context, accessToken string) (string, bool, error) {
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.github.com/user/emails", nil)
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/vnd.github+json")
	resp, err := h.client.Do(req)
	if err != nil {
		return "", false, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return "", false, fmt.Errorf("github emails http %d", resp.StatusCode)
	}
	var emails []githubEmail
	if err := json.NewDecoder(resp.Body).Decode(&emails); err != nil {
		return "", false, err
	}
	// Primary + verified wins. If there's no primary verified, the account
	// doesn't have a confirmed email we're willing to trust.
	for _, e := range emails {
		if e.Primary && e.Verified {
			return e.Email, true, nil
		}
	}
	// Fall back: any verified email (still safe — GitHub verified it).
	for _, e := range emails {
		if e.Verified {
			return e.Email, true, nil
		}
	}
	return "", false, nil
}

// decodeJSON runs a request and decodes into T. Non-2xx responses become errors.
func decodeJSON[T any](client *http.Client, req *http.Request) (*T, error) {
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return nil, fmt.Errorf("http %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	var out T
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	return &out, nil
}
