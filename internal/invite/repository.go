package invite

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) Create(ctx context.Context, createdBy string, expiresAt *time.Time, plan string) (*InviteCode, error) {
	id := generateID()
	code := generateCode()

	if plan == "" {
		plan = "free"
	}

	_, err := r.pool.Exec(ctx,
		`INSERT INTO invite_codes (id, code, created_by, expires_at, plan)
		 VALUES ($1, $2, $3, $4, $5)`,
		id, code, createdBy, expiresAt, plan,
	)
	if err != nil {
		return nil, fmt.Errorf("creating invite: %w", err)
	}

	return &InviteCode{
		ID:        id,
		Code:      code,
		CreatedBy: createdBy,
		ExpiresAt: expiresAt,
		Plan:      plan,
		CreatedAt: time.Now().UTC(),
	}, nil
}

func (r *Repository) GetByCode(ctx context.Context, code string) (*InviteCode, error) {
	var inv InviteCode
	err := r.pool.QueryRow(ctx,
		`SELECT id, code, created_by, used_by, used_at, expires_at, plan, created_at
		 FROM invite_codes WHERE code = $1`,
		code,
	).Scan(&inv.ID, &inv.Code, &inv.CreatedBy, &inv.UsedBy, &inv.UsedAt, &inv.ExpiresAt, &inv.Plan, &inv.CreatedAt)

	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("querying invite: %w", err)
	}
	return &inv, nil
}

func (r *Repository) MarkUsed(ctx context.Context, code string, userID string) error {
	now := time.Now().UTC()
	_, err := r.pool.Exec(ctx,
		`UPDATE invite_codes SET used_by = $1, used_at = $2 WHERE code = $3 AND used_by IS NULL`,
		userID, now, code,
	)
	if err != nil {
		return fmt.Errorf("marking invite used: %w", err)
	}
	return nil
}

func (r *Repository) ListAll(ctx context.Context) ([]InviteCode, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, code, created_by, used_by, used_at, expires_at, plan, created_at
		 FROM invite_codes ORDER BY created_at DESC LIMIT 100`,
	)
	if err != nil {
		return nil, fmt.Errorf("listing invites: %w", err)
	}
	defer rows.Close()

	var result []InviteCode
	for rows.Next() {
		var inv InviteCode
		if err := rows.Scan(&inv.ID, &inv.Code, &inv.CreatedBy, &inv.UsedBy, &inv.UsedAt, &inv.ExpiresAt, &inv.Plan, &inv.CreatedAt); err != nil {
			return nil, fmt.Errorf("scanning invite: %w", err)
		}
		result = append(result, inv)
	}
	return result, nil
}

// ValidateAndUse atomically claims an invite for a user and returns the plan
// attached to it. Single UPDATE...WHERE...RETURNING statement so two
// concurrent registrations can't both claim the same code (audit #3): only
// the transaction whose WHERE clause matches an unused, unexpired row gets
// a RETURNING row; the other gets sql.ErrNoRows and fails with "invalid".
//
// Implements auth.InviteValidator.
func (r *Repository) ValidateAndUse(ctx context.Context, code string, userID string) (string, error) {
	var plan string
	err := r.pool.QueryRow(ctx,
		`UPDATE invite_codes
		 SET used_by = $1, used_at = NOW()
		 WHERE code = $2
		   AND used_by IS NULL
		   AND (expires_at IS NULL OR expires_at > NOW())
		 RETURNING plan`,
		userID, code,
	).Scan(&plan)
	if err == pgx.ErrNoRows {
		// Either the code doesn't exist, was already used, or has expired.
		// Callers only need "not usable" — don't leak which condition failed.
		return "", fmt.Errorf("invalid invite")
	}
	if err != nil {
		return "", fmt.Errorf("claiming invite: %w", err)
	}
	return plan, nil
}

func (r *Repository) Delete(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM invite_codes WHERE id = $1`, id)
	return err
}

func generateID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func generateCode() string {
	b := make([]byte, 6)
	rand.Read(b)
	return hex.EncodeToString(b) // 12-char hex code
}
