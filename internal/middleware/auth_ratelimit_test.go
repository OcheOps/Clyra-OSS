package middleware

import (
	"context"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

func TestRedisAuthRateLimitNamespacesAreIsolated(t *testing.T) {
	server := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: server.Addr()})
	t.Cleanup(func() { _ = client.Close() })

	authLimiter := NewAuthRateLimit(client, "auth", 2)
	demoLimiter := NewAuthRateLimit(client, "demo", 1)
	ctx := context.Background()

	if !authLimiter.allow(ctx, "203.0.113.10") {
		t.Fatal("first auth request was rejected")
	}
	if !demoLimiter.allow(ctx, "203.0.113.10") {
		t.Fatal("auth request consumed the separate demo bucket")
	}
	if demoLimiter.allow(ctx, "203.0.113.10") {
		t.Fatal("second demo request should exceed the one-request limit")
	}
	if !authLimiter.allow(ctx, "203.0.113.10") {
		t.Fatal("demo request consumed the separate auth bucket")
	}
	if authLimiter.allow(ctx, "203.0.113.10") {
		t.Fatal("third auth request should exceed the two-request limit")
	}
}
