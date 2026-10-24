package config

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Server    ServerConfig
	Database  DatabaseConfig
	Redis     RedisConfig
	Session   SessionConfig
	RateLimit RateLimitConfig
	CORS      CORSConfig
	Bedrock   BedrockConfig
	Abuse     AbuseConfig
	Resend    ResendConfig
	Paystack  PaystackConfig
	Polar     PolarConfig
	OAuth     OAuthConfig
	Storage   StorageConfig
	Pricing   PricingConfig
	Admin     AdminConfig
	Registration RegistrationConfig
	Env       string // "development" or "production"
}

// OAuthConfig bundles all third-party sign-in providers. Each provider has
// an Enabled flag so they can be rolled out independently without code
// changes. The StateSecret signs the short-lived state cookie used to
// defend the OAuth redirect from CSRF (shared across providers since the
// cookie is scoped per flow anyway).
type OAuthConfig struct {
	StateSecret string
	FrontendURL string // absolute base used for post-callback redirects (e.g. "/dashboard")
	Google      OAuthProvider
	GitHub      OAuthProvider
}

type OAuthProvider struct {
	Enabled      bool
	ClientID     string
	ClientSecret string
	RedirectURL  string
}

// RegistrationConfig controls whether signup requires an invite code or is
// open to the public. Values: "invite" (default — must present a valid code)
// or "open" (anyone with a valid email may register). Set via
// REGISTRATION_MODE env var.
type RegistrationConfig struct {
	Mode string
}

const (
	RegistrationInvite = "invite"
	RegistrationOpen   = "open"
)

// AdminConfig carries the bootstrap admin credentials. In production both
// fields are required — the server refuses to start without them — so nobody
// accidentally ships with the hardcoded dev defaults.
type AdminConfig struct {
	Email    string
	Password string
}

// PaystackConfig — Nigerian payment rails. NGN + accepts international cards.
// When SecretKey is empty the payment service falls back to mock mode:
// it mints an auth URL that auto-succeeds in the verify step.
type PaystackConfig struct {
	SecretKey string
	PublicKey string
	CallbackURL string // optional frontend redirect after checkout
}

// PolarConfig — merchant-of-record provider for non-NG customers.
// Polar handles VAT/sales tax globally, so the tenant doesn't need a
// US / EU entity. Leave AccessToken empty to run in mock mode locally.
//
// BaseURL chooses production vs sandbox — use PolarBaseSandbox ("https://
// sandbox-api.polar.sh") for test-card payments, live URL for real money.
//
// Per-plan product IDs are required because Polar's checkout model is
// product-centric, not amount-centric (the price lives on the product).
type PolarConfig struct {
	AccessToken   string // organization access token ("polar_oat_...")
	BaseURL       string // "" → production; PolarBaseSandbox to force sandbox
	CallbackURL   string // where Polar redirects after successful checkout
	WebhookSecret string // whsec_... — from Polar webhook endpoint
	ProductIDPro  string // Polar product UUID for the Pro plan
	ProductIDBYO  string // Polar product UUID for the BYO plan
}

// PricingConfig lets admins force a region when detection is inconclusive
// (e.g. running on localhost in Nigeria). Leave empty for "global / USD".
type PricingConfig struct {
	DefaultRegion string
}

// StorageConfig governs the optional S3-compatible project storage.
// Leaving Endpoint + Bucket empty falls back to in-database storage.
type StorageConfig struct {
	Endpoint        string // e.g. https://garage.example.com (empty for AWS S3)
	Region          string // arbitrary for Garage/MinIO; real AWS region for S3
	Bucket          string
	AccessKeyID     string
	SecretAccessKey string
	ForcePathStyle  bool // required for MinIO / Garage
	PublicBaseURL   string // optional CDN-facing base URL for previews; empty → signed S3 URLs
}

type ResendConfig struct {
	APIKey    string
	FromEmail string
	FromName  string
	AppURL    string // e.g. https://useclyra.com
}

