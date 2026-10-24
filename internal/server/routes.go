package server

import (
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"

	"github.com/oche/diagram/internal/abuse"
	"github.com/oche/diagram/internal/account"
	"github.com/oche/diagram/internal/admin"
	"github.com/oche/diagram/internal/apitoken"
	"github.com/oche/diagram/internal/auth"
	"github.com/oche/diagram/internal/cloudsync"
	"github.com/oche/diagram/internal/config"
	"github.com/oche/diagram/internal/diagram"
	"github.com/oche/diagram/internal/feedback"
	"github.com/oche/diagram/internal/iac"
	"github.com/oche/diagram/internal/iac/reverse"
	"github.com/oche/diagram/internal/icons"
	"github.com/oche/diagram/internal/invite"
	"github.com/oche/diagram/internal/mcp"
	"github.com/oche/diagram/internal/middleware"
	"github.com/oche/diagram/internal/payments"
	"github.com/oche/diagram/internal/pricing"
	"github.com/oche/diagram/internal/publicdemo"
	"github.com/oche/diagram/pkg/response"
)

type Handlers struct {
	Auth       *auth.Handler
	OAuth      *auth.OAuthHandler
	Diagram    *diagram.Handler
	Invite     *invite.Handler
	IaC        *iac.Handler
	IaCReverse *reverse.Handler
	Account    *account.Handler
	Pricing    *pricing.Handler
	Payments   *payments.Handler
	CloudSync  *cloudsync.Handler
	Admin      *admin.Handler
	Feedback   *feedback.Handler
	APIToken   *apitoken.Handler
	MCP        *mcp.Handler
	// APITokenAuth is the Bearer middleware for the /mcp endpoint. Set by
	// main.go alongside MCP so the router doesn't have to construct it.
	APITokenAuth func(http.Handler) http.Handler
	// Icons proxies vendor architecture-icon SVGs through our origin so
	// canvas exports (PNG / SVG / PDF) can serialize them without hitting
	// the browser's tainted-canvas guard.
	Icons *icons.Handler
	// PublicDemo serves the unauth landing-page generator. Kept here so
	// the route wiring stays in NewRouter and main.go only has to pass
	// the constructed handler through.
	PublicDemo *publicdemo.Handler
}

