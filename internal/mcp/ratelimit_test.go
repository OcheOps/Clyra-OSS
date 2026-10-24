package mcp

import (
	"context"
	"testing"
)

func TestDailyLimiterAllowsUpToCap(t *testing.T) {
	l := NewMemoryDailyLimiter(3)
	for i := 1; i <= 3; i++ {
		ok, remaining, _ := l.Allow(context.Background(), "u1")
		if !ok || remaining != 3-i {
			t.Fatalf("call %d: ok=%v remaining=%d", i, ok, remaining)
		}
	}
	if ok, _, _ := l.Allow(context.Background(), "u1"); ok {
		t.Fatal("fourth call should be blocked")
	}
}

func TestDailyLimiterCapsEveryManagedCaller(t *testing.T) {
	l := NewMemoryDailyLimiter(1)
	if ok, _, _ := l.Allow(context.Background(), "byo1"); !ok {
		t.Fatal("first managed call blocked")
	}
	if ok, _, _ := l.Allow(context.Background(), "byo1"); ok {
		t.Fatal("second managed call should be blocked regardless of account plan")
	}
}

func TestDailyLimiterPerUserIsolation(t *testing.T) {
	l := NewMemoryDailyLimiter(1)
	if ok, _, _ := l.Allow(context.Background(), "A"); !ok {
		t.Fatal("A first call blocked")
	}
	if ok, _, _ := l.Allow(context.Background(), "A"); ok {
		t.Fatal("A second call should be blocked")
	}
	if ok, _, _ := l.Allow(context.Background(), "B"); !ok {
		t.Fatal("B should not be blocked by A")
	}
}

func TestDailyLimiterUnlimitedWhenZero(t *testing.T) {
	l := NewMemoryDailyLimiter(0)
	for i := range 50 {
		if ok, _, _ := l.Allow(context.Background(), "u"); !ok {
			t.Fatalf("call %d blocked", i)
		}
	}
}

func TestDailyLimiterRejectsEmptyUser(t *testing.T) {
	l := NewMemoryDailyLimiter(10)
	if ok, _, _ := l.Allow(context.Background(), ""); ok {
		t.Fatal("empty user should be rejected")
	}
}
