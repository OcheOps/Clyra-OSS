package observability

import (
	"regexp"
)

// scrubPatterns catches common secret shapes that would be a disaster to
// ship to Sentry. Order matters only in the sense that longer, more
// specific matches should come first; a `sk-ant-` prefix is matched before
// the generic `sk-` would otherwise fire.
//
// New shapes get added here, never in local call sites — a central redactor
// means a future bug that accidentally logs a key gets caught regardless of
// which package produced the log line.
var scrubPatterns = []*regexp.Regexp{
	// Google API keys (what Gemini BYO users paste).
	regexp.MustCompile(`AIza[0-9A-Za-z_\-]{35}`),
	// Anthropic BYO keys.
	regexp.MustCompile(`sk-ant-[A-Za-z0-9_\-]{20,}`),
	// OpenAI BYO keys — covers legacy `sk-` and project-scoped `sk-proj-`.
	regexp.MustCompile(`sk-(?:proj-)?[A-Za-z0-9_\-]{32,}`),
	// Clyra Personal Access Tokens.
	regexp.MustCompile(`clyra_pat_[0-9a-f]{32,}`),
	// AWS access key IDs — the paired secret can also show up in signed URLs.
	regexp.MustCompile(`AKIA[0-9A-Z]{16}`),
	regexp.MustCompile(`ASIA[0-9A-Z]{16}`),
	// Polar webhook secret.
	regexp.MustCompile(`whsec_[A-Za-z0-9+/=]{20,}`),
	// Generic Authorization: Bearer line (headers that slipped into a log).
	regexp.MustCompile(`(?i)Bearer\s+[A-Za-z0-9_\-\.]{20,}`),
}

const redactedPlaceholder = "[REDACTED]"

// scrubString replaces any match of a known secret shape with
// [REDACTED]. Safe to call on arbitrary strings — if no match is found
// the original string is returned unchanged (allocating a new string
// only when a match actually fires).
func scrubString(s string) string {
	out := s
	for _, re := range scrubPatterns {
		out = re.ReplaceAllString(out, redactedPlaceholder)
	}
	return out
}

// scrubValue applies scrubString to any attribute value we're about to
// hand to Sentry. Non-string values are returned as-is; errors are
// wrapped in a wrapper so their message is scrubbed while the type is
// kept recognisable in Sentry.
func scrubValue(v any) any {
	switch typed := v.(type) {
	case string:
		return scrubString(typed)
	case error:
		return scrubbedError{inner: typed, msg: scrubString(typed.Error())}
	default:
		return v
	}
}

// scrubbedError lets us hand Sentry a redacted message while keeping the
// original error chain intact for anything downstream that unwraps it.
type scrubbedError struct {
	inner error
	msg   string
}

func (e scrubbedError) Error() string { return e.msg }
func (e scrubbedError) Unwrap() error { return e.inner }