func NewRouter(cfg *config.Config, h Handlers, mw *middleware.Middleware, abuseMW *abuse.Middleware, authLimit *middleware.AuthRateLimit, demoLimit *middleware.AuthRateLimit) http.Handler {
	r := chi.NewRouter()

	// Global middleware
	r.Use(chiMiddleware.RequestID)
	r.Use(chiMiddleware.RealIP)
	r.Use(mw.SecurityHeaders)
	r.Use(mw.CORS)
	r.Use(mw.Logger)
	r.Use(mw.Recovery)
	r.Use(mw.RateLimit)
	r.Use(routeBodySize(cfg.Server.MaxBodyBytes))

	// Health check
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		response.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	// MCP endpoint. Lives at the top level (not under /api/v1) because MCP
	// clients like Claude Desktop expect a stable /mcp URL. Auth is
	// Bearer-only (no cookies), so it sits outside the /api/v1 CSRF group.
	// tools/list is public per spec; ServeHTTP dispatches auth check per
	// method — the middleware runs unconditionally so tools/call always
	// has a SessionUser context when it needs one.
	if h.MCP != nil {
		r.Get("/.well-known/mcp", h.MCP.WellKnown)
		r.Get("/mcp/tools", h.MCP.ListToolsHTTP)
		if h.APITokenAuth != nil {
			// The token middleware writes 401 on missing bearer, but MCP
			// clients hit initialize + tools/list before they have any
			// auth context. Route both public + auth'd through the same
			// handler; the handler chooses which methods require auth.
			r.Handle("/mcp", http.HandlerFunc(chainMiddleware(h.MCP.ServeHTTP, optionalBearer(h.APITokenAuth))))
		} else {
			r.Handle("/mcp", h.MCP)
		}
	}

	// OAuth flows are GETs that redirect to external providers — they can't
	// obey the JSON content-type gate the rest of /api/v1 uses, so they
	// live outside that subrouter. Registration is conditional on the
	// per-provider feature flag.
	if h.OAuth != nil {
		if cfg.OAuth.Google.Enabled {
			r.Get("/api/v1/auth/google/start", h.OAuth.GoogleStart)
			r.Get("/api/v1/auth/google/callback", h.OAuth.GoogleCallback)
		}
		if cfg.OAuth.GitHub.Enabled {
			r.Get("/api/v1/auth/github/start", h.OAuth.GitHubStart)
			r.Get("/api/v1/auth/github/callback", h.OAuth.GitHubCallback)
		}
	}

	// Icon proxy — public, GET-only, no JSON content-type gate. Lives
	// under /api/v1 for consistency but bypasses the subrouter's JSON
	// requirement since responses are image/svg+xml.
	if h.Icons != nil {
		r.Get("/api/v1/icons/vendor/*", h.Icons.Proxy)
	}

	// API v1
	r.Route("/api/v1", func(r chi.Router) {
		r.Use(chiMiddleware.AllowContentType("application/json"))

		// Public routes
		r.Group(func(r chi.Router) {
			// Public config — tells the frontend whether signup requires an
			// invite code and which OAuth buttons to render. Kept small:
			// no secrets, no plan data.
			r.Get("/config", func(w http.ResponseWriter, req *http.Request) {
				response.JSON(w, http.StatusOK, map[string]any{
					"registration_mode":    cfg.Registration.Mode,
					"google_oauth_enabled": cfg.OAuth.Google.Enabled,
					"github_oauth_enabled": cfg.OAuth.GitHub.Enabled,
				})
			})
			// Tight per-IP cap on the auth surface — runs IN FRONT OF the
			// handlers themselves so an attacker hammering credential pairs
			// or enumerating emails hits the wall before any DB or bcrypt
			// work happens. Defaults to 30/min/IP via RATE_LIMIT_AUTH_RPM.
			r.Group(func(r chi.Router) {
				if authLimit != nil {
					r.Use(authLimit.Middleware)
				}
				r.Post("/auth/register", h.Auth.Register)
				r.Post("/auth/login", h.Auth.Login)
				r.Post("/auth/forgot-password", h.Auth.ForgotPassword)
				r.Post("/auth/reset-password", h.Auth.ResetPassword)
			})
			r.Post("/invites/validate", h.Invite.ValidateCode)
			r.Post("/invites/request", h.Invite.RequestAccess)
			if h.CloudSync != nil {
				// Public read of shared projects — no auth.
				r.Get("/public/share/{token}", h.CloudSync.LoadShared)
			}
			if h.PublicDemo != nil {
				// Unauth landing-page generator. Own IP throttle
				// (1/min) so credential-stuffing + demo hammering
				// don't share a budget.
				r.Group(func(r chi.Router) {
					if demoLimit != nil {
						r.Use(demoLimit.Middleware)
					}
					r.Post("/public/demo/generate", h.PublicDemo.Generate)
				})
			}
			if h.Pricing != nil {
				r.Get("/pricing", h.Pricing.Get)
			}
			// Payment webhooks are public (signature-verified), not user-authenticated.
			if h.Payments != nil && h.Payments.PaystackEnabled() {
				r.Post("/payments/paystack/webhook", h.Payments.Webhook)
			}
			if h.Payments != nil && h.Payments.PolarEnabled() {
				r.Post("/payments/polar/webhook", h.Payments.PolarWebhook)
			}
		})

		// Public-but-session-aware: the support form attaches user_id when
		// the caller happens to be logged in. CSRF is intentionally NOT
		// required so a brand-new visitor can submit without first hitting
		// any state-establishing endpoint.
		if h.Feedback != nil {
			r.Group(func(r chi.Router) {
				r.Use(mw.OptionalAuthenticate)
				r.Post("/feedback", h.Feedback.Submit)
			})
		}

		// /auth/me is the SPA hydration probe — it must report a clean 401
		// when there is no session rather than triggering the global
		// Authenticate redirect. Uses OptionalAuthenticate + explicit
		// nil-check inside the handler so we can also log diagnostic
		// breadcrumbs (cookie present? validated?) when something looks off.
		r.Group(func(r chi.Router) {
			r.Use(mw.OptionalAuthenticate)
			r.Get("/auth/me", h.Auth.Me)
		})

		// Protected routes — session cookie + CSRF double-submit for mutations.
		r.Group(func(r chi.Router) {
			r.Use(mw.Authenticate)
			r.Use(mw.CSRF)
			r.Post("/auth/logout", h.Auth.Logout)

			if h.Account != nil {
				r.Get("/me/usage", h.Account.Usage)
			}
			if h.APIToken != nil {
				// MCP Personal Access Tokens. Session-only (SPA users
				// mint tokens for external MCP clients). See docs/mcp.md.
				r.Post("/settings/mcp/tokens", h.APIToken.Create)
				r.Get("/settings/mcp/tokens", h.APIToken.List)
				r.Delete("/settings/mcp/tokens/{id}", h.APIToken.Revoke)
			}
			if h.Payments != nil {
				r.Post("/payments/initialize", h.Payments.Initialize)
				r.Post("/payments/cancel", h.Payments.Cancel)
				r.Get("/payments/verify/{reference}", h.Payments.Verify)
				r.Get("/payments", h.Payments.List)
			}
			if h.CloudSync != nil {
				r.Group(func(r chi.Router) {
					r.Get("/sync/projects", h.CloudSync.List)
					r.Get("/sync/projects/{id}", h.CloudSync.Get)
					r.Put("/sync/projects/{id}", h.CloudSync.Put)
					r.Delete("/sync/projects/{id}", h.CloudSync.Delete)
					r.Get("/sync/projects/{id}/preview", h.CloudSync.Preview)
				})
				r.Group(func(r chi.Router) {
					r.Use(mw.RequirePaid)
					r.Get("/sync/projects/{id}/versions", h.CloudSync.ListVersions)
					r.Post("/sync/projects/{id}/versions/{versionId}/restore", h.CloudSync.RestoreVersion)
					r.Post("/sync/projects/{id}/share", h.CloudSync.CreateShare)
					r.Get("/sync/projects/{id}/share", h.CloudSync.ListShares)
					r.Delete("/sync/projects/{id}/share/{token}", h.CloudSync.RevokeShare)
				})
			}

			// Generation endpoints with abuse protection
			r.Group(func(r chi.Router) {
				r.Use(abuseMW.UserRateLimit)
				r.Use(abuseMW.UsageQuota)
				r.Post("/diagrams/generate", h.Diagram.Generate)
				r.Post("/diagrams/generate/stream", h.Diagram.GenerateStream)
				if h.IaC != nil {
					r.Post("/iac/generate", h.IaC.Generate)
				}
				if h.IaCReverse != nil {
					// Reverse ingestion shares the same per-user rate-limit and
					// daily-quota gates as forward generation — an import
					// "costs" the same as an AI generation. See handler comment
					// for why we didn't carve out a separate bucket.
					r.Post("/iac/ingest", h.IaCReverse.Ingest)
				}
			})

			// Admin routes
			r.Group(func(r chi.Router) {
				r.Use(mw.RequireAdmin)
				r.Post("/admin/invites", h.Invite.CreateInvite)
				r.Get("/admin/invites", h.Invite.ListInvites)
				r.Delete("/admin/invites", h.Invite.DeleteInvite)
				r.Get("/admin/invite-requests", h.Invite.ListRequests)
				r.Post("/admin/invite-requests/{id}/approve", h.Invite.ApproveRequest)
				r.Post("/admin/invite-requests/{id}/decline", h.Invite.DeclineRequest)

				if h.Admin != nil {
					r.Get("/admin/users", h.Admin.ListUsers)
					r.Get("/admin/metrics", h.Admin.Metrics)
					r.Get("/admin/audit", h.Admin.ListAudit)
				}
				if h.Feedback != nil {
					r.Get("/admin/feedback", h.Feedback.List)
					r.Post("/admin/feedback/{id}/status", h.Feedback.SetStatus)
				}
				if h.Admin != nil {
					// User-scoped actions. `pathTail` in the handler pulls the
					// user id out of the URL regardless of chi params vs. plain.
					r.Post("/admin/users/{id}/role", h.Admin.SetRole)
					r.Post("/admin/users/{id}/plan", h.Admin.GrantPlan)
					r.Post("/admin/users/{id}/email", h.Admin.EmailUser)
				}
			})
		})
	})

	// Serve frontend SPA in production (when web/dist exists)
	spaHandler := spaFileServer("web/dist")
	if spaHandler != nil {
		r.NotFound(spaHandler.ServeHTTP)
	}

	return r
}

