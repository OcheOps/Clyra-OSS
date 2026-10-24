package abuse

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrDailyLimitReached = errors.New("daily_limit_reached")
)

// QuotaRepository handles daily usage count persistence.
type QuotaRepository struct {
	pool *pgxpool.Pool
}

func NewQuotaRepository(pool *pgxpool.Pool) *QuotaRepository {
	return &QuotaRepository{pool: pool}
}

// GetDailyCount returns the number of requests a user has made today.
func (r *QuotaRepository) GetDailyCount(ctx context.Context, userID string) (int, error) {
	if r.pool == nil {
		return 0, nil // no DB — fail open
	}
	var count int
	err := r.pool.QueryRow(ctx,
		`SELECT COALESCE(request_count, 0) FROM usage_records
		 WHERE user_id = $1 AND usage_date = CURRENT_DATE`,
		userID,
	).Scan(&count)

	if err != nil {
		// No row means 0 usage today
		if err.Error() == "no rows in result set" {
			return 0, nil
		}
		return 0, fmt.Errorf("querying daily count: %w", err)
	}
	return count, nil
}

// IncrementDailyCount atomically increments the user's request count for today.
// Uses UPSERT to handle the first request of the day.
func (r *QuotaRepository) IncrementDailyCount(ctx context.Context, userID string) error {
	if r.pool == nil {
		return nil
	}
	_, err := r.pool.Exec(ctx,
		`INSERT INTO usage_records (user_id, usage_date, request_count)
		 VALUES ($1, CURRENT_DATE, 1)
		 ON CONFLICT (user_id, usage_date)
		 DO UPDATE SET request_count = usage_records.request_count + 1`,
		userID,
	)
	if err != nil {
		return fmt.Errorf("incrementing daily count: %w", err)
	}
	return nil
}

// QuotaService checks and enforces daily usage limits.
// Limits are resolved per-plan via planLimits, with fallbackLimit used when
// the plan is empty or unknown.
type QuotaService struct {
	repo          *QuotaRepository
	fallbackLimit int
	planLimits    map[string]int
}

// NewQuotaService constructs a plan-aware quota service.
// planLimits may be nil — in that case all plans fall back to fallbackLimit.
func NewQuotaService(repo *QuotaRepository, fallbackLimit int, planLimits map[string]int) *QuotaService {
	return &QuotaService{
		repo:          repo,
		fallbackLimit: fallbackLimit,
		planLimits:    planLimits,
	}
}

// LimitFor returns the daily limit for the given plan.
func (s *QuotaService) LimitFor(plan string) int {
	if s.planLimits != nil {
		if v, ok := s.planLimits[plan]; ok {
			return v
		}
	}
	return s.fallbackLimit
}

// Check returns an error if the user has exceeded their plan's daily limit.
func (s *QuotaService) Check(ctx context.Context, userID, plan string) error {
	count, err := s.repo.GetDailyCount(ctx, userID)
	if err != nil {
		// Fail open — don't block users if DB is having issues.
		return nil
	}
	if count >= s.LimitFor(plan) {
		return ErrDailyLimitReached
	}
	return nil
}

// Increment records a successful request against the user's daily quota.
func (s *QuotaService) Increment(ctx context.Context, userID string) error {
	return s.repo.IncrementDailyCount(ctx, userID)
}

// Remaining returns how many requests the user has left today given their plan.
func (s *QuotaService) Remaining(ctx context.Context, userID, plan string) (int, error) {
	limit := s.LimitFor(plan)
	count, err := s.repo.GetDailyCount(ctx, userID)
	if err != nil {
		return limit, nil // fail open
	}
	remaining := limit - count
	if remaining < 0 {
		remaining = 0
	}
	return remaining, nil
}

// UsageLogRepository logs individual requests for observability.
type UsageLogRepository struct {
	pool *pgxpool.Pool
}

func NewUsageLogRepository(pool *pgxpool.Pool) *UsageLogRepository {
	return &UsageLogRepository{pool: pool}
}

// Log records a generation request. Does NOT store prompts or API keys.
func (r *UsageLogRepository) Log(ctx context.Context, userID, provider string, success bool, latency time.Duration) {
	if r.pool == nil {
		return
	}
	// Fire and forget — don't block the response on logging
	_, _ = r.pool.Exec(ctx,
		`INSERT INTO usage_log (user_id, provider, success, latency_ms)
		 VALUES ($1, $2, $3, $4)`,
		userID, provider, success, latency.Milliseconds(),
	)
}