// AbuseConfig holds all abuse prevention and cost control settings.
type AbuseConfig struct {
	// Per-user rate limiting (generation endpoints only)
	UserRatePerMinute int // Default: 5

	// Daily quota fallback used when a user has no known plan.
	DailyLimit int // Default: 30

	// Per-plan daily limits. Keys are plan names (free, pro, byo).
	// Unknown plans fall back to DailyLimit.
	PlanLimits map[string]int

	// Prompt guardrails
	MaxPromptLength int           // Default: 2000 chars
	MaxTokens       int           // Default: 1000
	ProviderTimeout time.Duration // Default: 120s (free baseline; paid plans get 180s via abuse.PaidPlanProviderTimeout)

	// Diagram size limits
	MaxNodes int // Default: 25

	// Circuit breaker
	CircuitBreakerThreshold int           // Failures before tripping. Default: 5
	CircuitBreakerCooldown  time.Duration // How long to stay open. Default: 30s
}

type BedrockConfig struct {
	Enabled         bool
	Region          string
	ModelID         string
	Profile         string // AWS named profile (optional — falls back to default chain)
	AccessKeyID     string // Optional — falls back to AWS credential chain if empty
	SecretAccessKey string
}


type ServerConfig struct {
	Host         string
	Port         int
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
	MaxBodyBytes int64
}

type DatabaseConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	Name     string
	SSLMode  string
	MaxConns int32
}

func (d DatabaseConfig) DSN() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		d.User, d.Password, d.Host, d.Port, d.Name, d.SSLMode,
	)
}

type RedisConfig struct {
	Host     string
	Port     int
	Password string
	DB       int
}

type SessionConfig struct {
	Secret   string
	CookieName string
	MaxAge     int  // seconds
	Secure     bool
	HttpOnly   bool
	SameSite   string // "lax", "strict", "none"
}

type RateLimitConfig struct {
	RequestsPerMinute int
	Burst             int
	// AuthPerMinute caps unauthenticated auth-surface attempts per IP per
	// minute (login, register, forgot-password, reset-password). Tighter
	// than the global RequestsPerMinute because the auth surface is the
	// preferred target for credential stuffing + email enumeration.
	AuthPerMinute int
}

type CORSConfig struct {
	AllowedOrigins []string
	MaxAge         int // seconds
}

