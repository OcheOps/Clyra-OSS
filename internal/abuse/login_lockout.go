package abuse

import (
	"errors"
	"strings"
	"sync"
	"time"
)

// ErrAccountLocked is returned by LoginLockout.Check when an account is
// temporarily blocked from login attempts.
var ErrAccountLocked = errors.New("account_locked")

// LoginLockout guards against credential stuffing by slowing down repeated
// failures against the same email address. After `threshold` failures inside
// `window`, the email is locked for `cooldown`. A successful login resets it.
//
// This is intentionally in-memory: the same user coming from many IPs still
// hits one counter. For horizontal scale behind multiple instances, move to
// Redis (the shape matches abuse.RedisUserRateLimiter).
type LoginLockout struct {
	threshold int
	window    time.Duration
	cooldown  time.Duration

	mu      sync.Mutex
	entries map[string]*lockoutEntry
	// lastSweep bounds the entries map; we GC stale rows on every Record call.
	lastSweep time.Time
}

type lockoutEntry struct {
	failures    int
	firstFailAt time.Time
	lockedUntil time.Time
}

func NewLoginLockout(threshold int, window, cooldown time.Duration) *LoginLockout {
	if threshold < 1 {
		threshold = 5
	}
	if window <= 0 {
		window = 15 * time.Minute
	}
	if cooldown <= 0 {
		cooldown = 15 * time.Minute
	}
	return &LoginLockout{
		threshold: threshold,
		window:    window,
		cooldown:  cooldown,
		entries:   make(map[string]*lockoutEntry),
	}
}

// Check reports whether login attempts are currently allowed for this email.
func (l *LoginLockout) Check(email string) error {
	key := normalize(email)
	if key == "" {
		return nil
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	e := l.entries[key]
	if e == nil {
		return nil
	}
	if time.Now().Before(e.lockedUntil) {
		return ErrAccountLocked
	}
	return nil
}

// RecordFailure increments the failure counter and locks the account when
// the threshold is hit.
func (l *LoginLockout) RecordFailure(email string) {
	key := normalize(email)
	if key == "" {
		return
	}
	now := time.Now()
	l.mu.Lock()
	defer l.mu.Unlock()
	l.gcLocked(now)

	e := l.entries[key]
	if e == nil || now.Sub(e.firstFailAt) > l.window {
		e = &lockoutEntry{firstFailAt: now}
		l.entries[key] = e
	}
	e.failures++
	if e.failures >= l.threshold {
		e.lockedUntil = now.Add(l.cooldown)
	}
}

// RecordSuccess clears any tracked failures for this email.
func (l *LoginLockout) RecordSuccess(email string) {
	key := normalize(email)
	if key == "" {
		return
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	delete(l.entries, key)
}

func (l *LoginLockout) gcLocked(now time.Time) {
	if now.Sub(l.lastSweep) < time.Minute {
		return
	}
	l.lastSweep = now
	for k, e := range l.entries {
		if now.Before(e.lockedUntil) {
			continue
		}
		if now.Sub(e.firstFailAt) > l.window {
			delete(l.entries, k)
		}
	}
}

func normalize(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}
