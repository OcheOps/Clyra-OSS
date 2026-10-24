package abuse

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/oche/diagram/internal/config"
)

// UserRateLimiter checks whether a specific user has exceeded their
// per-minute request rate for generation endpoints.
type UserRateLimiter interface {
	Allow(ctx context.Context, userID string) bool
}

// --- Redis-backed per-user rate limiter ---

type redisUserRateLimiter struct {
	client *redis.Client
	rate   int
	window time.Duration
}

func NewRedisUserRateLimiter(client *redis.Client, cfg config.AbuseConfig) UserRateLimiter {
	return &redisUserRateLimiter{
		client: client,
		rate:   cfg.UserRatePerMinute,
		window: time.Minute,
	}
}

func (r *redisUserRateLimiter) Allow(ctx context.Context, userID string) bool {
	key := fmt.Sprintf("user_rate:%s", userID)
	now := time.Now().UnixMilli()
	windowStart := now - r.window.Milliseconds()

	pipe := r.client.Pipeline()
	pipe.ZRemRangeByScore(ctx, key, "0", fmt.Sprintf("%d", windowStart))
	countCmd := pipe.ZCard(ctx, key)
	pipe.ZAdd(ctx, key, redis.Z{Score: float64(now), Member: now})
	pipe.Expire(ctx, key, r.window+time.Second)

	if _, err := pipe.Exec(ctx); err != nil {
		return true // fail open
	}

	return countCmd.Val() < int64(r.rate)
}

// --- In-memory per-user rate limiter (fallback) ---

type memoryUserRateLimiter struct {
	mu       sync.Mutex
	visitors map[string]*userVisitor
	rate     float64
	burst    float64
}

type userVisitor struct {
	tokens   float64
	lastSeen time.Time
}

func NewMemoryUserRateLimiter(cfg config.AbuseConfig) UserRateLimiter {
	rl := &memoryUserRateLimiter{
		visitors: make(map[string]*userVisitor),
		rate:     float64(cfg.UserRatePerMinute) / 60.0,
		burst:    float64(cfg.UserRatePerMinute), // allow up to full minute's worth in burst
	}
	go rl.cleanup()
	return rl
}

func (rl *memoryUserRateLimiter) Allow(_ context.Context, userID string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	v, exists := rl.visitors[userID]
	if !exists {
		rl.visitors[userID] = &userVisitor{tokens: rl.burst - 1, lastSeen: now}
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

func (rl *memoryUserRateLimiter) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	for range ticker.C {
		rl.mu.Lock()
		for id, v := range rl.visitors {
			if time.Since(v.lastSeen) > 10*time.Minute {
				delete(rl.visitors, id)
			}
		}
		rl.mu.Unlock()
	}
}
