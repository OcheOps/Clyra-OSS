package middleware

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/oche/diagram/internal/config"
	"github.com/oche/diagram/pkg/response"
)

// RateLimiter is the interface for rate limiting backends.
type RateLimiter interface {
	Allow(ctx context.Context, key string) bool
}

// --- Redis rate limiter ---

type redisRateLimiter struct {
	client *redis.Client
	rate   int
	window time.Duration
}

// NewRedisRateLimiter creates a rate limiter backed by Redis using a sliding window counter.
func NewRedisRateLimiter(client *redis.Client, cfg config.RateLimitConfig) RateLimiter {
	return &redisRateLimiter{
		client: client,
		rate:   cfg.RequestsPerMinute,
		window: time.Minute,
	}
}

func (r *redisRateLimiter) Allow(ctx context.Context, key string) bool {
	redisKey := fmt.Sprintf("ratelimit:%s", key)
	now := time.Now().UnixMilli()
	windowStart := now - r.window.Milliseconds()

	pipe := r.client.Pipeline()

	// Remove entries outside the window
	pipe.ZRemRangeByScore(ctx, redisKey, "0", fmt.Sprintf("%d", windowStart))

	// Count entries in current window
	countCmd := pipe.ZCard(ctx, redisKey)

	// Add current request
	pipe.ZAdd(ctx, redisKey, redis.Z{Score: float64(now), Member: now})

	// Set TTL so keys don't linger forever
	pipe.Expire(ctx, redisKey, r.window+time.Second)

	_, err := pipe.Exec(ctx)
	if err != nil {
		// On Redis failure, allow the request (fail open)
		return true
	}

	return countCmd.Val() < int64(r.rate)
}

// --- In-memory rate limiter (fallback) ---

type visitor struct {
	tokens   float64
	lastSeen time.Time
}

type memoryRateLimiter struct {
	mu       sync.Mutex
	visitors map[string]*visitor
	rate     float64
	burst    float64
}

// NewMemoryRateLimiter creates an in-memory token bucket rate limiter.
func NewMemoryRateLimiter(cfg config.RateLimitConfig) RateLimiter {
	rl := &memoryRateLimiter{
		visitors: make(map[string]*visitor),
		rate:     float64(cfg.RequestsPerMinute) / 60.0,
		burst:    float64(cfg.Burst),
	}
	go rl.cleanup()
	return rl
}

func (rl *memoryRateLimiter) Allow(_ context.Context, key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	v, exists := rl.visitors[key]
	if !exists {
		rl.visitors[key] = &visitor{tokens: rl.burst - 1, lastSeen: now}
		return true
	}

	elapsed := now.Sub(v.lastSeen).Seconds()
	v.tokens += elapsed * rl.rate
	if v.tokens > rl.burst {
		v.tokens = rl.burst
	}
	v.lastSeen = now

	if v.tokens < 1 {
		return false
	}

	v.tokens--
	return true
}

func (rl *memoryRateLimiter) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	for range ticker.C {
		rl.mu.Lock()
		for ip, v := range rl.visitors {
			if time.Since(v.lastSeen) > 10*time.Minute {
				delete(rl.visitors, ip)
			}
		}
		rl.mu.Unlock()
	}
}

// RateLimit applies per-IP rate limiting using the configured backend.
func (m *Middleware) RateLimit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !m.limiter.Allow(r.Context(), r.RemoteAddr) {
			response.Error(w, http.StatusTooManyRequests, "rate limit exceeded")
			return
		}
		next.ServeHTTP(w, r)
	})
}
