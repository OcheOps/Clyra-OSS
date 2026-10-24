package abuse

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/oche/diagram/internal/config"
)

// stubQuotaService simulates quota behavior without a database.
type stubQuotaService struct {
	counts map[string]int
	limit  int
}

func newStubQuotaService(limit int) *stubQuotaService {
	return &stubQuotaService{counts: make(map[string]int), limit: limit}
}

func (s *stubQuotaService) check(userID string) error {
	if s.counts[userID] >= s.limit {
		return ErrDailyLimitReached
	}
	return nil
}

func (s *stubQuotaService) increment(userID string) {
	s.counts[userID]++
}

func newTestGuard(dailyLimit int) *Guard {
	cfg := config.AbuseConfig{
		UserRatePerMinute:       10,
		DailyLimit:              dailyLimit,
		MaxPromptLength:         500,
		MaxTokens:               1000,
		ProviderTimeout:         30 * time.Second,
		MaxNodes:                20,
		CircuitBreakerThreshold: 3,
		CircuitBreakerCooldown:  50 * time.Millisecond,
	}
	quotaRepo := &QuotaRepository{pool: nil} // won't be called directly
	quotaService := NewQuotaService(quotaRepo, dailyLimit, nil)
	return NewGuard(cfg, quotaService, nil, nil)
}

// --- PreCheck tests ---

func TestGuard_PreCheck_ValidPrompt(t *testing.T) {
	g := newTestGuard(30)
	cleaned, err := g.PreCheck(context.Background(), "user-1", "free", "Build a 3-tier web app")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cleaned != "Build a 3-tier web app" {
		t.Fatalf("expected cleaned prompt, got %q", cleaned)
	}
}

func TestGuard_PreCheck_EmptyPrompt(t *testing.T) {
	g := newTestGuard(30)
	_, err := g.PreCheck(context.Background(), "user-1", "free", "")
	if !errors.Is(err, ErrPromptEmpty) {
		t.Fatalf("expected ErrPromptEmpty, got: %v", err)
	}
}

func TestGuard_PreCheck_CircuitOpen(t *testing.T) {
	g := newTestGuard(30)

	// Trip the circuit
	g.CircuitBreaker.RecordFailure()
	g.CircuitBreaker.RecordFailure()
	g.CircuitBreaker.RecordFailure()

	_, err := g.PreCheck(context.Background(), "user-1", "free", "Build something")
	if !errors.Is(err, ErrCircuitOpen) {
		t.Fatalf("expected ErrCircuitOpen, got: %v", err)
	}
}

func TestGuard_PreCheck_QuotaExhausted(t *testing.T) {
	g := newTestGuard(0) // fallback limit = 0 means every plan hits quota immediately
	_, err := g.PreCheck(context.Background(), "user-1", "free", "Build a web app with load balancer")
	if !errors.Is(err, ErrDailyLimitReached) {
		t.Fatalf("expected ErrDailyLimitReached, got: %v", err)
	}
}

func TestGuard_PreCheck_PromptTooLong(t *testing.T) {
	g := newTestGuard(30) // free plan capped at FreePlanPromptMaxLength regardless of cfg

	longPrompt := ""
	for range FreePlanPromptMaxLength + 100 {
		longPrompt += "x"
	}
	_, err := g.PreCheck(context.Background(), "user-1", "free", longPrompt)
	if !errors.Is(err, ErrPromptTooLong) {
		t.Fatalf("expected ErrPromptTooLong, got: %v", err)
	}
}

// --- PostSuccess tests ---

func TestGuard_PostSuccess_ResetsCircuit(t *testing.T) {
	g := newTestGuard(30)

	g.CircuitBreaker.RecordFailure()
	g.CircuitBreaker.RecordFailure()

	g.PostSuccess(context.Background(), "user-1", "managed_bedrock", 10, 5*time.Second)

	if g.CircuitBreaker.Failures() != 0 {
		t.Fatal("PostSuccess should reset circuit breaker failures")
	}
	if g.CircuitBreaker.State() != CircuitClosed {
		t.Fatal("PostSuccess should close the circuit")
	}
}

// --- PostFailure tests ---

func TestGuard_PostFailure_IncrementsCircuit(t *testing.T) {
	g := newTestGuard(30)

	g.PostFailure(context.Background(), "user-1", "managed_bedrock", 5*time.Second)

	if g.CircuitBreaker.Failures() != 1 {
		t.Fatalf("expected 1 failure, got %d", g.CircuitBreaker.Failures())
	}
}

func TestGuard_PostFailure_TripsCircuitAfterThreshold(t *testing.T) {
	g := newTestGuard(30) // threshold = 3

	g.PostFailure(context.Background(), "user-1", "bedrock", 1*time.Second)
	g.PostFailure(context.Background(), "user-1", "bedrock", 1*time.Second)
	g.PostFailure(context.Background(), "user-1", "bedrock", 1*time.Second)

	if g.CircuitBreaker.State() != CircuitOpen {
		t.Fatalf("expected open after 3 failures, got %s", g.CircuitBreaker.State())
	}

	// Next PreCheck should fail
	_, err := g.PreCheck(context.Background(), "user-2", "free", "Build something nice")
	if !errors.Is(err, ErrCircuitOpen) {
		t.Fatalf("expected ErrCircuitOpen, got: %v", err)
	}
}

// --- TokenGuard integration ---

func TestGuard_TokenClamp(t *testing.T) {
	g := newTestGuard(30) // maxTokens = 1000

	if got := g.TokenGuard.ClampTokens(5000, "free"); got != 1000 {
		t.Fatalf("expected 1000, got %d", got)
	}
	if got := g.TokenGuard.ClampTokens(500, "free"); got != 500 {
		t.Fatalf("expected 500, got %d", got)
	}
	if got := g.TokenGuard.ClampTokens(0, "free"); got != 1000 {
		t.Fatalf("expected 1000 for 0, got %d", got)
	}
}
