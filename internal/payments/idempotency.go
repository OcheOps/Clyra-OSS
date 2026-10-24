package payments

import (
	"context"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
)

// Two-phase webhook idempotency.
//
// The naive single-phase model (SETNX with 24h TTL, treat any existing key
// as duplicate) has a subtle failure mode: if we claim the key, then fail
// to process the payment (DB blip, plan-upgrade error, panic), the provider
// retries — but the retry hits the claimed key and gets a 200 "duplicate"
// ack even though we never actually processed it. The customer paid and
// their plan stays free. See audit finding #1.
//
// Two-phase fix:
//   1. Claim(): SETNX an "in_progress" marker with a SHORT TTL (5 min). The
//      short TTL is the auto-recovery window: if the worker crashes or the
//      DB write fails, the next provider retry after 5m sees the marker
//      expired and gets to try again. Concurrent retries within the window
//      race — one wins the lock, the others get ErrProcessingInFlight and
//      must be re-sent (we return 5xx from the handler).
//   2. Complete(): SET the key to "complete" with the full 24h retention.
//      Only calls to Complete() count as "definitively processed" for
//      idempotency purposes; all subsequent retries against a completed
//      key get ErrAlreadyComplete → 200 duplicate ack.
//
// Failure semantics:
//   - Processing succeeds → Complete() → future retries: ErrAlreadyComplete → 200
//   - Processing fails    → don't call Complete → 5m later the key expires
//                          → next retry gets a fresh claim → chance to succeed
//   - Two retries race    → one gets Claim=nil, the other gets
//                           ErrProcessingInFlight (handler returns 5xx →
//                           provider retries later after our worker finishes)
//   - Redis down          → fail open (return nil) — better a rare duplicate
//                          than dropped legitimate webhooks. Service-layer
//                          idempotency (StatusPaid short-circuit in
//                          HandleWebhook, GetByOrderID dedup in HandleRenewal)
//                          is the second line of defence.

var (
	// ErrAlreadyComplete is returned by Claim when the event has already
	// been processed successfully. Handler should 200-ack as a duplicate.
	ErrAlreadyComplete = errors.New("webhook already processed")

	// ErrProcessingInFlight is returned by Claim when another worker is
	// currently processing this event (or a previous attempt is still
	// within the 5-minute lock window). Handler should return 5xx so the
	// provider retries later.
	ErrProcessingInFlight = errors.New("webhook processing in flight")

	// ErrDuplicateWebhook retained for backward compat — deprecated. Old
	// callers using errors.Is(err, ErrDuplicateWebhook) will keep working
	// since we return ErrAlreadyComplete for the finished case.
	ErrDuplicateWebhook = ErrAlreadyComplete
)

const (
	stateInProgress = "in_progress"
	stateComplete   = "complete"

	// InFlightTTL is how long a Claim holds the in-progress lock before
	// auto-releasing. Long enough to cover a slow DB write + retry, short
	// enough that a crashed worker doesn't strand the event for hours.
	InFlightTTL = 5 * time.Minute

	// CompleteTTL is how long we remember successfully-processed events.
	// Providers typically retry for 24-72h on failure; matching the top of
	// that range keeps us safe against very late retries.
	CompleteTTL = 24 * time.Hour
)

// IdempotencyCache tracks webhook processing state in Redis. Safe against
// concurrent handlers across multiple API instances. With no Redis (nil
// client) all methods no-op — fine for local dev where webhooks aren't
// reaching the machine anyway.
type IdempotencyCache struct {
	client *redis.Client
}

func NewIdempotencyCache(client *redis.Client) *IdempotencyCache {
	return &IdempotencyCache{client: client}
}

// Claim attempts to lock an event for processing. Returns:
//   - nil                     → caller owns the lock, must call Complete on success
//   - ErrAlreadyComplete      → event already processed, ack as duplicate
//   - ErrProcessingInFlight   → another worker owns the lock, retry later
//   - other error             → Redis error, treated as fail-open (returns nil)
//
// Provider must be a short tag ("polar" / "paystack") so different providers
// can't collide on the same event ID.
func (c *IdempotencyCache) Claim(ctx context.Context, provider, eventID string) error {
	if c == nil || c.client == nil || eventID == "" {
		return nil
	}
	key := c.key(provider, eventID)

	// SETNX with the in-progress marker + short TTL. If it succeeds, we
	// own the lock.
	ok, err := c.client.SetNX(ctx, key, stateInProgress, InFlightTTL).Result()
	if err != nil {
		// Fail open — see failure semantics comment on the package.
		return nil
	}
	if ok {
		return nil
	}

	// Key already exists — inspect its state to decide the outcome.
	val, gerr := c.client.Get(ctx, key).Result()
	if gerr != nil {
		// Race: key was there when SETNX ran, gone now (TTL expired
		// between the two commands). Retry the SETNX once so the caller
		// gets the lock in the common case; if that also loses, fall
		// through to ErrProcessingInFlight to be safe.
		if ok2, _ := c.client.SetNX(ctx, key, stateInProgress, InFlightTTL).Result(); ok2 {
			return nil
		}
		return ErrProcessingInFlight
	}
	if val == stateComplete {
		return ErrAlreadyComplete
	}
	return ErrProcessingInFlight
}

// Complete promotes a claimed lock to permanent complete. Call this only
// after the payment / plan-upgrade side effects have committed. It's safe
// (idempotent) to call more than once for the same event ID.
func (c *IdempotencyCache) Complete(ctx context.Context, provider, eventID string) error {
	if c == nil || c.client == nil || eventID == "" {
		return nil
	}
	key := c.key(provider, eventID)
	if err := c.client.Set(ctx, key, stateComplete, CompleteTTL).Err(); err != nil {
		// Non-fatal: the worst case is that a very late retry (>5m) gets
		// a fresh Claim and re-processes. Service-layer idempotency
		// prevents double activation. Log the error via caller — we don't
		// carry a logger here to keep the type dependency-free.
		return err
	}
	return nil
}

// Release drops the in-progress lock without marking complete. Call this
// when processing fails permanently (e.g. malformed payload we can't retry)
// so a subsequent legitimate delivery of the same event ID can proceed
// immediately instead of waiting the 5m TTL. Rarely needed — the TTL
// handles the crash case; Release is for known-fatal outcomes.
func (c *IdempotencyCache) Release(ctx context.Context, provider, eventID string) error {
	if c == nil || c.client == nil || eventID == "" {
		return nil
	}
	// Only delete if we still hold the in-progress lock. Prevents wiping
	// a complete marker that raced ahead of us.
	key := c.key(provider, eventID)
	script := redis.NewScript(`
		if redis.call('GET', KEYS[1]) == ARGV[1] then
			return redis.call('DEL', KEYS[1])
		end
		return 0
	`)
	_, _ = script.Run(ctx, c.client, []string{key}, stateInProgress).Result()
	return nil
}

func (c *IdempotencyCache) key(provider, eventID string) string {
	return "webhook:seen:" + provider + ":" + eventID
}
