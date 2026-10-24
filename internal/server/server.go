package server

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/oche/diagram/internal/config"
)

type Server struct {
	httpServer *http.Server
	logger     *slog.Logger
}

func New(cfg *config.Config, handler http.Handler, logger *slog.Logger) *Server {
	return &Server{
		httpServer: &http.Server{
			Addr:         fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
			Handler:      handler,
			ReadTimeout:  cfg.Server.ReadTimeout,
			WriteTimeout: cfg.Server.WriteTimeout,
			IdleTimeout:  cfg.Server.IdleTimeout,
		},
		logger: logger,
	}
}

// GracefulShutdownTimeout is how long we let in-flight requests finish
// after SIGTERM before force-closing the server. Tuned for the longest
// legitimate request path — IaC generation on Opus 4.8 sits around
// 3 minutes; 90 seconds is enough for the median in-flight batch to
// drain while a compose-level stop_grace_period of 120s gives us
// headroom against docker's SIGKILL. Bump both together.
const GracefulShutdownTimeout = 90 * time.Second

// Start begins listening and blocks until a shutdown signal is received.
// Performs graceful shutdown with a GracefulShutdownTimeout window.
func (s *Server) Start() error {
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	serverErr := make(chan error, 1)

	go func() {
		s.logger.Info("server starting", "addr", s.httpServer.Addr)
		serverErr <- s.httpServer.ListenAndServe()
	}()

	select {
	case err := <-serverErr:
		if err != nil && err != http.ErrServerClosed {
			return fmt.Errorf("server error: %w", err)
		}
	case sig := <-shutdown:
		s.logger.Info("shutdown signal received, draining in-flight requests",
			"signal", sig.String(),
			"timeout", GracefulShutdownTimeout.String())

		ctx, cancel := context.WithTimeout(context.Background(), GracefulShutdownTimeout)
		defer cancel()

		if err := s.httpServer.Shutdown(ctx); err != nil {
			s.logger.Error("graceful shutdown failed, forcing close", "error", err)
			s.httpServer.Close()
			return fmt.Errorf("shutdown: %w", err)
		}

		s.logger.Info("server stopped gracefully")
	}

	return nil
}
