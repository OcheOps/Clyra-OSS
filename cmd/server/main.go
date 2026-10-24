package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/getsentry/sentry-go"
	sentryhttp "github.com/getsentry/sentry-go/http"
	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"

	"github.com/oche/diagram/internal/abuse"

	"github.com/oche/diagram/internal/account"
	"github.com/oche/diagram/internal/activation"
	"github.com/oche/diagram/internal/admin"
	"github.com/oche/diagram/internal/apitoken"
	"github.com/oche/diagram/internal/audit"
	"github.com/oche/diagram/internal/auth"
	"github.com/oche/diagram/internal/cloudsync"
	"github.com/oche/diagram/internal/config"
	"github.com/oche/diagram/internal/database"
	"github.com/oche/diagram/internal/diagram"
	"github.com/oche/diagram/internal/feedback"
	"github.com/oche/diagram/internal/iac"
	"github.com/oche/diagram/internal/iac/reverse"
	"github.com/oche/diagram/internal/icons"
	"github.com/oche/diagram/internal/invite"
	"github.com/oche/diagram/internal/mcp"
	"github.com/oche/diagram/internal/middleware"
	"github.com/oche/diagram/internal/observability"
	"github.com/oche/diagram/internal/payments"
	"github.com/oche/diagram/internal/pricing"
	"github.com/oche/diagram/internal/provider"
	"github.com/oche/diagram/internal/publicdemo"
	"github.com/oche/diagram/internal/repository"
	"github.com/oche/diagram/internal/server"
)

