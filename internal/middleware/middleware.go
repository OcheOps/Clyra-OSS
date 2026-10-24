package middleware

import (
	"log/slog"

	"github.com/oche/diagram/internal/auth"
	"github.com/oche/diagram/internal/config"
)

// Middleware holds shared dependencies for all middleware handlers.
type Middleware struct {
	cfg         *config.Config
	logger      *slog.Logger
	authService *auth.Service
	limiter     RateLimiter
}

func New(cfg *config.Config, logger *slog.Logger, authService *auth.Service, limiter RateLimiter) *Middleware {
	return &Middleware{
		cfg:         cfg,
		logger:      logger,
		authService: authService,
		limiter:     limiter,
	}
}