// Load reads configuration from environment variables with sensible defaults.
func Load() (*Config, error) {
	cfg := &Config{
		Env: envOrDefault("APP_ENV", "development"),
		Server: ServerConfig{
			Host:         envOrDefault("SERVER_HOST", "0.0.0.0"),
			Port:         envIntOrDefault("SERVER_PORT", 8080),
			ReadTimeout:  envDurationOrDefault("SERVER_READ_TIMEOUT", 10*time.Second),
			WriteTimeout: envDurationOrDefault("SERVER_WRITE_TIMEOUT", 120*time.Second),
			IdleTimeout:  envDurationOrDefault("SERVER_IDLE_TIMEOUT", 60*time.Second),
			MaxBodyBytes: int64(envIntOrDefault("SERVER_MAX_BODY_BYTES", 1<<20)), // 1MB
		},
		Database: DatabaseConfig{
			Host:     envOrDefault("DB_HOST", "localhost"),
			Port:     envIntOrDefault("DB_PORT", 5432),
			User:     envOrDefault("DB_USER", "diagram"),
			Password: envOrDefault("DB_PASSWORD", ""),
			Name:     envOrDefault("DB_NAME", "diagram"),
			SSLMode:  envOrDefault("DB_SSLMODE", "disable"),
			MaxConns: int32(envIntOrDefault("DB_MAX_CONNS", 25)),
		},
		Redis: RedisConfig{
			Host:     envOrDefault("REDIS_HOST", "localhost"),
			Port:     envIntOrDefault("REDIS_PORT", 6379),
			Password: envOrDefault("REDIS_PASSWORD", ""),
			DB:       envIntOrDefault("REDIS_DB", 0),
		},
		Session: SessionConfig{
			Secret:     os.Getenv("SESSION_SECRET"),
			CookieName: envOrDefault("SESSION_COOKIE_NAME", "diagram_session"),
			MaxAge:     envIntOrDefault("SESSION_MAX_AGE", 86400), // 24 hours
			Secure:     envBoolOrDefault("SESSION_SECURE", false),
			HttpOnly:   envBoolOrDefault("SESSION_HTTP_ONLY", true),
			SameSite:   envOrDefault("SESSION_SAME_SITE", "lax"),
		},
		RateLimit: RateLimitConfig{
			RequestsPerMinute: envIntOrDefault("RATE_LIMIT_RPM", 60),
			Burst:             envIntOrDefault("RATE_LIMIT_BURST", 10),
			AuthPerMinute:     envIntOrDefault("RATE_LIMIT_AUTH_RPM", 30),
		},
		CORS: CORSConfig{
			AllowedOrigins: envSliceOrDefault("CORS_ALLOWED_ORIGINS", []string{"http://localhost:3000"}),
			MaxAge:         envIntOrDefault("CORS_MAX_AGE", 300),
		},
		Abuse: AbuseConfig{
			UserRatePerMinute: envIntOrDefault("ABUSE_USER_RATE_PER_MINUTE", 5),
			DailyLimit:        envIntOrDefault("ABUSE_DAILY_LIMIT", 30),
			PlanLimits: map[string]int{
				"free": envIntOrDefault("PLAN_FREE_DAILY", 10),
				"pro":  envIntOrDefault("PLAN_PRO_DAILY", 200),
				"byo":  envIntOrDefault("PLAN_BYO_DAILY", 1000),
			},
			MaxPromptLength:         envIntOrDefault("ABUSE_MAX_PROMPT_LENGTH", 2000),
			MaxTokens:               envIntOrDefault("ABUSE_MAX_TOKENS", 4096),
			ProviderTimeout:         envDurationOrDefault("ABUSE_PROVIDER_TIMEOUT", 120*time.Second),
			MaxNodes:                envIntOrDefault("ABUSE_MAX_NODES", 25),
			CircuitBreakerThreshold: envIntOrDefault("ABUSE_CB_THRESHOLD", 5),
			CircuitBreakerCooldown:  envDurationOrDefault("ABUSE_CB_COOLDOWN", 30*time.Second),
		},
		Bedrock: BedrockConfig{
			Enabled:         envBoolOrDefault("BEDROCK_ENABLED", false),
			Region:          envOrDefault("BEDROCK_REGION", "us-east-1"),
			// Default targets Opus 4.8 via the global cross-region inference
			// profile — works from any Bedrock region. Override with a
			// region-scoped profile (e.g. eu.anthropic.claude-opus-4-8)
			// for lower latency when your region is known. See
			// bedrock.go for the confirmed inference-profile menu.
			ModelID:         envOrDefault("BEDROCK_MODEL_ID", "global.anthropic.claude-opus-4-8"),
			Profile:         os.Getenv("BEDROCK_PROFILE"),
			AccessKeyID:     os.Getenv("BEDROCK_ACCESS_KEY_ID"),
			SecretAccessKey: os.Getenv("BEDROCK_SECRET_ACCESS_KEY"),
		},
		Resend: ResendConfig{
			APIKey:    os.Getenv("RESEND_API_KEY"),
			FromEmail: envOrDefault("RESEND_FROM_EMAIL", "admin@example.com"),
			FromName:  envOrDefault("RESEND_FROM_NAME", "Oche from Clyra"),
			AppURL:    envOrDefault("RESEND_APP_URL", "http://localhost:3000"),
		},
		Paystack: PaystackConfig{
			SecretKey:   os.Getenv("PAYSTACK_SECRET_KEY"),
			PublicKey:   os.Getenv("PAYSTACK_PUBLIC_KEY"),
			CallbackURL: envOrDefault("PAYSTACK_CALLBACK_URL", ""),
		},
		Polar: PolarConfig{
			AccessToken:   os.Getenv("POLAR_ACCESS_TOKEN"),
			BaseURL:       os.Getenv("POLAR_BASE_URL"),
			CallbackURL:   envOrDefault("POLAR_CALLBACK_URL", ""),
			WebhookSecret: os.Getenv("POLAR_WEBHOOK_SECRET"),
			ProductIDPro:  os.Getenv("POLAR_PRODUCT_ID_PRO"),
			ProductIDBYO:  os.Getenv("POLAR_PRODUCT_ID_BYO"),
		},
		Pricing: PricingConfig{
			DefaultRegion: os.Getenv("PRICING_DEFAULT_REGION"),
		},
		Admin: AdminConfig{
			Email:    os.Getenv("ADMIN_EMAIL"),
			Password: os.Getenv("ADMIN_PASSWORD"),
		},
		Registration: RegistrationConfig{
			Mode: normalizeRegistrationMode(envOrDefault("REGISTRATION_MODE", RegistrationInvite)),
		},
		OAuth: OAuthConfig{
			StateSecret: os.Getenv("OAUTH_STATE_SECRET"),
			FrontendURL: envOrDefault("OAUTH_FRONTEND_URL", "http://localhost:3000"),
			Google: OAuthProvider{
				Enabled:      envBoolOrDefault("GOOGLE_OAUTH_ENABLED", false),
				ClientID:     os.Getenv("GOOGLE_OAUTH_CLIENT_ID"),
				ClientSecret: os.Getenv("GOOGLE_OAUTH_CLIENT_SECRET"),
				RedirectURL:  os.Getenv("GOOGLE_OAUTH_REDIRECT_URL"),
			},
			GitHub: OAuthProvider{
				Enabled:      envBoolOrDefault("GITHUB_OAUTH_ENABLED", false),
				ClientID:     os.Getenv("GITHUB_OAUTH_CLIENT_ID"),
				ClientSecret: os.Getenv("GITHUB_OAUTH_CLIENT_SECRET"),
				RedirectURL:  os.Getenv("GITHUB_OAUTH_REDIRECT_URL"),
			},
		},
		Storage: StorageConfig{
			Endpoint:        os.Getenv("S3_ENDPOINT"),
			Region:          envOrDefault("S3_REGION", "us-east-1"),
			Bucket:          os.Getenv("S3_BUCKET"),
			AccessKeyID:     os.Getenv("S3_ACCESS_KEY_ID"),
			SecretAccessKey: os.Getenv("S3_SECRET_ACCESS_KEY"),
			ForcePathStyle:  envBoolOrDefault("S3_FORCE_PATH_STYLE", true),
			PublicBaseURL:   os.Getenv("S3_PUBLIC_BASE_URL"),
		},
	}

	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("config validation: %w", err)
	}

	return cfg, nil
}

