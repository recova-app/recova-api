// Package http defines the HTTP runtime assembly for the API service.
package http

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/recova-app/backend-v2/internal/platform/config"
)

// Server wraps the Fiber application lifecycle.
type Server struct {
	app    *fiber.App
	addr   string
	logger *slog.Logger
}

// NewServer creates an HTTP server instance using the runtime configuration.
func NewServer(cfg config.AppConfig, logger *slog.Logger) *Server {
	app := fiber.New(fiber.Config{
		AppName: cfg.AppName,
	})

	registerBaseRoutes(app)

	return &Server{
		app:    app,
		addr:   fmt.Sprintf(":%s", cfg.Port),
		logger: logger,
	}
}

// Start runs the server and performs graceful shutdown when context is canceled.
func (s *Server) Start(ctx context.Context) error {
	errCh := make(chan error, 1)

	go func() {
		errCh <- s.app.Listen(s.addr)
	}()

	select {
	case <-ctx.Done():
		s.logger.Info("shutdown dimulai")

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := s.app.ShutdownWithContext(shutdownCtx); err != nil {
			return fmt.Errorf("shutdown server gagal: %w", err)
		}

		s.logger.Info("shutdown selesai")
		return nil
	case err := <-errCh:
		if err != nil {
			return fmt.Errorf("server gagal listen: %w", err)
		}
		return nil
	}
}

// registerBaseRoutes registers baseline runtime routes required for local checks.
func registerBaseRoutes(app *fiber.App) {
	app.Get("/health/live", func(c fiber.Ctx) error {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"status": "ok",
		})
	})

	app.Get("/health/ready", func(c fiber.Ctx) error {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"status": "ok",
		})
	})
}
