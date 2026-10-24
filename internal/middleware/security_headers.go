package middleware

import "net/http"

// SecurityHeaders adds a conservative default set of response headers.
// The CSP allows:
//   - self for scripts/styles/connect (plus 'unsafe-inline' for Tailwind utility styles)
//   - self + data:/blob:/https: for images (covers S3 signed URLs + previews)
//   - fonts.googleapis.com / fonts.gstatic.com (Inter + JetBrains Mono from index.html)
//   - api.fontshare.com / cdn.fontshare.com (Switzer for the v2 design system)
//   - api.iconify.design (provider logo SVGs — Azure/Cloudflare/k8s)
//   - icon.icepanel.io (official AWS + GCP architecture SVGs) — uses
//     img-src https: wildcard so no explicit allow needed
//   - api.country.is and ip-api.com for client-side geo detection
//
// HSTS is only sent in production to avoid pinning dev HTTP localhost.
func (m *Middleware) SecurityHeaders(next http.Handler) http.Handler {
	isProd := m.cfg.Env == "production"
	csp := "default-src 'self'; " +
		"script-src 'self'; " +
		"style-src 'self' 'unsafe-inline' https://fonts.googleapis.com https://api.fontshare.com; " +
		"font-src 'self' https://fonts.gstatic.com https://cdn.fontshare.com; " +
		"img-src 'self' data: blob: https:; " +
		"connect-src 'self' https://api.country.is http://ip-api.com https://ip-api.com https://api.fontshare.com https://cdn.fontshare.com https://api.iconify.design https://*.ingest.sentry.io https://*.ingest.de.sentry.io https://*.ingest.us.sentry.io; " +
		"frame-ancestors 'none'; " +
		"base-uri 'self'; " +
		"form-action 'self'"

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("X-Frame-Options", "DENY")
		h.Set("Referrer-Policy", "strict-origin-when-cross-origin")
		h.Set("Permissions-Policy", "camera=(), microphone=(), geolocation=(), interest-cohort=()")
		h.Set("Content-Security-Policy", csp)
		if isProd {
			// 6 months HSTS — long enough to matter, short enough to recover from misconfig.
			h.Set("Strict-Transport-Security", "max-age=15552000; includeSubDomains")
		}
		next.ServeHTTP(w, r)
	})
}