func (c *Config) validate() error {
	if c.Session.Secret == "" {
		return fmt.Errorf("SESSION_SECRET is required")
	}
	if len(c.Session.Secret) < 32 && c.Env == "production" {
		return fmt.Errorf("SESSION_SECRET must be at least 32 characters in production")
	}
	if c.Database.Password == "" && c.Env == "production" {
		return fmt.Errorf("DB_PASSWORD is required in production")
	}
	if c.Server.Port < 1 || c.Server.Port > 65535 {
		return fmt.Errorf("SERVER_PORT must be between 1 and 65535")
	}
	if c.Database.MaxConns < 1 {
		return fmt.Errorf("DB_MAX_CONNS must be at least 1")
	}
	if c.RateLimit.RequestsPerMinute < 1 {
		return fmt.Errorf("RATE_LIMIT_RPM must be at least 1")
	}
	if c.Env != "development" && c.Env != "production" {
		return fmt.Errorf("APP_ENV must be 'development' or 'production'")
	}
	if c.Env == "production" && c.Database.SSLMode == "disable" && !isInternalDBHost(c.Database.Host) {
		// SSL is required when the DB lives off-host — the connection crosses
		// untrusted networks. When it lives on the same docker network (compose
		// service name, loopback, RFC1918 address) the connection never leaves
		// the host, so plaintext is fine and SSL would just need extra cert
		// plumbing.
		return fmt.Errorf("DB_SSLMODE must not be 'disable' in production for external hosts")
	}
	if c.Env == "production" && !c.Session.Secure {
		return fmt.Errorf("SESSION_SECURE must be true in production")
	}
	if err := validateOAuthProvider("GOOGLE", c.OAuth.Google, c.OAuth.StateSecret); err != nil {
		return err
	}
	if err := validateOAuthProvider("GITHUB", c.OAuth.GitHub, c.OAuth.StateSecret); err != nil {
		return err
	}
	if c.Env == "production" {
		if c.Admin.Email == "" || c.Admin.Password == "" {
			return fmt.Errorf("ADMIN_EMAIL and ADMIN_PASSWORD are required in production (no default admin seed allowed)")
		}
		if len(c.Admin.Password) < 12 {
			return fmt.Errorf("ADMIN_PASSWORD must be at least 12 characters in production")
		}
		for _, o := range c.CORS.AllowedOrigins {
			if o == "*" {
				return fmt.Errorf(`CORS_ALLOWED_ORIGINS must not contain "*" in production (credentials are on)`)
			}
		}
		// Webhook secrets must be set in production. An empty secret puts the
		// provider in mock mode (every signature accepted) which converts an
		// env-var typo into a plan-upgrade fraud surface.
		if c.Paystack.SecretKey == "" {
			return fmt.Errorf("PAYSTACK_SECRET_KEY is required in production (mock mode accepts all webhooks)")
		}
		if c.Polar.AccessToken != "" && c.Polar.WebhookSecret == "" {
			return fmt.Errorf("POLAR_WEBHOOK_SECRET is required when POLAR_ACCESS_TOKEN is set in production")
		}
	}
	return nil
}

