package middleware

import (
	"log/slog"
	"net/http"
	"time"

	chiMiddleware "github.com/go-chi/chi/v5/middleware"
)

// Logger logs each request with method, path, status, duration, and request ID.
// Healthcheck probes are skipped to keep the log stream signal-heavy — Coolify
// hits /health every 15s and the noise drowns out anything interesting.
func (m *Middleware) Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Wrap the response writer to capture status code
		ww := chiMiddleware.NewWrapResponseWriter(w, r.ProtoMajor)

		next.ServeHTTP(ww, r)

		if r.URL.Path == "/health" {
			return
		}

		duration := time.Since(start)
		reqID := chiMiddleware.GetReqID(r.Context())

		m.logger.LogAttrs(r.Context(), slog.LevelInfo, "request",
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
			slog.Int("status", ww.Status()),
			slog.Duration("duration", duration),
			slog.String("remote", r.RemoteAddr),
			slog.String("request_id", reqID),
			slog.Int("bytes", ww.BytesWritten()),
		)
	})
}
