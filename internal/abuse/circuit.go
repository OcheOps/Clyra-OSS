package abuse

import (
	"errors"
	"sync"
	"time"
)

var ErrCircuitOpen = errors.New("ai_temporarily_unavailable")

// CircuitState represents the current state of the circuit breaker.
type CircuitState int

const (
	CircuitClosed   CircuitState = iota // Normal — requests pass through
	CircuitOpen                         // Tripped — all requests rejected
	CircuitHalfOpen                     // Testing — one request allowed to probe
)

// CircuitBreaker protects against cascading failures from unhealthy providers.
// It tracks consecutive failures and trips open when the threshold is exceeded.
// After a cooldown period, it transitions to half-open and allows a single probe request.
type CircuitBreaker struct {
	mu              sync.Mutex
	state           CircuitState
	failures        int
	threshold       int
	cooldown        time.Duration
	lastFailureTime time.Time
}

func NewCircuitBreaker(threshold int, cooldown time.Duration) *CircuitBreaker {
	if threshold <= 0 {
		threshold = 5
	}
	if cooldown <= 0 {
		cooldown = 30 * time.Second
	}
	return &CircuitBreaker{
		state:     CircuitClosed,
		threshold: threshold,
		cooldown:  cooldown,
	}
}

// Allow checks if a request should be permitted through the circuit.
// Returns true if the request can proceed, false if the circuit is open.
func (cb *CircuitBreaker) Allow() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case CircuitClosed:
		return true

	case CircuitOpen:
		// Check if cooldown has elapsed
		if time.Since(cb.lastFailureTime) >= cb.cooldown {
			cb.state = CircuitHalfOpen
			return true // allow one probe request
		}
		return false

	case CircuitHalfOpen:
		// Only one request at a time in half-open — block others
		return false

	default:
		return true
	}
}

// RecordSuccess signals that a request completed successfully.
// Resets the circuit to closed state.
func (cb *CircuitBreaker) RecordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.failures = 0
	cb.state = CircuitClosed
}

// RecordFailure signals that a request failed.
// If failures exceed the threshold, the circuit trips open.
func (cb *CircuitBreaker) RecordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.failures++
	cb.lastFailureTime = time.Now()

	if cb.failures >= cb.threshold {
		cb.state = CircuitOpen
	}
}

// State returns the current circuit state (for observability).
func (cb *CircuitBreaker) State() CircuitState {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	return cb.state
}

// Failures returns the current failure count (for observability).
func (cb *CircuitBreaker) Failures() int {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	return cb.failures
}

func (s CircuitState) String() string {
	switch s {
	case CircuitClosed:
		return "closed"
	case CircuitOpen:
		return "open"
	case CircuitHalfOpen:
		return "half-open"
	default:
		return "unknown"
	}
}