// validateOAuthProvider ensures that enabling a provider also supplies the
// triple of client id, secret, and redirect URL plus a shared state secret.
// Misconfiguration at boot beats a confusing 500 on the first sign-in.
// isInternalDBHost reports whether the DB host points at something on the
// same machine / private network — i.e. the connection never crosses an
// untrusted hop. We treat:
//   - localhost / loopback (127.0.0.0/8, ::1)
//   - any private RFC1918 address (10/8, 172.16/12, 192.168/16)
//   - link-local IPv6 (fc00::/7)
//   - bare hostnames with no dot (docker compose service names like "postgres")
//
// as internal. Anything else (a DNS name with a dot, a public IP) is treated
// as external and SSL is required.
func isInternalDBHost(host string) bool {
	if host == "" {
		return false
	}
	h := strings.ToLower(strings.TrimSpace(host))
	if h == "localhost" {
		return true
	}
	if ip := net.ParseIP(h); ip != nil {
		return ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast()
	}
	// No dot → not a public DNS name; almost certainly a docker compose
	// service alias resolving to a private docker bridge IP.
	return !strings.Contains(h, ".")
}

func validateOAuthProvider(prefix string, p OAuthProvider, stateSecret string) error {
	if !p.Enabled {
		return nil
	}
	if p.ClientID == "" {
		return fmt.Errorf("%s_OAUTH_CLIENT_ID is required when %s_OAUTH_ENABLED=true", prefix, prefix)
	}
	if p.ClientSecret == "" {
		return fmt.Errorf("%s_OAUTH_CLIENT_SECRET is required when %s_OAUTH_ENABLED=true", prefix, prefix)
	}
	if p.RedirectURL == "" {
		return fmt.Errorf("%s_OAUTH_REDIRECT_URL is required when %s_OAUTH_ENABLED=true", prefix, prefix)
	}
	if stateSecret == "" {
		return fmt.Errorf("OAUTH_STATE_SECRET is required when any OAuth provider is enabled")
	}
	return nil
}

// normalizeRegistrationMode keeps REGISTRATION_MODE tolerant to casing/typos.
// Anything other than the documented "open" falls back to "invite" — fail
// closed, so a misspelled env var never silently opens registration.
func normalizeRegistrationMode(v string) string {
	if strings.EqualFold(strings.TrimSpace(v), RegistrationOpen) {
		return RegistrationOpen
	}
	return RegistrationInvite
}

func envOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func envIntOrDefault(key string, fallback int) int {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	i, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return i
}

func envBoolOrDefault(key string, fallback bool) bool {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		return fallback
	}
	return b
}

func envSliceOrDefault(key string, fallback []string) []string {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	parts := strings.Split(v, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		if trimmed := strings.TrimSpace(p); trimmed != "" {
			result = append(result, trimmed)
		}
	}
	if len(result) == 0 {
		return fallback
	}
	return result
}

func envDurationOrDefault(key string, fallback time.Duration) time.Duration {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		return fallback
	}
	return d
}
