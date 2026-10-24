package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PasswordResetRecord is the internal shape of a password reset token row.
// The plaintext token never touches the DB — only its SHA-256 hex digest.
type PasswordResetRecord struct {
	ID        string
	UserID    string
	TokenHash string
	ExpiresAt time.Time
	UsedAt    *time.Time
	CreatedAt time.Time
}

type PasswordResetPostgres struct {
	pool *pgxpool.Pool
}

func NewPasswordResetPostgres(pool *pgxpool.Pool) *PasswordResetPostgres {
	return &PasswordResetPostgres{pool: pool}
}

// Create inserts a new pending reset row. The caller is responsible for
// generating the token + hash and keeping them in sync.
func (r *PasswordResetPostgres) Create(ctx context.Context, rec *PasswordResetRecord) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO password_resets (id, user_id, token_hash, expires_at, created_at)
		 VALUES ($1, $2, $3, $4, $5)`,
		rec.ID, rec.UserID, rec.TokenHash, rec.ExpiresAt, rec.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("creating password reset: %w", err)
	}
	return nil
}

// GetByTokenHash looks up a reset row by its SHA-256 hex digest.
// Returns (nil, nil) when no match — callers should treat that as "invalid".
func (r *PasswordResetPostgres) GetByTokenHash(ctx context.Context, hash string) (*PasswordResetRecord, error) {
	var rec PasswordResetRecord
	err := r.pool.QueryRow(ctx,
		`SELECT id, user_id, token_hash, expires_at, used_at, created_at
		 FROM password_resets WHERE token_hash = $1`,
		hash,
	).Scan(&rec.ID, &rec.UserID, &rec.TokenHash, &rec.ExpiresAt, &rec.UsedAt, &rec.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("querying password reset: %w", err)
	}
	return &rec, nil
}

// MarkUsed stamps used_at on a token row — reset tokens are single-use.
// Retained for backward compat with existing callers; new code should use
// ClaimByTokenHash for the atomic path.
func (r *PasswordResetPostgres) MarkUsed(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE password_resets SET used_at = NOW() WHERE id = $1`,
		id,
	)
	if err != nil {
		return fmt.Errorf("marking reset used: %w", err)
	}
	return nil
}

// ClaimByTokenHash atomically consumes a reset token in a single statement
// and returns the associated user_id. Only succeeds if the token is unused
// AND unexpired. Two concurrent reset requests for the same token can't
// both succeed — the UPDATE...WHERE...RETURNING is atomic per row and the
// second attempt gets sql.ErrNoRows.
//
// Audit finding — the previous flow (change password, then mark token used
// in a separate statement) allowed the same token to change the password
// multiple times under concurrent requests or if MarkUsed failed.
func (r *PasswordResetPostgres) ClaimByTokenHash(ctx context.Context, hash string) (string, error) {
	var userID string
	err := r.pool.QueryRow(ctx,
		`UPDATE password_resets
		 SET used_at = NOW()
		 WHERE token_hash = $1
		   AND used_at IS NULL
		   AND expires_at > NOW()
		 RETURNING user_id`,
		hash,
	).Scan(&userID)
	if errors.Is(err, pgx.ErrNoRows) {
		// Token missing, already used, or expired — callers treat as "invalid".
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("claiming password reset: %w", err)
	}
	return userID, nil
}

// RecentCountForUser returns how many reset tokens a user has minted within the
// given window. Used to throttle abuse.
func (r *PasswordResetPostgres) RecentCountForUser(ctx context.Context, userID string, within time.Duration) (int, error) {
	var n int
	err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM password_resets
		 WHERE user_id = $1 AND created_at > NOW() - $2::interval`,
		userID, fmt.Sprintf("%d seconds", int(within.Seconds())),
	).Scan(&n)
	if err != nil {
		return 0, fmt.Errorf("counting recent resets: %w", err)
	}
	return n, nil
}
