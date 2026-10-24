package mcp

import (
	"context"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

// DailyLimiter caps calls that spend Clyra-managed inference. The caller
// decides which tools are chargeable; local parsers never touch this limiter.
type DailyLimiter interface {
	Allow(ctx context.Context, userID string) (ok bool, remaining int, resetAt time.Time)
}

type memoryDailyLimiter struct {
	mu       sync.Mutex
	limit    int
	counters map[string]dailyCounter
}

type dailyCounter struct {
	day   string
	count int
}

func NewMemoryDailyLimiter(perDay int) DailyLimiter {
	return &memoryDailyLimiter{limit: perDay, counters: map[string]dailyCounter{}}
}

func (m *memoryDailyLimiter) Allow(_ context.Context, userID string) (bool, int, time.Time) {
	nowUTC := time.Now().UTC()
	reset := nextUTCMidnight(nowUTC)
	if userID == "" {
		return false, 0, reset
	}
	if m.limit <= 0 {
		return true, -1, reset
	}
	today := nowUTC.Format("20060102")
	m.mu.Lock()
	defer m.mu.Unlock()
	c := m.counters[userID]
	if c.day != today {
		c = dailyCounter{day: today}
	}
	if c.count >= m.limit {
		return false, 0, reset
	}
	c.count++
	m.counters[userID] = c
	return true, m.limit - c.count, reset
}

type redisDailyLimiter struct {
	client *redis.Client
	limit  int
}

func NewRedisDailyLimiter(client *redis.Client, perDay int) DailyLimiter {
	if client == nil {
		return NewMemoryDailyLimiter(perDay)
	}
	return &redisDailyLimiter{client: client, limit: perDay}
}

var allowDailyScript = redis.NewScript(`
local count = redis.call('INCR', KEYS[1])
if count == 1 then redis.call('EXPIREAT', KEYS[1], ARGV[1]) end
return count
`)

func (l *redisDailyLimiter) Allow(ctx context.Context, userID string) (bool, int, time.Time) {
	nowUTC := time.Now().UTC()
	reset := nextUTCMidnight(nowUTC)
	if userID == "" {
		return false, 0, reset
	}
	if l.limit <= 0 {
		return true, -1, reset
	}
	key := "mcp:usage:" + userID + ":" + nowUTC.Format("20060102")
	count, err := allowDailyScript.Run(ctx, l.client, []string{key}, reset.Unix()).Int()
	if err != nil {
		// The regular generation abuse guard remains active during Redis failure.
		return true, -1, reset
	}
	remaining := l.limit - count
	if remaining < 0 {
		remaining = 0
	}
	return count <= l.limit, remaining, reset
}

func nextUTCMidnight(now time.Time) time.Time {
	y, m, d := now.Date()
	return time.Date(y, m, d+1, 0, 0, 0, 0, time.UTC)
}
