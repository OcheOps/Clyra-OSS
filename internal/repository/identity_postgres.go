package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/oche/diagram/internal/auth"
)

type IdentityPostgres struct {
	pool *pgxpool.Pool
}

func NewIdentityPostgres(pool *pgxpool.Pool) *IdentityPostgres {
	return &IdentityPostgres{pool: pool}
}

func (r *IdentityPostgres) GetByProviderSub(ctx context.Context, provider, sub string) (*auth.Identity, error) {
	var id auth.Identity
	err := r.pool.QueryRow(ctx,
		`SELECT id, user_id, provider, provider_sub, email, created_at
		 FROM user_identities WHERE provider = $1 AND provider_sub = $2`,
		provider, sub,
	).Scan(&id.ID, &id.UserID, &id.Provider, &id.ProviderSub, &id.Email, &id.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("querying identity: %w", err)
	}
	return &id, nil
}

func (r *IdentityPostgres) Link(ctx context.Context, id *auth.Identity) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO user_identities (id, user_id, provider, provider_sub, email, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6)`,
		id.ID, id.UserID, id.Provider, id.ProviderSub, id.Email, id.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("inserting identity: %w", err)
	}
	return nil
}

func (r *IdentityPostgres) ListForUser(ctx context.Context, userID string) ([]auth.Identity, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, user_id, provider, provider_sub, email, created_at
		 FROM user_identities WHERE user_id = $1 ORDER BY created_at`,
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("listing identities: %w", err)
	}
	defer rows.Close()
	var out []auth.Identity
	for rows.Next() {
		var id auth.Identity
		if err := rows.Scan(&id.ID, &id.UserID, &id.Provider, &id.ProviderSub, &id.Email, &id.CreatedAt); err != nil {
			return nil, fmt.Errorf("scanning identity: %w", err)
		}
		out = append(out, id)
	}
	return out, rows.Err()
}
