// Package http defines the HTTP runtime assembly for the API service.
package http

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/helmet"
	recoverer "github.com/gofiber/fiber/v3/middleware/recover"
	"github.com/gofiber/fiber/v3/middleware/requestid"
	"github.com/recova-app/backend-v2/internal/platform/config"
	"github.com/recova-app/backend-v2/internal/shared/errs"
	"github.com/recova-app/backend-v2/internal/shared/response"
)

const (
	defaultShutdownTimeout = 5 * time.Second
	requestLogMessage      = "request selesai"
)

// Server wraps the Fiber application lifecycle.
type Server struct {
	app              *fiber.App
	addr             string
	logger           *slog.Logger
	readinessChecks  []ReadinessCheck
	readinessTimeout time.Duration
}

// ServerOption customizes server runtime assembly.
type ServerOption func(*Server)

// WithReadinessChecks overrides default readiness checks with explicit dependencies.
func WithReadinessChecks(checks []ReadinessCheck) ServerOption {
	return func(s *Server) {
		if len(checks) == 0 {
			return
		}

		cloned := make([]ReadinessCheck, 0, len(checks))
		for _, check := range checks {
			if strings.TrimSpace(check.Name) == "" || check.Probe == nil {
				continue
			}
			cloned = append(cloned, check)
		}

		if len(cloned) > 0 {
			s.readinessChecks = cloned
		}
	}
}

// NewServer creates an HTTP server instance using the runtime configuration.
func NewServer(cfg config.Config, logger *slog.Logger, opts ...ServerOption) (*Server, error) {
	bodyLimit, err := parseBodyLimitBytes(cfg.Security.RequestBodyLimit)
	if err != nil {
		return nil, fmt.Errorf("request body limit tidak valid: %w", err)
	}

	srv := &Server{
		addr:             fmt.Sprintf(":%s", cfg.Application.Port),
		logger:           logger,
		readinessChecks:  defaultReadinessChecks(),
		readinessTimeout: time.Duration(cfg.Observability.HealthCheckTimeoutMs) * time.Millisecond,
	}
	for _, opt := range opts {
		if opt == nil {
			continue
		}
		opt(srv)
	}

	app := fiber.New(fiber.Config{
		AppName:      cfg.Application.AppName,
		BodyLimit:    bodyLimit,
		ErrorHandler: srv.errorHandler(cfg.Observability.RequestIDHeader),
	})

	srv.app = app

	srv.registerMiddleware(cfg)
	srv.registerRoutes(cfg)
	srv.registerFallbackRoutes()

	return srv, nil
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

		shutdownCtx, cancel := context.WithTimeout(context.Background(), defaultShutdownTimeout)
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

func (s *Server) registerMiddleware(cfg config.Config) {
	s.app.Use(requestid.New(requestid.Config{
		Header: cfg.Observability.RequestIDHeader,
	}))
	s.app.Use(recoverer.New())
	s.app.Use(requestLogMiddleware(s.logger))
	s.app.Use(helmet.New())
	s.app.Use(cors.New(cors.Config{
		AllowOrigins: cfg.Security.CORSOrigins,
		AllowMethods: []string{
			fiber.MethodGet,
			fiber.MethodPost,
			fiber.MethodPut,
			fiber.MethodPatch,
			fiber.MethodDelete,
			fiber.MethodOptions,
		},
		AllowHeaders: []string{
			"Origin",
			"Content-Type",
			"Accept",
			"Authorization",
			"X-Request-ID",
		},
		ExposeHeaders: []string{
			cfg.Observability.RequestIDHeader,
		},
		AllowCredentials: true,
	}))
}

func (s *Server) registerRoutes(cfg config.Config) {
	s.app.Get("/health/live", func(c fiber.Ctx) error {
		payload := response.Success("Layanan aktif", fiber.Map{"status": "ok"}, nil)
		return c.Status(fiber.StatusOK).JSON(payload)
	})

	s.app.Get("/health/ready", func(c fiber.Ctx) error {
		checkSummary, ready := s.evaluateReadiness(c.Context())
		if !ready {
			return errs.New(
				errs.CodeServiceUnavailable,
				"Layanan belum siap",
				fiber.Map{
					"status": "not_ready",
					"checks": checkSummary,
				},
				nil,
			)
		}

		payload := response.Success("Layanan siap", fiber.Map{
			"status": "ready",
			"checks": checkSummary,
		}, nil)
		return c.Status(fiber.StatusOK).JSON(payload)
	})

	// API group baseline disiapkan sebagai titik registrasi modul domain.
	_ = s.app.Group(strings.TrimSpace(cfg.Application.APIPrefix))
}

func (s *Server) registerFallbackRoutes() {
	s.app.Use(func(c fiber.Ctx) error {
		return errs.New(errs.CodeNotFound, "Rute tidak ditemukan", nil, nil)
	})
}

func requestLogMiddleware(logger *slog.Logger) fiber.Handler {
	return func(c fiber.Ctx) error {
		started := time.Now()
		err := c.Next()
		latency := time.Since(started)

		reqID := requestid.FromContext(c)
		statusCode := c.Response().StatusCode()
		path := c.Path()
		method := c.Method()

		if err != nil {
			logger.Error(requestLogMessage,
				"requestId", strings.TrimSpace(reqID),
				"method", method,
				"path", path,
				"status", statusCode,
				"latencyMs", latency.Milliseconds(),
				"error", err,
			)
			return err
		}

		logger.Info(requestLogMessage,
			"requestId", strings.TrimSpace(reqID),
			"method", method,
			"path", path,
			"status", statusCode,
			"latencyMs", latency.Milliseconds(),
		)
		return nil
	}
}

func (s *Server) errorHandler(requestIDHeader string) fiber.ErrorHandler {
	return func(c fiber.Ctx, err error) error {
		mapped := errs.Map(err)
		reqID := strings.TrimSpace(requestid.FromContext(c))
		if reqID == "" {
			reqID = strings.TrimSpace(c.Get(requestIDHeader))
		}

		s.logger.Error("request gagal",
			"requestId", reqID,
			"method", c.Method(),
			"path", c.Path(),
			"status", mapped.Status,
			"errorCode", mapped.Code,
			"error", err,
		)

		payload := response.Error(mapped.Message, string(mapped.Code), mapped.Details, reqID)
		return c.Status(mapped.Status).JSON(payload)
	}
}
