package abuse

import (
	"context"
	"log/slog"
	"time"

	"github.com/oche/diagram/internal/config"
)

// Guard is the central abuse prevention facade injected into the diagram service.
// It coordinates all checks: circuit breaker, token clamping, node limits, usage logging.
type Guard struct {
	CircuitBreaker  *CircuitBreaker
	TokenGuard      *TokenGuard
	NodeGuard       *NodeGuard
	PromptValidator *PromptValidator
	QuotaService    *QuotaService
	UsageLog        *UsageLogRepository
	ProviderTimeout time.Duration
	Logger          *slog.Logger
}

func NewGuard(cfg config.AbuseConfig, quotaService *QuotaService, usageLog *UsageLogRepository, logger *slog.Logger) *Guard {
	return &Guard{
		CircuitBreaker:  NewCircuitBreaker(cfg.CircuitBreakerThreshold, cfg.CircuitBreakerCooldown),
		TokenGuard:      NewTokenGuard(cfg),
		NodeGuard:       NewNodeGuard(cfg),
		PromptValidator: NewPromptValidator(cfg),
		QuotaService:    quotaService,
		UsageLog:        usageLog,
		ProviderTimeout: cfg.ProviderTimeout,
		Logger:          logger,
	}
}

// PaidPlanProviderTimeout is the per-request LLM ceiling for Pro/BYO users.
// Paid users routinely generate denser diagrams (raised token cap → longer
// streaming), so they need more wall-clock than the free baseline.
const PaidPlanProviderTimeout = 180 * time.Second

// TimeoutForPlan returns the provider timeout for a given plan. Free users
// stay on the configured ProviderTimeout (operator tuning knob); paid plans
// get whichever is larger of the configured value or PaidPlanProviderTimeout
// so an operator who raises the env baseline isn't silently capped.
func (g *Guard) TimeoutForPlan(plan string) time.Duration {
	if g == nil {
		return 0
	}
	if isPaidPlan(plan) && PaidPlanProviderTimeout > g.ProviderTimeout {
		return PaidPlanProviderTimeout
	}
	return g.ProviderTimeout
}

// PreCheck runs all pre-request validations.
// Call this before making the provider call. The plan determines the daily limit.
func (g *Guard) PreCheck(ctx context.Context, userID, plan, prompt string) (string, error) {
	// 1. Circuit breaker
	if !g.CircuitBreaker.Allow() {
		return "", ErrCircuitOpen
	}

	// 2. Validate and sanitize prompt
	cleaned, err := g.PromptValidator.Validate(prompt, plan)
	if err != nil {
		return "", err
	}

	// 3. Check daily quota — plan-aware.
	if err := g.QuotaService.Check(ctx, userID, plan); err != nil {
		return "", err
	}

	return cleaned, nil
}

// PostSuccess runs after a successful provider response.
func (g *Guard) PostSuccess(ctx context.Context, userID, providerName string, nodeCount int, latency time.Duration) error {
	// Reset circuit breaker
	g.CircuitBreaker.RecordSuccess()

	// Log usage
	if g.UsageLog != nil {
		g.UsageLog.Log(ctx, userID, providerName, true, latency)
	}

	// Increment daily quota — always, since the limit itself reflects plan/cost policy.
	if err := g.QuotaService.Increment(ctx, userID); err != nil {
		g.Logger.Error("failed to increment quota", "user_id", userID, "error", err)
	}

	// Check node count
	if err := g.NodeGuard.Check(nodeCount); err != nil {
		g.Logger.Warn("diagram exceeds node limit",
			"user_id", userID,
			"provider", providerName,
			"nodes", nodeCount,
		)
		// Don't reject — just log. The diagram is already generated.
	}

	return nil
}

// PostFailure runs after a failed provider call.
func (g *Guard) PostFailure(ctx context.Context, userID, providerName string, latency time.Duration) {
	g.CircuitBreaker.RecordFailure()

	if g.UsageLog != nil {
		g.UsageLog.Log(ctx, userID, providerName, false, latency)
	}
}
