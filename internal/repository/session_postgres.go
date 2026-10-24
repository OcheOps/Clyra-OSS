package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/oche/diagram/internal/auth"
)

type SessionPostgres struct {
	pool *pgxpool.Pool
}

func NewSessionPostgres(pool *pgxpool.Pool) *SessionPostgres {
	return &SessionPostgres{pool: pool}
}

func (r *SessionPostgres) Create(ctx context.Context, session *auth.Session) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO sessions (id, user_id, expires_at, created_at)
		 VALUES ($1, $2, $3, $4)`,
		session.ID, session.UserID, session.ExpiresAt, session.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("creating session: %w", err)
	}
	return nil
}

func (r *SessionPostgres) GetByID(ctx context.Context, id string) (*auth.Session, error) {
	var s auth.Session
	err := r.pool.QueryRow(ctx,
		`SELECT id, user_id, expires_at, created_at
		 FROM sessions WHERE id = $1`,
		id,
	).Scan(&s.ID, &s.UserID, &s.ExpiresAt, &s.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("session not found")
	}
	if err != nil {
		return nil, fmt.Errorf("querying session: %w", err)
	}
	return &s, nil
}

func (r *SessionPostgres) Delete(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM sessions WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("deleting session: %w", err)
	}
	return nil
}

func (r *SessionPostgres) DeleteByUserID(ctx context.Context, userID string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM sessions WHERE user_id = $1`, userID)
	if err != nil {
		return fmt.Errorf("deleting user sessions: %w", err)
	}
	return nil
}
