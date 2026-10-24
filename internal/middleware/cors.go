package middleware

import (
	"net/http"
	"strconv"
)

// CORS handles Cross-Origin Resource Sharing for SPA frontends.
//
// Security: we ship the session cookie, so we can NEVER respond with
//
//	Access-Control-Allow-Origin: *
//	Access-Control-Allow-Credentials: true
//
// (the combo is forbidden by the spec and leaks auth across origins).
// Every `*` in the config is therefore a misconfiguration and we refuse
// to reflect credentials in that case — only exact-match origins get credentials.
func (m *Middleware) CORS(next http.Handler) http.Handler {
	allowed := make(map[string]struct{}, len(m.cfg.CORS.AllowedOrigins))
	wildcard := false
	for _, o := range m.cfg.CORS.AllowedOrigins {
		if o == "*" {
			wildcard = true
			continue
		}
		allowed[o] = struct{}{}
	}
	maxAge := strconv.Itoa(m.cfg.CORS.MaxAge)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")

		if _, ok := allowed[origin]; ok && origin != "" {
			// Credentialed, exact origin match — safe.
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Vary", "Origin")
		} else if wildcard && origin != "" {
			// Wildcard: advertise "*" but strip credentials for safety.
			w.Header().Set("Access-Control-Allow-Origin", "*")
		}

		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type, X-CSRF-Token")
		w.Header().Set("Access-Control-Max-Age", maxAge)
		w.Header().Add("Vary", "Access-Control-Request-Method")

		// Handle preflight
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}
