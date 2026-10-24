package abuse

import (
	"context"
	"testing"

	"github.com/oche/diagram/internal/config"
)

func TestMemoryUserRateLimiter_AllowsUpToLimit(t *testing.T) {
	cfg := config.AbuseConfig{UserRatePerMinute: 3}
	rl := NewMemoryUserRateLimiter(cfg)
	ctx := context.Background()

	for i := 0; i < 3; i++ {
		if !rl.Allow(ctx, "user-1") {
			t.Fatalf("request %d should be allowed", i+1)
		}
	}
}

func TestMemoryUserRateLimiter_BlocksOverLimit(t *testing.T) {
	cfg := config.AbuseConfig{UserRatePerMinute: 2}
	rl := NewMemoryUserRateLimiter(cfg)
	ctx := context.Background()

	rl.Allow(ctx, "user-1") // 1
	rl.Allow(ctx, "user-1") // 2

	if rl.Allow(ctx, "user-1") {
		t.Fatal("3rd request should be blocked")
	}
}

func TestMemoryUserRateLimiter_PerUserIsolation(t *testing.T) {
	cfg := config.AbuseConfig{UserRatePerMinute: 1}
	rl := NewMemoryUserRateLimiter(cfg)
	ctx := context.Background()

	if !rl.Allow(ctx, "user-a") {
		t.Fatal("user-a first request should be allowed")
	}

	// user-a is now rate limited, but user-b should still work
	if !rl.Allow(ctx, "user-b") {
		t.Fatal("user-b should not be affected by user-a's limit")
	}

	if rl.Allow(ctx, "user-a") {
		t.Fatal("user-a second request should be blocked")
	}
}
