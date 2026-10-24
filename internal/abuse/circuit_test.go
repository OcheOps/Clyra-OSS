package abuse

import (
	"testing"
	"time"
)

func TestCircuitBreaker_StartsClosedAndAllows(t *testing.T) {
	cb := NewCircuitBreaker(3, 5*time.Second)

	if cb.State() != CircuitClosed {
		t.Fatalf("expected closed, got %s", cb.State())
	}
	if !cb.Allow() {
		t.Fatal("closed circuit should allow requests")
	}
}

func TestCircuitBreaker_TripsAfterThreshold(t *testing.T) {
	cb := NewCircuitBreaker(3, 5*time.Second)

	cb.RecordFailure()
	cb.RecordFailure()

	// Still closed at 2 failures (threshold is 3)
	if cb.State() != CircuitClosed {
		t.Fatalf("expected closed after 2 failures, got %s", cb.State())
	}
	if !cb.Allow() {
		t.Fatal("should still allow at 2 failures")
	}

	cb.RecordFailure() // 3rd failure — trips

	if cb.State() != CircuitOpen {
		t.Fatalf("expected open after 3 failures, got %s", cb.State())
	}
	if cb.Allow() {
		t.Fatal("open circuit should block requests")
	}
}

func TestCircuitBreaker_ResetOnSuccess(t *testing.T) {
	cb := NewCircuitBreaker(3, 5*time.Second)

	cb.RecordFailure()
	cb.RecordFailure()
	cb.RecordSuccess() // reset

	if cb.Failures() != 0 {
		t.Fatalf("expected 0 failures after success, got %d", cb.Failures())
	}
	if cb.State() != CircuitClosed {
		t.Fatalf("expected closed after success, got %s", cb.State())
	}
}

func TestCircuitBreaker_HalfOpenAfterCooldown(t *testing.T) {
	// Use a very short cooldown for testing
	cb := NewCircuitBreaker(2, 50*time.Millisecond)

	cb.RecordFailure()
	cb.RecordFailure() // trips open

	if cb.State() != CircuitOpen {
		t.Fatalf("expected open, got %s", cb.State())
	}
	if cb.Allow() {
		t.Fatal("should block while open")
	}

	// Wait for cooldown
	time.Sleep(60 * time.Millisecond)

	// First call transitions to half-open and allows
	if !cb.Allow() {
		t.Fatal("should allow one probe request after cooldown")
	}
	if cb.State() != CircuitHalfOpen {
		t.Fatalf("expected half-open, got %s", cb.State())
	}

	// Second call while half-open should block
	if cb.Allow() {
		t.Fatal("should block second request in half-open")
	}
}

func TestCircuitBreaker_HalfOpenSuccessCloses(t *testing.T) {
	cb := NewCircuitBreaker(2, 50*time.Millisecond)

	cb.RecordFailure()
	cb.RecordFailure()
	time.Sleep(60 * time.Millisecond)

	cb.Allow() // transitions to half-open
	cb.RecordSuccess() // probe succeeded

	if cb.State() != CircuitClosed {
		t.Fatalf("expected closed after half-open success, got %s", cb.State())
	}
	if !cb.Allow() {
		t.Fatal("should allow after recovery")
	}
}

func TestCircuitBreaker_HalfOpenFailureReopens(t *testing.T) {
	cb := NewCircuitBreaker(2, 50*time.Millisecond)

	cb.RecordFailure()
	cb.RecordFailure()
	time.Sleep(60 * time.Millisecond)

	cb.Allow() // transitions to half-open
	cb.RecordFailure() // probe failed

	if cb.State() != CircuitOpen {
		t.Fatalf("expected open after half-open failure, got %s", cb.State())
	}
}

func TestCircuitBreaker_StateString(t *testing.T) {
	tests := []struct {
		state CircuitState
		want  string
	}{
		{CircuitClosed, "closed"},
		{CircuitOpen, "open"},
		{CircuitHalfOpen, "half-open"},
		{CircuitState(99), "unknown"},
	}
	for _, tt := range tests {
		if got := tt.state.String(); got != tt.want {
			t.Errorf("State(%d).String() = %q, want %q", tt.state, got, tt.want)
		}
	}
}