func main() {
	godotenv.Load()

	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	logger := setupLogger(cfg.Env)

	// Sentry — best-effort error reporting. No DSN = silent skip (safe for local dev).
	if dsn := os.Getenv("SENTRY_DSN"); dsn != "" {
		if err := sentry.Init(sentry.ClientOptions{
			Dsn:              dsn,
			Environment:      cfg.Env,
			TracesSampleRate: 0.0, // errors only — keep free-tier event budget for what matters
			AttachStacktrace: true,
		}); err != nil {
			logger.Warn("sentry init failed", "error", err)
		} else {
			defer sentry.Flush(2 * time.Second)
			// Forward every Error-level slog record to Sentry. Wrap the existing
			// handler so stdout logging keeps working unchanged.
			logger = slog.New(observability.NewSentryHandler(logger.Handler()))
			slog.SetDefault(logger)
			logger.Info("sentry enabled", "env", cfg.Env)
		}
	}

	// Database
	pool, err := database.Connect(cfg)
	if err != nil {
		logger.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer pool.Close()
	logger.Info("database connected")

	// Redis (optional)
	redisClient, err := database.ConnectRedis(cfg)
	if err != nil {
		logger.Warn("redis unavailable, falling back to in-memory rate limiting", "error", err)
	} else {
		logger.Info("redis connected")
	}

	// Repositories
	userRepo := repository.NewUserPostgres(pool)
	sessionRepo := repository.NewSessionPostgres(pool)
	inviteRepo := invite.NewRepository(pool)

	// Seed admin user (no-op if ADMIN_EMAIL/ADMIN_PASSWORD unset in dev).
	seedAdmin(context.Background(), cfg, userRepo, logger)

	// Auth (with invite validation + per-account login lockout + password reset)
	loginLockout := abuse.NewLoginLockout(5, 15*time.Minute, 15*time.Minute)
	resetRepo := repository.NewPasswordResetPostgres(pool)

	// Email service (Resend)
	emailService := invite.NewEmailService(cfg.Resend, logger)
	if emailService.Enabled() {
		logger.Info("email service enabled", "from", cfg.Resend.FromEmail)
	}

	identityRepo := repository.NewIdentityPostgres(pool)
	activationRecorder := activation.NewRecorder(pool, logger)
	auditRecorder := audit.NewRecorder(pool, logger)

	authService := auth.NewService(cfg, userRepo, sessionRepo, inviteRepo, logger).
		WithLoginLockout(loginLockout).
		WithIdentities(identityRepo).
		WithActivation(&activationAdapter{r: activationRecorder}).
		WithAudit(auditRecorder).
		WithPasswordReset(&passwordResetStoreAdapter{repo: resetRepo}, emailService).
		WithWelcomeMailer(emailService)

	// Providers
	providerRegistry := provider.NewRegistry()

	// Managed Bedrock is disabled: Clyra is open source and no longer runs
	// a hosted managed-inference path. BYO (openai/anthropic/gemini) is the
	// only route. The code is kept in internal/provider/bedrock.go for a
	// future self-hoster who wants to wire their own AWS credentials, but
	// the registry does NOT advertise it on this build.
	//
	// if cfg.Bedrock.Enabled {
	// 	bedrock, err := provider.NewBedrock(context.Background(), provider.BedrockConfig{
	// 		Region:          cfg.Bedrock.Region,
	// 		ModelID:         cfg.Bedrock.ModelID,
	// 		Profile:         cfg.Bedrock.Profile,
	// 		AccessKeyID:     cfg.Bedrock.AccessKeyID,
	// 		SecretAccessKey: cfg.Bedrock.SecretAccessKey,
	// 	})
	// 	if err != nil {
	// 		logger.Error("failed to initialize Bedrock provider", "error", err)
	// 		os.Exit(1)
	// 	}
	// 	providerRegistry.Register("managed_bedrock", bedrock)
	// 	logger.Info("bedrock provider enabled", "region", cfg.Bedrock.Region, "model", cfg.Bedrock.ModelID)
	// }

	providerRegistry.Register("openai", provider.NewOpenAI())
	providerRegistry.Register("anthropic", provider.NewAnthropic())
	providerRegistry.Register("gemini", provider.NewGemini())

	// Abuse prevention
	quotaRepo := abuse.NewQuotaRepository(pool)
	quotaService := abuse.NewQuotaService(quotaRepo, cfg.Abuse.DailyLimit, cfg.Abuse.PlanLimits)
	usageLog := abuse.NewUsageLogRepository(pool)
	abuseGuard := abuse.NewGuard(cfg.Abuse, quotaService, usageLog, logger)

	var userRateLimiter abuse.UserRateLimiter
	if redisClient != nil {
		userRateLimiter = abuse.NewRedisUserRateLimiter(redisClient, cfg.Abuse)
	} else {
		userRateLimiter = abuse.NewMemoryUserRateLimiter(cfg.Abuse)
	}

	// Services
	diagramService := diagram.NewService(providerRegistry, abuseGuard, logger).
		WithActivation(&activationAdapter{r: activationRecorder})
	iacService := iac.NewService(providerRegistry, abuseGuard, logger).
		WithActivation(&activationAdapter{r: activationRecorder}).
		WithAudit(auditRecorder)

	// IP rate limiter
	var ipLimiter middleware.RateLimiter
	if redisClient != nil {
		ipLimiter = middleware.NewRedisRateLimiter(redisClient, cfg.RateLimit)
	} else {
		ipLimiter = middleware.NewMemoryRateLimiter(cfg.RateLimit)
	}

	// Tight auth-surface limiter — runs in front of login/register/forgot/reset
	// so credential stuffing and email enumeration can't ride the looser global
	// limit. Uses Redis when available so the cap holds across replicas.
	authLimiter := middleware.NewAuthRateLimit(redisClient, "auth", cfg.RateLimit.AuthPerMinute)
	// Public landing demo throttle — one call per IP per minute is the
	// spec'd budget (backlogfast.md #1). Reuses AuthRateLimit's per-IP
	// sliding window; namespace stays separate via a different limiter
	// instance so credential-stuffing and demo-hammering don't share a
	// budget.
	demoLimiter := middleware.NewAuthRateLimit(redisClient, "demo", 1)

	// Payments — Paystack and Polar wiring kept in-tree for reference, but
	// no provider is constructed in the open-source build. Clyra is no
	// longer sold as a hosted product. A self-hoster who wants a paid tier
	// can uncomment the block below and bring their own Paystack secret
	// key or Polar access token via env.
	//
	// var paystack *payments.Paystack
	// if cfg.Paystack.SecretKey != "" || cfg.Env == "development" {
	// 	paystack = payments.NewPaystack(cfg.Paystack.SecretKey, cfg.Paystack.CallbackURL)
	// }
	// var polar *payments.Polar
	// if cfg.Polar.AccessToken != "" || cfg.Env == "development" {
	// 	polar = payments.NewPolar(
	// 		cfg.Polar.AccessToken,
	// 		cfg.Polar.BaseURL,
	// 		cfg.Polar.CallbackURL,
	// 		cfg.Polar.WebhookSecret,
	// 		map[string]string{
	// 			"pro": cfg.Polar.ProductIDPro,
	// 			"byo": cfg.Polar.ProductIDBYO,
	// 		},
	// 	)
	// }
	var paystack *payments.Paystack
	var polar *payments.Polar
	logger.Info("payments: disabled (open-source build)")

	paymentRepo := payments.NewRepository(pool)
	paymentService := payments.NewService(paymentRepo, userRepo, logger, paystack, polar).
		WithAudit(auditRecorder)
	idemCache := payments.NewIdempotencyCache(redisClient)

	// Sweep abandoned pending rows every 10m (threshold 2h) so they don't
	// wedge users out of fresh checkouts. See internal/payments/cleanup.go.
	payments.StartStaleCleaner(context.Background(), paymentRepo, logger, 2*time.Hour, 10*time.Minute)

	// Cloud sync storage (S3 when S3_BUCKET is set, else DB fallback)
	storage, err := cloudsync.NewStorage(cfg.Storage)
	if err != nil {
		logger.Error("failed to initialize cloud storage", "error", err)
		os.Exit(1)
	}
	logger.Info("cloud sync storage", "kind", storage.Kind())
	cloudRepo := cloudsync.NewRepository(pool)
	cloudService := cloudsync.NewService(cloudRepo, storage, func(userID string) (string, bool) {
		u, err := userRepo.GetByID(context.Background(), userID)
		if err != nil || u == nil {
			return "free", false
		}
		return string(u.EffectivePlan()), true
	}, logger).WithAudit(auditRecorder)

	// Pricing / geo detector
	pricingDetector := pricing.NewDetector(cfg.Pricing.DefaultRegion)
	pricingHandler := pricing.NewHandler(pricingDetector)
	if cfg.Pricing.DefaultRegion != "" {
		logger.Info("pricing: default region override", "region", cfg.Pricing.DefaultRegion)
	}

	// Handlers
	handlers := server.Handlers{
		Auth:       auth.NewHandler(authService),
		OAuth:      auth.NewOAuthHandler(cfg, authService),
		Diagram:    diagram.NewHandler(diagramService),
		Invite:     invite.NewHandler(inviteRepo, emailService),
		IaC:        iac.NewHandler(iacService),
		IaCReverse: reverse.NewHandler(abuseGuard),
		Account:    account.NewHandler(quotaService),
		Pricing:    pricingHandler,
		Payments: payments.NewHandler(paymentService, paystack, polar).
			WithIdempotency(idemCache).
			WithAudit(auditRecorder),
		CloudSync:  cloudsync.NewHandler(cloudService),
		Admin:      admin.NewHandler(userRepo, emailService, pool, logger).WithAudit(auditRecorder),
		Feedback:   feedback.NewHandler(feedback.NewService(pool, &feedbackMailerAdapter{e: emailService}, cfg.Admin.Email, logger)),
		PublicDemo: publicdemo.New(diagramService, logger, cfg.CORS.AllowedOrigins),
	}
	// MCP + PAT wiring. Kept separate from the struct literal above so
	// the shared apiTokenRepo is reachable by both the settings handler
	// and the /mcp bearer middleware without spinning up two repos.
	apiTokenRepo := apitoken.NewRepository(pool)
	handlers.APIToken = apitoken.NewHandler(apitoken.NewService(apiTokenRepo), logger)
	handlers.Icons = icons.NewHandler()

	mcpHandler := mcp.NewHandler(mcp.HandlerOpts{
		PlanGate: mcp.DefaultPlanGate(),
		Logger:   logger,
		AppURL:   cfg.Resend.AppURL,
		Audit:    auditRecorder,
		Limiter:  mcp.NewRedisDailyLimiter(redisClient, mcp.DefaultDailyLimit),
	})
	// Wire the phase-2 tools. Both call the same domain services the SPA
	// uses, so plan gates, abuse checks, normalization, and audit events
	// fire consistently across in-app and MCP paths.
	mcpDeps := mcp.ToolDeps{
		Diagram:          diagramService,
		Cloud:            cloudService,
		AppURL:           cfg.Resend.AppURL,
		AllowedProviders: []string{"managed_bedrock"},
	}
	// generate_diagram is disabled in the open-source build: it depends on
	// managed Bedrock and MCP callers can't safely pass their own API key
	// through the tool contract yet. reverse_from_iac is a local HCL parser
	// so it stays live.
	// mcpHandler.Register(mcp.NewGenerateDiagramTool(mcpDeps))
	mcpHandler.Register(mcp.NewReverseFromIaCTool(mcpDeps))
	handlers.MCP = mcpHandler
	handlers.APITokenAuth = middleware.AuthenticateAPIToken(apiTokenRepo, userRepo)

	// Middleware
	mw := middleware.New(cfg, logger, authService, ipLimiter)
	abuseMW := abuse.NewMiddleware(userRateLimiter, quotaService, abuseGuard.PromptValidator, logger)

	// Router
	router := server.NewRouter(cfg, handlers, mw, abuseMW, authLimiter, demoLimiter)

	// Sentry HTTP middleware — recovers panics, attaches request scope. Wraps
	// the whole router so every endpoint is covered.
	var handler http.Handler = router
	if os.Getenv("SENTRY_DSN") != "" {
		handler = sentryhttp.New(sentryhttp.Options{Repanic: true}).Handle(router)
	}

	// Start
	srv := server.New(cfg, handler, logger)
	if err := srv.Start(); err != nil {
		logger.Error("server failed", "error", err)
		os.Exit(1)
	}
}

// seedAdmin creates the admin user if missing. Credentials come from
// ADMIN_EMAIL / ADMIN_PASSWORD env. In dev we fall back to a documented
// default so the first-boot experience stays simple; in production the
// config validator has already rejected an empty value.
func seedAdmin(ctx context.Context, cfg *config.Config, users *repository.UserPostgres, logger *slog.Logger) {
	email := cfg.Admin.Email
	password := cfg.Admin.Password

	if email == "" || password == "" {
		if cfg.Env == "production" {
			// Shouldn't reach here — validate() blocks prod earlier. Belt-and-braces:
			logger.Error("admin seed skipped in production: ADMIN_EMAIL / ADMIN_PASSWORD unset")
			return
		}
		// Dev fallback — loud warning so operators know what's happening.
		email = "oche@ocheverse.ng"
		password = "admin12345678"
		logger.Warn("using development admin fallback — set ADMIN_EMAIL / ADMIN_PASSWORD in .env", "email", email)
	}

	existing, _ := users.GetByEmail(ctx, email)
	if existing != nil {
		// Keep the admin on the pro plan even if the row predates the plan column.
		if existing.Plan != auth.PlanPro {
			if err := users.UpdatePlan(ctx, existing.ID, auth.PlanPro); err != nil {
				logger.Warn("failed to promote admin plan", "error", err)
			}
		}

		// Password rotation: if ADMIN_PASSWORD is set in env AND it doesn't verify
		// against the stored hash, re-hash and update. This is idempotent —
		// providing the *current* password in env is a no-op.
		//
		// Scoped intentionally: only when cfg.Admin.Password is explicitly set
		// (not when we fell through to the dev default). Keeps dev restarts from
		// silently rotating a password the operator forgot about.
		if cfg.Admin.Password != "" {
			if err := bcrypt.CompareHashAndPassword([]byte(existing.PasswordHash), []byte(cfg.Admin.Password)); err != nil {
				newHash, hashErr := bcrypt.GenerateFromPassword([]byte(cfg.Admin.Password), 12)
				if hashErr != nil {
					logger.Error("admin password rotation: hash failed", "error", hashErr)
					return
				}
				if updErr := users.UpdatePasswordHash(ctx, existing.ID, string(newHash)); updErr != nil {
					logger.Error("admin password rotation: update failed", "error", updErr)
					return
				}
				logger.Warn("admin password rotated from ADMIN_PASSWORD env", "email", email)
			}
		}
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		logger.Error("failed to hash admin password", "error", err)
		return
	}

	admin := &auth.User{
		ID:           "admin-seed-001",
		Email:        email,
		PasswordHash: string(hash),
		Role:         auth.RoleAdmin,
		Plan:         auth.PlanPro,
	}

	if err := users.Create(ctx, admin); err != nil {
		logger.Error("failed to seed admin", "error", err)
		return
	}

	logger.Info("admin user seeded", "email", email)
}

// feedbackMailerAdapter bridges *invite.EmailService to the feedback
// package's EmailNotifier interface — same pattern we use elsewhere to
// keep the email transport in one place without leaking its concrete type.
type feedbackMailerAdapter struct {
	e *invite.EmailService
}

func (a *feedbackMailerAdapter) Enabled() bool {
	return a != nil && a.e != nil && a.e.Enabled()
}

func (a *feedbackMailerAdapter) NotifyFeedback(adminEmail string, item feedback.Item) error {
	if a == nil || a.e == nil {
		return nil
	}
	return a.e.SendFeedbackNotification(adminEmail, item.Name, item.Email, item.Category, item.Message, item.ID)
}

// activationAdapter wraps *activation.Recorder so it can satisfy the
// auth / diagram / iac packages' ActivationRecorder interfaces (each of
// which uses a `kind string` signature rather than the recorder's typed
// `activation.Kind`, to avoid cross-package coupling).
type activationAdapter struct {
	r *activation.Recorder
}

func (a *activationAdapter) Record(ctx context.Context, userID, kind string) {
	if a == nil || a.r == nil {
		return
	}
	a.r.Record(ctx, userID, activation.Kind(kind))
}

func (a *activationAdapter) HasKind(ctx context.Context, userID, kind string) bool {
	if a == nil || a.r == nil {
		return false
	}
	return a.r.HasKind(ctx, userID, activation.Kind(kind))
}

// passwordResetStoreAdapter bridges repository.PasswordResetPostgres (which
// uses its own record type) to auth.PasswordResetStore (which uses an
// auth-package type, to avoid an import cycle).
type passwordResetStoreAdapter struct {
	repo *repository.PasswordResetPostgres
}

func (a *passwordResetStoreAdapter) Create(ctx context.Context, rec *auth.PasswordResetRecord) error {
	return a.repo.Create(ctx, &repository.PasswordResetRecord{
		ID:        rec.ID,
		UserID:    rec.UserID,
		TokenHash: rec.TokenHash,
		ExpiresAt: rec.ExpiresAt,
		UsedAt:    rec.UsedAt,
		CreatedAt: rec.CreatedAt,
	})
}

func (a *passwordResetStoreAdapter) GetByTokenHash(ctx context.Context, hash string) (*auth.PasswordResetRecord, error) {
	r, err := a.repo.GetByTokenHash(ctx, hash)
	if err != nil || r == nil {
		return nil, err
	}
	return &auth.PasswordResetRecord{
		ID: r.ID, UserID: r.UserID, TokenHash: r.TokenHash,
		ExpiresAt: r.ExpiresAt, UsedAt: r.UsedAt, CreatedAt: r.CreatedAt,
	}, nil
}

func (a *passwordResetStoreAdapter) MarkUsed(ctx context.Context, id string) error {
	return a.repo.MarkUsed(ctx, id)
}

func (a *passwordResetStoreAdapter) RecentCountForUser(ctx context.Context, userID string, within time.Duration) (int, error) {
	return a.repo.RecentCountForUser(ctx, userID, within)
}

func setupLogger(env string) *slog.Logger {
	var handler slog.Handler
	if env == "production" {
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})
	} else {
		handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})
	}
	logger := slog.New(handler)
	slog.SetDefault(logger)
	return logger
}
