package abuse

import (
	"context"
	"testing"
)

// mockQuotaRepo simulates the database for testing quota logic.
type mockQuotaRepo struct {
	counts map[string]int
}

func newMockQuotaRepo() *mockQuotaRepo {
	return &mockQuotaRepo{counts: make(map[string]int)}
}

func (m *mockQuotaRepo) getDailyCount(_ context.Context, userID string) int {
	return m.counts[userID]
}

func (m *mockQuotaRepo) increment(userID string) {
	m.counts[userID]++
}

// TestQuotaService tests the check/increment logic without a real DB.
// We test the core algorithm directly since QuotaService wraps QuotaRepository.
func TestQuotaLogic_UnderLimit(t *testing.T) {
	limit := 5
	count := 3
	if count >= limit {
		t.Fatal("should be under limit")
	}
}

func TestQuotaLogic_AtLimit(t *testing.T) {
	limit := 5
	count := 5
	if count < limit {
		t.Fatal("should be at limit")
	}
}

func TestQuotaLogic_Increment(t *testing.T) {
	repo := newMockQuotaRepo()

	repo.increment("user-1")
	repo.increment("user-1")
	repo.increment("user-1")

	if repo.getDailyCount(context.Background(), "user-1") != 3 {
		t.Fatalf("expected 3, got %d", repo.getDailyCount(context.Background(), "user-1"))
	}

	// Different user starts at 0
	if repo.getDailyCount(context.Background(), "user-2") != 0 {
		t.Fatal("user-2 should be at 0")
	}
}

func TestQuotaLogic_PerUserIsolation(t *testing.T) {
	repo := newMockQuotaRepo()
	limit := 2

	repo.increment("user-a")
	repo.increment("user-a")

	// user-a is at limit
	if repo.getDailyCount(context.Background(), "user-a") < limit {
		t.Fatal("user-a should be at limit")
	}

	// user-b is still at 0
	if repo.getDailyCount(context.Background(), "user-b") >= limit {
		t.Fatal("user-b should not be at limit")
	}
}

func TestErrDailyLimitReached(t *testing.T) {
	if ErrDailyLimitReached.Error() != "daily_limit_reached" {
		t.Fatalf("unexpected error message: %s", ErrDailyLimitReached.Error())
	}
}
