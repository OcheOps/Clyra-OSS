package payments

import (
	"context"
	"log/slog"
	"time"
)

// StaleCleaner flips abandoned pending payments to failed so they stop
// blocking new checkouts. The service prevents a user from initializing a
// new payment while one is still pending for the same plan — without this
// sweep a user who bails out of Polar checkout would stay wedged until
// someone ran SQL.
//
// Defaults match the backlog spec: threshold 2h, tick 10m. Returns a stop
// function so tests / graceful shutdown can halt the loop.
func StartStaleCleaner(
	ctx context.Context,
	repo *Repository,
	logger *slog.Logger,
	threshold time.Duration,
	tick time.Duration,
) func() {
	loopCtx, cancel := context.WithCancel(ctx)

	go func() {
		// Run once at boot so a freshly-deployed server doesn't wait a
		// full tick to reconcile leftovers from the previous deployment.
		runSweep(loopCtx, repo, logger, threshold)

		t := time.NewTicker(tick)
		defer t.Stop()

		for {
			select {
			case <-loopCtx.Done():
				return
			case <-t.C:
				runSweep(loopCtx, repo, logger, threshold)
			}
		}
	}()

	return cancel
}

func runSweep(ctx context.Context, repo *Repository, logger *slog.Logger, threshold time.Duration) {
	sweepCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	n, err := repo.ExpireStalePending(sweepCtx, threshold)
	if err != nil {
		logger.Error("payments cleanup sweep failed", "error", err)
		return
	}
	if n > 0 {
		logger.Info("payments cleanup swept stale pending rows", "count", n, "threshold", threshold.String())
	}
}
