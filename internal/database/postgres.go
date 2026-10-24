package database

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/oche/diagram/internal/config"
)

// Connect creates a connection pool to PostgreSQL.
func Connect(cfg *config.Config) (*pgxpool.Pool, error) {
	poolCfg, err := pgxpool.ParseConfig(cfg.Database.DSN())
	if err != nil {
		return nil, fmt.Errorf("parsing database config: %w", err)
	}

	poolCfg.MaxConns = cfg.Database.MaxConns
	poolCfg.MinConns = 2
	poolCfg.MaxConnLifetime = 30 * time.Minute
	poolCfg.MaxConnIdleTime = 5 * time.Minute

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return nil, fmt.Errorf("creating pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("pinging database: %w", err)
	}

	if err := EnsureSchema(ctx, pool); err != nil {
		pool.Close()
		return nil, err
	}

	return pool, nil
}
