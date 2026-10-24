// Package observability wires slog to Sentry so any Error-level log
// becomes a Sentry event without having to sprinkle CaptureException
// calls at every error site.
package observability

import (
	"context"
	"errors"
	"log/slog"

	"github.com/getsentry/sentry-go"
)

// SentryHandler wraps an slog.Handler and forwards Error-level records
// to Sentry. The underlying handler still receives every record, so
// stdout logging is unaffected.
type SentryHandler struct {
	inner slog.Handler
}

func NewSentryHandler(inner slog.Handler) *SentryHandler {
	return &SentryHandler{inner: inner}
}

func (h *SentryHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.inner.Enabled(ctx, level)
}

func (h *SentryHandler) Handle(ctx context.Context, r slog.Record) error {
	if r.Level >= slog.LevelError {
		h.capture(r)
	}
	return h.inner.Handle(ctx, r)
}

func (h *SentryHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &SentryHandler{inner: h.inner.WithAttrs(attrs)}
}

func (h *SentryHandler) WithGroup(name string) slog.Handler {
	return &SentryHandler{inner: h.inner.WithGroup(name)}
}

// capture builds a Sentry event from the log record. If an "error"
// attribute is present we report it via CaptureException so the
// fingerprint stays tied to the error type; otherwise fall back to
// CaptureMessage with the log line.
func (h *SentryHandler) capture(r slog.Record) {
	var attached error
	extras := make(map[string]any, r.NumAttrs())
	r.Attrs(func(a slog.Attr) bool {
		v := a.Value.Any()
		if a.Key == "error" {
			if e, ok := v.(error); ok {
				// Wrap so the stringified error reaching Sentry is scrubbed
				// but the original error chain stays available to any
				// downstream unwrapping.
				attached = scrubValue(e).(error)
				return true
			}
			if s, ok := v.(string); ok {
				attached = errors.New(scrubString(s))
				return true
			}
		}
		extras[a.Key] = scrubValue(v)
		return true
	})

	// Scrub the message itself — a careless slog.Info("…", "key", k) that
	// accidentally interpolated a secret into the format string would
	// otherwise leak at this point.
	message := scrubString(r.Message)
	ctx := sentry.Context{"message": message}
	for k, v := range extras {
		ctx[k] = v
	}
	hub := sentry.CurrentHub().Clone()
	hub.WithScope(func(scope *sentry.Scope) {
		scope.SetLevel(sentry.LevelError)
		scope.SetContext("log", ctx)
		if attached != nil {
			hub.CaptureException(attached)
			return
		}
		hub.CaptureMessage(message)
	})
}
