package apitoken

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository persists API tokens. All lookups filter revoked_at IS NULL
// so a revoked token can never authenticate a request even if the caller
// somehow guessed / kept its hash.
type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

const cols = `id, user_id, name, source, last_used, created_at, revoked_at`

func scanRow(row pgx.Row, t *Token) error {
	var source string
	if err := row.Scan(&t.ID, &t.UserID, &t.Name, &source, &t.LastUsed, &t.CreatedAt, &t.RevokedAt); err != nil {
		return err
	}
	t.Source = Source(source)
	return nil
}

// Create inserts a new token row. Caller supplies the id, name, and
// pre-hashed value — this repo never sees the cleartext.
func (r *Repository) Create(ctx context.Context, id, userID, name, hash string, src Source) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO api_tokens (id, user_id, name, token_hash, source)
		 VALUES ($1, $2, $3, $4, $5)`,
		id, userID, name, hash, string(src),
	)
	if err != nil {
		return fmt.Errorf("inserting api token: %w", err)
	}
	return nil
}

// GetByHash finds a live token by its sha256 hex. Returns (nil, nil)
// when the token doesn't exist or has been revoked — callers treat
// both as "unauthorized" and shouldn't distinguish.
func (r *Repository) GetByHash(ctx context.Context, hash string) (*Token, error) {
	var t Token
	err := scanRow(
		r.pool.QueryRow(ctx,
			`SELECT `+cols+` FROM api_tokens
			 WHERE token_hash = $1 AND revoked_at IS NULL`,
			hash,
		), &t,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get api token by hash: %w", err)
	}
	return &t, nil
}

// ListByUser returns every token owned by a user (live + revoked), newest
// first. UI shows all so users can see historical revocations.
func (r *Repository) ListByUser(ctx context.Context, userID string) ([]Token, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT `+cols+` FROM api_tokens
		 WHERE user_id = $1
		 ORDER BY created_at DESC
		 LIMIT 100`,
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("list api tokens: %w", err)
	}
	defer rows.Close()
	var out []Token
	for rows.Next() {
		var t Token
		if err := scanRow(rows, &t); err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, nil
}

// Revoke marks a token revoked. Idempotent: revoking twice is a no-op.
// Only touches rows owned by userID so a compromised token can't be
// used to revoke someone else's key.
func (r *Repository) Revoke(ctx context.Context, id, userID string) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE api_tokens SET revoked_at = NOW()
		 WHERE id = $1 AND user_id = $2 AND revoked_at IS NULL`,
		id, userID,
	)
	if err != nil {
		return fmt.Errorf("revoke api token: %w", err)
	}
	return nil
}

// TouchLastUsed stamps last_used = NOW(). Called async from the auth
// middleware — a failure just means the UI shows stale freshness data,
// nothing security-critical.
func (r *Repository) TouchLastUsed(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE api_tokens SET last_used = NOW() WHERE id = $1`,
		id,
	)
	if err != nil {
		return fmt.Errorf("touch last used: %w", err)
	}
	return nil
}

// CountLiveByUser returns how many live tokens a user owns. Used to cap
// per-user token count so nobody spins up 10k tokens accidentally.
func (r *Repository) CountLiveByUser(ctx context.Context, userID string) (int, error) {
	var n int
	err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM api_tokens
		 WHERE user_id = $1 AND revoked_at IS NULL`,
		userID,
	).Scan(&n)
	if err != nil {
		return 0, err
	}
	return n, nil
}

// EnsureSchema exists for tests that spin up a pool without running the
// full schema.sql. No-op in production where the schema is already applied.
func (r *Repository) EnsureSchema(ctx context.Context) error {
	_, _ = r.pool.Exec(ctx, `CREATE TABLE IF NOT EXISTS api_tokens (
		id TEXT PRIMARY KEY,
		user_id TEXT NOT NULL,
		name TEXT NOT NULL,
		token_hash TEXT NOT NULL UNIQUE,
		source TEXT NOT NULL DEFAULT 'pat',
		last_used TIMESTAMPTZ,
		created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		revoked_at TIMESTAMPTZ
	)`)
	_ = time.Now() // silence unused import if we drop the schema line
	return nil
}
