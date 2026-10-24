package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/oche/diagram/internal/auth"
)

type UserPostgres struct {
	pool *pgxpool.Pool
}

func NewUserPostgres(pool *pgxpool.Pool) *UserPostgres {
	return &UserPostgres{pool: pool}
}

func (r *UserPostgres) Create(ctx context.Context, user *auth.User) error {
	plan := user.Plan
	if plan == "" {
		plan = auth.PlanFree
	}
	// OAuth-only users arrive with an empty PasswordHash — store NULL so a
	// DB leak can't feed an empty-string hash to bcrypt.Compare and succeed
	// on pathological inputs.
	var pwHash any
	if user.PasswordHash == "" {
		pwHash = nil
	} else {
		pwHash = user.PasswordHash
	}
	_, err := r.pool.Exec(ctx,
		`INSERT INTO users (id, email, password_hash, role, plan, plan_expiry, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		user.ID, user.Email, pwHash, user.Role, plan, user.PlanExpiry, user.CreatedAt, user.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("inserting user: %w", err)
	}
	user.Plan = plan
	return nil
}

func (r *UserPostgres) GetByEmail(ctx context.Context, email string) (*auth.User, error) {
	return r.scanUser(ctx,
		`SELECT id, email, password_hash, role, plan, plan_expiry, created_at, updated_at
		 FROM users WHERE email = $1`,
		email,
	)
}

func (r *UserPostgres) GetByID(ctx context.Context, id string) (*auth.User, error) {
	return r.scanUser(ctx,
		`SELECT id, email, password_hash, role, plan, plan_expiry, created_at, updated_at
		 FROM users WHERE id = $1`,
		id,
	)
}

// UpdatePlan changes the plan for an existing user (leaves expiry unchanged).
func (r *UserPostgres) UpdatePlan(ctx context.Context, userID string, plan auth.Plan) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE users SET plan = $1, updated_at = NOW() WHERE id = $2`,
		plan, userID,
	)
	if err != nil {
		return fmt.Errorf("updating plan: %w", err)
	}
	return nil
}

// UpdatePasswordHash replaces the bcrypt hash for an existing user.
func (r *UserPostgres) UpdatePasswordHash(ctx context.Context, userID, hash string) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE users SET password_hash = $1, updated_at = NOW() WHERE id = $2`,
		hash, userID,
	)
	if err != nil {
		return fmt.Errorf("updating password hash: %w", err)
	}
	return nil
}

// UpgradePlan sets plan and plan_expiry atomically (used by payment verification).
func (r *UserPostgres) UpgradePlan(ctx context.Context, userID string, plan auth.Plan, expiry time.Time) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE users SET plan = $1, plan_expiry = $2, updated_at = NOW() WHERE id = $3`,
		plan, expiry, userID,
	)
	if err != nil {
		return fmt.Errorf("upgrading plan: %w", err)
	}
	return nil
}

// ExtendPlan bumps plan_expiry by `days` from the later of (current expiry,
// now). Used for subscription renewals: if the renewal webhook fires before
// the previous period ends, we accumulate the remaining paid days rather
// than restarting the clock at NOW(). Also sets plan in case the user was
// downgraded and is now re-upgrading (e.g. Pro after a brief lapse).
func (r *UserPostgres) ExtendPlan(ctx context.Context, userID string, plan auth.Plan, days int) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE users
		 SET plan = $1,
		     plan_expiry = GREATEST(COALESCE(plan_expiry, NOW()), NOW()) + ($2 || ' days')::interval,
		     updated_at = NOW()
		 WHERE id = $3`,
		plan, days, userID,
	)
	if err != nil {
		return fmt.Errorf("extending plan: %w", err)
	}
	return nil
}

// Delete removes a user by ID. Compensation path for register flows that
// fail after Create — the user's email would otherwise be stuck as taken.
// Cascading FKs (sessions, payments, projects) drop with the row.
func (r *UserPostgres) Delete(ctx context.Context, userID string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM users WHERE id = $1`, userID)
	if err != nil {
		return fmt.Errorf("deleting user: %w", err)
	}
	return nil
}

