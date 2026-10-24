package abuse

import (
	"errors"
	"testing"

	"github.com/oche/diagram/internal/config"
)

func newTestValidator(maxLen int) *PromptValidator {
	return NewPromptValidator(config.AbuseConfig{MaxPromptLength: maxLen})
}

func TestPromptValidator_Valid(t *testing.T) {
	v := newTestValidator(200)
	cleaned, err := v.Validate("Build a 3-tier web app with load balancer and database", "pro")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cleaned == "" {
		t.Fatal("cleaned prompt should not be empty")
	}
}

func TestPromptValidator_Empty(t *testing.T) {
	v := newTestValidator(200)
	_, err := v.Validate("", "pro")
	if !errors.Is(err, ErrPromptEmpty) {
		t.Fatalf("expected ErrPromptEmpty, got: %v", err)
	}
}

func TestPromptValidator_WhitespaceOnly(t *testing.T) {
	v := newTestValidator(200)
	_, err := v.Validate("   \t\n   ", "pro")
	if !errors.Is(err, ErrPromptEmpty) {
		t.Fatalf("expected ErrPromptEmpty for whitespace-only, got: %v", err)
	}
}

func TestPromptValidator_TooShort(t *testing.T) {
	v := newTestValidator(200)
	_, err := v.Validate("hi", "pro")
	if !errors.Is(err, ErrPromptTooShort) {
		t.Fatalf("expected ErrPromptTooShort, got: %v", err)
	}
}

func TestPromptValidator_TooLong(t *testing.T) {
	// Free plan is hard-capped at FreePlanPromptMaxLength (2000) regardless of cfg.
	v := newTestValidator(8000)
	long := ""
	for i := 0; i < FreePlanPromptMaxLength+10; i++ {
		long += "x"
	}
	_, err := v.Validate(long, "free")
	if !errors.Is(err, ErrPromptTooLong) {
		t.Fatalf("expected ErrPromptTooLong, got: %v", err)
	}
}

func TestPromptValidator_SanitizesWhitespace(t *testing.T) {
	v := newTestValidator(500)
	cleaned, err := v.Validate("  Build   a\t\n  web   app  ", "pro")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cleaned != "Build a web app" {
		t.Fatalf("expected 'Build a web app', got %q", cleaned)
	}
}

func TestPromptValidator_RemovesControlChars(t *testing.T) {
	v := newTestValidator(500)
	cleaned, err := v.Validate("Build a web\x00 app\x01 now", "pro")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cleaned != "Build a web app now" {
		t.Fatalf("expected 'Build a web app now', got %q", cleaned)
	}
}

// --- TokenGuard tests ---

func TestTokenGuard_ClampDefault(t *testing.T) {
	g := NewTokenGuard(config.AbuseConfig{MaxTokens: 1000})
	if got := g.ClampTokens(0, "free"); got != 1000 {
		t.Fatalf("expected 1000 for 0 input, got %d", got)
	}
}

func TestTokenGuard_ClampOverLimit(t *testing.T) {
	g := NewTokenGuard(config.AbuseConfig{MaxTokens: 1000})
	if got := g.ClampTokens(5000, "free"); got != 1000 {
		t.Fatalf("expected 1000 for 5000 input, got %d", got)
	}
}

func TestTokenGuard_PassThrough(t *testing.T) {
	g := NewTokenGuard(config.AbuseConfig{MaxTokens: 1000})
	if got := g.ClampTokens(500, "free"); got != 500 {
		t.Fatalf("expected 500, got %d", got)
	}
}

func TestTokenGuard_PaidPlanRaisesCeiling(t *testing.T) {
	g := NewTokenGuard(config.AbuseConfig{MaxTokens: 4096})
	if got := g.ClampTokens(0, "pro"); got != PaidPlanMaxTokens {
		t.Fatalf("pro: expected %d for 0 input, got %d", PaidPlanMaxTokens, got)
	}
	if got := g.ClampTokens(0, "byo"); got != PaidPlanMaxTokens {
		t.Fatalf("byo: expected %d for 0 input, got %d", PaidPlanMaxTokens, got)
	}
	if got := g.ClampTokens(20000, "pro"); got != PaidPlanMaxTokens {
		t.Fatalf("pro: expected clamp to %d, got %d", PaidPlanMaxTokens, got)
	}
	if got := g.ClampTokens(0, "free"); got != 4096 {
		t.Fatalf("free: expected 4096 (config baseline), got %d", got)
	}
}

// --- NodeGuard tests ---

func TestNodeGuard_UnderLimit(t *testing.T) {
	g := NewNodeGuard(config.AbuseConfig{MaxNodes: 25})
	if err := g.Check(15); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestNodeGuard_AtLimit(t *testing.T) {
	g := NewNodeGuard(config.AbuseConfig{MaxNodes: 25})
	if err := g.Check(25); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestNodeGuard_OverLimit(t *testing.T) {
	g := NewNodeGuard(config.AbuseConfig{MaxNodes: 25})
	if err := g.Check(30); err == nil {
		t.Fatal("expected error for 30 nodes")
	}
}
