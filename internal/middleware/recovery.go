package middleware

import (
	"net/http"
	"runtime/debug"

	"github.com/oche/diagram/pkg/response"
)

// Recovery catches panics in handlers, logs the stack trace,
// and returns a 500 instead of crashing the server.
func (m *Middleware) Recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				m.logger.Error("panic recovered",
					"error", err,
					"stack", string(debug.Stack()),
					"method", r.Method,
					"path", r.URL.Path,
				)
				response.Error(w, http.StatusInternalServerError, "internal server error")
			}
		}()

		next.ServeHTTP(w, r)
	})
}
