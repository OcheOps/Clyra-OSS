package middleware

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/oche/diagram/pkg/response"
)

// AuthRateLimit is a tight per-IP limiter for the unauthenticated auth
// endpoints (login, register, forgot-password, reset-password). It exists
// separately from the general /api rate limit because the auth surface
// needs a much smaller cap to defend against:
//
//   - Credential stuffing — one IP trying many email/password pairs against
//     /auth/login. Per-user lockout doesn't help here because each email is
//     tried at most once.
//   - Email enumeration — /auth/register returning 409 for known emails
//     means an unbounded prober could harvest the user list quickly.
//   - Reset-token guessing — repeatedly POSTing /auth/reset-password with
//     guessed tokens. Tokens are 256-bit so guessing is infeasible, but the
//     limiter caps the per-IP throughput regardless.
//
// Backed by Redis when available (cluster-correct), falls back to an
// in-process map otherwise. Keys are namespaced "authrl:<namespace>:<ip>" so this
// limiter doesn't collide with the general "ratelimit:<ip>" bucket.
type AuthRateLimit struct {
	namespace string
	rate      int
	window    time.Duration
	redis     *redis.Client

	// Memory fallback fields (used only when redis is nil).
	mu  sync.Mutex
	mem map[string]*authBucket
}

type authBucket struct {
	count    int
	resetAt  time.Time
	lastSeen time.Time
}

// NewAuthRateLimit returns a limiter scoped to the auth surface. ratePerMin
// is the cap per IP within a rolling 1-minute window. Pass redisClient = nil
// to use the in-process map; the map self-cleans on access.
func NewAuthRateLimit(redisClient *redis.Client, namespace string, ratePerMin int) *AuthRateLimit {
	if ratePerMin <= 0 {
		ratePerMin = 30
	}
	if namespace == "" {
		namespace = "auth"
	}
	return &AuthRateLimit{
		namespace: namespace,
		rate:      ratePerMin,
		window:    time.Minute,
		redis:     redisClient,
		mem:       make(map[string]*authBucket),
	}
}

// Middleware returns the chi-compatible handler. 429 with a generic message
// — we do not echo the limit value, to avoid telling a tuner what wall to
// stay under.
func (l *AuthRateLimit) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := clientIP(r)
		if !l.allow(r.Context(), ip) {
			response.Error(w, http.StatusTooManyRequests, "too many attempts — please slow down")
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (l *AuthRateLimit) allow(ctx context.Context, ip string) bool {
	if l.redis != nil {
		return l.allowRedis(ctx, ip)
	}
	return l.allowMemory(ip)
}

func (l *AuthRateLimit) allowRedis(ctx context.Context, ip string) bool {
	key := "authrl:" + l.namespace + ":" + ip
	now := time.Now().UnixMilli()
	windowStart := now - l.window.Milliseconds()
	memberBytes := make([]byte, 8)
	if _, err := rand.Read(memberBytes); err != nil {
		return true
	}
	member := hex.EncodeToString(memberBytes)
	allowed, err := authRateLimitScript.Run(ctx, l.redis, []string{key},
		windowStart, now, member, l.window.Milliseconds()+1000, l.rate,
	).Int()
	if err != nil {
		// Fail open — better to let a legitimate user log in than to lock
		// everyone out during a Redis blip.
		return true
	}
	return allowed == 1
}

var authRateLimitScript = redis.NewScript(`
redis.call('ZREMRANGEBYSCORE', KEYS[1], '0', ARGV[1])
local count = redis.call('ZCARD', KEYS[1])
if count >= tonumber(ARGV[5]) then return 0 end
redis.call('ZADD', KEYS[1], ARGV[2], ARGV[3])
redis.call('PEXPIRE', KEYS[1], ARGV[4])
return 1
`)

func (l *AuthRateLimit) allowMemory(ip string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	b, ok := l.mem[ip]
	if !ok || now.After(b.resetAt) {
		l.mem[ip] = &authBucket{count: 1, resetAt: now.Add(l.window), lastSeen: now}
		// Opportunistic cleanup: every Nth call, prune anyone we haven't
		// seen for 10× the window. Keeps the map bounded under churn.
		if len(l.mem)%128 == 0 {
			cutoff := now.Add(-10 * l.window)
			for k, v := range l.mem {
				if v.lastSeen.Before(cutoff) {
					delete(l.mem, k)
				}
			}
		}
		return true
	}
	b.lastSeen = now
	if b.count >= l.rate {
		return false
	}
	b.count++
	return true
}

// clientIP picks a useful identifier from the request. It prefers the
// X-Forwarded-For first hop when present (we trust the platform proxy in
// front of us), falling back to RemoteAddr otherwise. Always returns a
// non-empty string so the limiter never keys on "".
func clientIP(r *http.Request) string {
	if v := r.Header.Get("X-Forwarded-For"); v != "" {
		for i := 0; i < len(v); i++ {
			if v[i] == ',' {
				return v[:i]
			}
		}
		return v
	}
	if v := r.Header.Get("X-Real-IP"); v != "" {
		return v
	}
	return r.RemoteAddr
}
