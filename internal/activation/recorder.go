// Package activation records coarse funnel markers. Each (user_id, kind)
// is unique — the first time a user does a thing is what we care about.
// Subsequent calls are silent no-ops (ON CONFLICT DO NOTHING).
//
// This is intentionally minimal: no metadata, no PostHog, no analytics
// queue. A dashboard query is just `SELECT kind, COUNT(*) FROM
// activation_events GROUP BY kind`. Upgrade when we outgrow it.
package activation

import (
	"context"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Kind string

const (
	KindSignup         Kind = "signup"
	KindFirstDiagram   Kind = "first_diagram"
	KindFirstIaCExport Kind = "first_iac_export"
)

type Recorder struct {
	pool   *pgxpool.Pool
	logger *slog.Logger
}

func NewRecorder(pool *pgxpool.Pool, logger *slog.Logger) *Recorder {
	return &Recorder{pool: pool, logger: logger}
}

// Record inserts a funnel marker for (userID, kind). Repeated calls are
// no-ops thanks to the PK + ON CONFLICT clause. Errors are logged and
// swallowed — activation tracking must never break a user-facing path.
func (r *Recorder) Record(ctx context.Context, userID string, kind Kind) {
	if r == nil || r.pool == nil || userID == "" {
		return
	}
	_, err := r.pool.Exec(ctx,
		`INSERT INTO activation_events (user_id, kind)
		 VALUES ($1, $2)
		 ON CONFLICT (user_id, kind) DO NOTHING`,
		userID, string(kind),
	)
	if err != nil && r.logger != nil {
		r.logger.Warn("activation record failed", "user_id", userID, "kind", kind, "error", err)
	}
}

// HasKind returns true if (userID, kind) has ever been recorded. Used to
// gate lifetime-one freebies (e.g. the free plan's single IaC export).
// A query error is treated as "not present" so the caller fails open —
// we'd rather grant an extra export than wrongly lock a paying flow.
func (r *Recorder) HasKind(ctx context.Context, userID string, kind Kind) bool {
	if r == nil || r.pool == nil || userID == "" {
		return false
	}
	var exists bool
	err := r.pool.QueryRow(ctx,
		`SELECT EXISTS (
		   SELECT 1 FROM activation_events WHERE user_id = $1 AND kind = $2
		 )`,
		userID, string(kind),
	).Scan(&exists)
	if err != nil {
		if r.logger != nil {
			r.logger.Warn("activation has-kind failed", "user_id", userID, "kind", kind, "error", err)
		}
		return false
	}
	return exists
}