func maxBodySize(maxBytes int64) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r.Body = http.MaxBytesReader(w, r.Body, maxBytes)
			next.ServeHTTP(w, r)
		})
	}
}

// routeBodySize chooses the transport limit before wrapping the request body.
// Wrapping a body a second time cannot increase an earlier MaxBytesReader cap.
func routeBodySize(defaultMax int64) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			limit := defaultMax
			if r.Method == http.MethodPut && strings.HasPrefix(r.URL.Path, "/api/v1/sync/projects/") {
				limit = cloudSyncMaxBodyBytes
			}
			r.Body = http.MaxBytesReader(w, r.Body, limit)
			next.ServeHTTP(w, r)
		})
	}
}

// chainMiddleware wraps a handler with zero-or-more middleware, applied in
// reverse so the first one in the list is the outermost.
func chainMiddleware(h http.HandlerFunc, mws ...func(http.Handler) http.Handler) http.HandlerFunc {
	var handler http.Handler = h
	for i := len(mws) - 1; i >= 0; i-- {
		handler = mws[i](handler)
	}
	return handler.ServeHTTP
}

// optionalBearer runs the token middleware only when an Authorization
// header is present. MCP's initialize + tools/list are public per spec —
// they must succeed without any token — while tools/call needs the auth
// context. The handler itself does the "require auth for this method"
// check; middleware just populates context when possible.
func optionalBearer(auth func(http.Handler) http.Handler) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Header.Get("Authorization") == "" {
				next.ServeHTTP(w, r)
				return
			}
			auth(next).ServeHTTP(w, r)
		})
	}
}

