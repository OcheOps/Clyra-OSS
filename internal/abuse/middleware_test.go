package abuse

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/oche/diagram/internal/auth"
	"github.com/oche/diagram/internal/config"
)

// stubRateLimiter with configurable behavior.
type stubRateLimiter struct {
	allowed bool
}

func (s *stubRateLimiter) Allow(_ context.Context, _ string) bool {
	return s.allowed
}

// stubHandler records whether it was called.
type stubHandler struct {
	called bool
}

func (h *stubHandler) ServeHTTP(w http.ResponseWriter, _ *http.Request) {
	h.called = true
	w.WriteHeader(http.StatusOK)
}

func newTestRequest(userID string) *http.Request {
	r := httptest.NewRequest(http.MethodPost, "/api/v1/diagrams/generate", nil)
	if userID != "" {
		ctx := auth.NewContext(r.Context(), &auth.SessionUser{UserID: userID, Email: "test@test.com", Role: "user"})
		r = r.WithContext(ctx)
	}
	return r
}

// --- UserRateLimit middleware tests ---

func TestMiddleware_UserRateLimit_Allows(t *testing.T) {
	mw := NewMiddleware(&stubRateLimiter{allowed: true}, nil, nil, nil)
	handler := &stubHandler{}
	recorder := httptest.NewRecorder()

	mw.UserRateLimit(handler).ServeHTTP(recorder, newTestRequest("user-1"))

	if !handler.called {
		t.Fatal("handler should have been called")
	}
	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}
}

func TestMiddleware_UserRateLimit_Blocks(t *testing.T) {
	mw := NewMiddleware(&stubRateLimiter{allowed: false}, nil, nil, nil)
	handler := &stubHandler{}
	recorder := httptest.NewRecorder()

	mw.UserRateLimit(handler).ServeHTTP(recorder, newTestRequest("user-1"))

	if handler.called {
		t.Fatal("handler should NOT have been called")
	}
	if recorder.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429, got %d", recorder.Code)
	}
}

func TestMiddleware_UserRateLimit_NoUser_PassesThrough(t *testing.T) {
	mw := NewMiddleware(&stubRateLimiter{allowed: false}, nil, nil, nil)
	handler := &stubHandler{}
	recorder := httptest.NewRecorder()

	// No user in context
	mw.UserRateLimit(handler).ServeHTTP(recorder, newTestRequest(""))

	if !handler.called {
		t.Fatal("should pass through when no user (auth middleware handles it)")
	}
}

// --- UsageQuota middleware tests ---

func TestMiddleware_UsageQuota_Allows(t *testing.T) {
	quotaRepo := &QuotaRepository{pool: nil}
	quotaService := NewQuotaService(quotaRepo, 100, nil) // high limit
	mw := NewMiddleware(nil, quotaService, nil, nil)
	handler := &stubHandler{}
	recorder := httptest.NewRecorder()

	// QuotaService.Check will fail open (nil pool) so request passes
	mw.UsageQuota(handler).ServeHTTP(recorder, newTestRequest("user-1"))

	if !handler.called {
		t.Fatal("handler should have been called")
	}
}

func TestMiddleware_UsageQuota_NoUser_PassesThrough(t *testing.T) {
	quotaService := NewQuotaService(&QuotaRepository{pool: nil}, 0, nil)
	mw := NewMiddleware(nil, quotaService, nil, nil)
	handler := &stubHandler{}
	recorder := httptest.NewRecorder()

	mw.UsageQuota(handler).ServeHTTP(recorder, newTestRequest(""))

	if !handler.called {
		t.Fatal("should pass through when no user")
	}
}

// --- Integration: full rate limiter behavior ---

func TestMiddleware_UserRateLimit_PerUserIsolation(t *testing.T) {
	cfg := config.AbuseConfig{UserRatePerMinute: 1}
	rl := NewMemoryUserRateLimiter(cfg)
	mw := NewMiddleware(rl, nil, nil, nil)

	// user-a first request passes
	handler1 := &stubHandler{}
	rec1 := httptest.NewRecorder()
	mw.UserRateLimit(handler1).ServeHTTP(rec1, newTestRequest("user-a"))
	if !handler1.called {
		t.Fatal("user-a first request should pass")
	}

	// user-a second request blocked
	handler2 := &stubHandler{}
	rec2 := httptest.NewRecorder()
	mw.UserRateLimit(handler2).ServeHTTP(rec2, newTestRequest("user-a"))
	if handler2.called {
		t.Fatal("user-a second request should be blocked")
	}
	if rec2.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429, got %d", rec2.Code)
	}

	// user-b still passes
	handler3 := &stubHandler{}
	rec3 := httptest.NewRecorder()
	mw.UserRateLimit(handler3).ServeHTTP(rec3, newTestRequest("user-b"))
	if !handler3.called {
		t.Fatal("user-b should not be affected by user-a")
	}
}