// CancelPlan immediately downgrades the user to free and clears the expiry.
// No refund is issued — the user agrees to forfeit remaining paid days.
func (r *UserPostgres) CancelPlan(ctx context.Context, userID string) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE users SET plan = 'free', plan_expiry = NULL, updated_at = NOW() WHERE id = $1`,
		userID,
	)
	if err != nil {
		return fmt.Errorf("canceling plan: %w", err)
	}
	return nil
}

// UpdateRole sets the role for an existing user.
func (r *UserPostgres) UpdateRole(ctx context.Context, userID string, role auth.Role) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE users SET role = $1, updated_at = NOW() WHERE id = $2`,
		role, userID,
	)
	if err != nil {
		return fmt.Errorf("updating role: %w", err)
	}
	return nil
}

// CountAdmins returns the current number of admin users. Used to prevent
// removing the last admin and locking yourself out.
func (r *UserPostgres) CountAdmins(ctx context.Context) (int, error) {
	var n int
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM users WHERE role = 'admin'`).Scan(&n)
	if err != nil {
		return 0, fmt.Errorf("counting admins: %w", err)
	}
	return n, nil
}

// UserListRow mirrors the admin-facing user summary. Kept in the repo
// package so the admin package doesn't depend on the SQL shape.
type UserListRow struct {
	ID         string
	Email      string
	Role       string
	Plan       string
	PlanExpiry *time.Time
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// List returns users for the admin panel. `query` matches email (ILIKE).
// Ordered newest first. Both limit and offset are clamped by the caller.
func (r *UserPostgres) List(ctx context.Context, query string, limit, offset int) ([]UserListRow, int, error) {
	if limit <= 0 {
		limit = 25
	}
	if limit > 200 {
		limit = 200
	}
	var rows []UserListRow
	var total int

	if query == "" {
		if err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM users`).Scan(&total); err != nil {
			return nil, 0, fmt.Errorf("counting users: %w", err)
		}
		rs, err := r.pool.Query(ctx,
			`SELECT id, email, role, plan, plan_expiry, created_at, updated_at
			 FROM users ORDER BY created_at DESC LIMIT $1 OFFSET $2`,
			limit, offset,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("listing users: %w", err)
		}
		defer rs.Close()
		for rs.Next() {
			var r UserListRow
			if err := rs.Scan(&r.ID, &r.Email, &r.Role, &r.Plan, &r.PlanExpiry, &r.CreatedAt, &r.UpdatedAt); err != nil {
				return nil, 0, err
			}
			rows = append(rows, r)
		}
		return rows, total, rs.Err()
	}

	pattern := "%" + query + "%"
	if err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM users WHERE email ILIKE $1`, pattern).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("counting users: %w", err)
	}
	rs, err := r.pool.Query(ctx,
		`SELECT id, email, role, plan, plan_expiry, created_at, updated_at
		 FROM users WHERE email ILIKE $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`,
		pattern, limit, offset,
	)
	if err != nil {
		return nil, 0, fmt.Errorf("listing users: %w", err)
	}
	defer rs.Close()
	for rs.Next() {
		var r UserListRow
		if err := rs.Scan(&r.ID, &r.Email, &r.Role, &r.Plan, &r.PlanExpiry, &r.CreatedAt, &r.UpdatedAt); err != nil {
			return nil, 0, err
		}
		rows = append(rows, r)
	}
	return rows, total, rs.Err()
}

func (r *UserPostgres) scanUser(ctx context.Context, query string, args ...any) (*auth.User, error) {
	var u auth.User
	var pwHash *string
	err := r.pool.QueryRow(ctx, query, args...).Scan(
		&u.ID, &u.Email, &pwHash, &u.Role, &u.Plan, &u.PlanExpiry, &u.CreatedAt, &u.UpdatedAt,
	)
	if pwHash != nil {
		u.PasswordHash = *pwHash
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("user not found")
	}
	if err != nil {
		return nil, fmt.Errorf("querying user: %w", err)
	}
	return &u, nil
}