// cloudSyncMaxBodyBytes is the upload ceiling for PUT /sync/projects/{id}.
// Paid quota allows up to ~2 MB diagram + ~700 KB base64 preview; 4 MB
// leaves comfortable headroom for JSON framing without inviting abuse.
// The cloudsync service still enforces tighter per-field quotas after
// the body is parsed, so this cap is purely the transport floor.
const cloudSyncMaxBodyBytes = 4 << 20

// spaFileServer serves static files from dir and falls back to index.html
// for client-side routing. Returns nil if the directory doesn't exist.
func spaFileServer(dir string) http.Handler {
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return nil
	}
	if _, err := os.Stat(absDir); os.IsNotExist(err) {
		return nil
	}

	fileServer := http.FileServer(http.Dir(absDir))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path

		// Skip API routes
		if strings.HasPrefix(path, "/api/") {
			http.NotFound(w, r)
			return
		}

		// Try to serve the file directly
		fullPath := filepath.Join(absDir, filepath.Clean(path))
		if info, err := os.Stat(fullPath); err == nil && !info.IsDir() {
			fileServer.ServeHTTP(w, r)
			return
		}

		// Check if it's a static asset that should 404
		ext := filepath.Ext(path)
		if ext != "" && ext != ".html" {
			http.NotFound(w, r)
			return
		}

		// Fall back to index.html for SPA client-side routing.
		// Force revalidation on every load — Vite emits hashed asset filenames
		// so JS/CSS stay cacheable, but the HTML must always reflect the latest
		// deploy (otherwise stale shells silently break things like the post-
		// OAuth /auth/me hydration probe).
		indexPath := filepath.Join(absDir, "index.html")
		if _, err := fs.Stat(os.DirFS(absDir), "index.html"); err == nil {
			w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
			http.ServeFile(w, r, indexPath)
			return
		}

		http.NotFound(w, r)
	})
}
