package database

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"sort"

	"github.com/jackc/pgx/v5/pgxpool"
)

// schemaSQL bootstraps the application schema for fresh or reused databases.
//
//go:embed schema.sql
var schemaSQL string

// runtimeMigrations contains append-only changes created after the tracked
// schema baseline. Historical files in migrations/ document the early schema;
// they predate migration tracking and are deliberately not replayed.
//
//go:embed migrations/runtime/*.sql
var runtimeMigrations embed.FS

func EnsureSchema(ctx context.Context, pool *pgxpool.Pool) error {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("beginning schema migration: %w", err)
	}
	defer tx.Rollback(ctx) // no-op after Commit

	// Only one application instance may inspect or apply migrations at once.
	if _, err := tx.Exec(ctx, `SELECT pg_advisory_xact_lock(43559721)`); err != nil {
		return fmt.Errorf("locking schema migration: %w", err)
	}
	if _, err := tx.Exec(ctx, `CREATE TABLE IF NOT EXISTS schema_migrations (
		version TEXT PRIMARY KEY,
		applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
	)`); err != nil {
		return fmt.Errorf("creating schema migration ledger: %w", err)
	}

	var baselineApplied bool
	err = tx.QueryRow(ctx,
		`SELECT EXISTS (SELECT 1 FROM schema_migrations WHERE version = 'baseline-2026-09-11')`,
	).Scan(&baselineApplied)
	if err != nil {
		return fmt.Errorf("checking schema baseline: %w", err)
	}
	if !baselineApplied {
		if _, err := tx.Exec(ctx, schemaSQL); err != nil {
			return fmt.Errorf("applying schema baseline: %w", err)
		}
		if _, err := tx.Exec(ctx, `INSERT INTO schema_migrations (version) VALUES ('baseline-2026-09-11')`); err != nil {
			return fmt.Errorf("recording schema baseline: %w", err)
		}
	}

	entries, err := fs.ReadDir(runtimeMigrations, "migrations/runtime")
	if err != nil {
		return fmt.Errorf("reading runtime migrations: %w", err)
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].Name() < entries[j].Name() })
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		version := entry.Name()
		var applied bool
		if err := tx.QueryRow(ctx,
			`SELECT EXISTS (SELECT 1 FROM schema_migrations WHERE version = $1)`, version,
		).Scan(&applied); err != nil {
			return fmt.Errorf("checking migration %s: %w", version, err)
		}
		if applied {
			continue
		}
		sql, err := fs.ReadFile(runtimeMigrations, "migrations/runtime/"+version)
		if err != nil {
			return fmt.Errorf("reading migration %s: %w", version, err)
		}
		if _, err := tx.Exec(ctx, string(sql)); err != nil {
			return fmt.Errorf("applying migration %s: %w", version, err)
		}
		if _, err := tx.Exec(ctx, `INSERT INTO schema_migrations (version) VALUES ($1)`, version); err != nil {
			return fmt.Errorf("recording migration %s: %w", version, err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("committing schema migrations: %w", err)
	}
	return nil
}
