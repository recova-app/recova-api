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
	"github.com/gofiber/fiber/v3/middleware/limiter"
	recoverer "github.com/gofiber/fiber/v3/middleware/recover"
	"github.com/gofiber/fiber/v3/middleware/requestid"
	authmodule "github.com/recova-app/backend-v2/internal/modules/auth"
	communitymodule "github.com/recova-app/backend-v2/internal/modules/community"
	contentmodule "github.com/recova-app/backend-v2/internal/modules/content"
	educationmodule "github.com/recova-app/backend-v2/internal/modules/education"
	journalsmodule "github.com/recova-app/backend-v2/internal/modules/journals"
	routinemodule "github.com/recova-app/backend-v2/internal/modules/routine"
	usersmodule "github.com/recova-app/backend-v2/internal/modules/users"
	"github.com/recova-app/backend-v2/internal/platform/config"
	"github.com/recova-app/backend-v2/internal/shared/errs"
	"github.com/recova-app/backend-v2/internal/shared/response"
)

const (
	defaultShutdownTimeout = 5 * time.Second
	requestLogMessage      = "request completed"
)

// Server wraps the Fiber application lifecycle.
type Server struct {
	app              *fiber.App
	addr             string
	logger           *slog.Logger
	readinessChecks  []ReadinessCheck
	readinessTimeout time.Duration
	moduleDeps       ModuleDependencies
}

// ModuleDependencies stores domain services required to register API routes.
type ModuleDependencies struct {
	AuthService      *authmodule.Service
	UsersService     *usersmodule.Service
	RoutineService   *routinemodule.Service
	JournalsService  *journalsmodule.Service
	CommunityService *communitymodule.Service
	EducationService *educationmodule.Service
	ContentService   *contentmodule.Service
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

// WithModuleDependencies configures runtime domain service dependencies for route registration.
func WithModuleDependencies(deps ModuleDependencies) ServerOption {
	return func(s *Server) {
		s.moduleDeps = deps
	}
}

// NewServer creates an HTTP server instance using the runtime configuration.
func NewServer(cfg config.Config, logger *slog.Logger, opts ...ServerOption) (*Server, error) {
	bodyLimit, err := parseBodyLimitBytes(cfg.Security.RequestBodyLimit)
	if err != nil {
		return nil, fmt.Errorf("invalid request body limit: %w", err)
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
		s.logger.Info("shutdown started")

		shutdownCtx, cancel := context.WithTimeout(context.Background(), defaultShutdownTimeout)
		defer cancel()

		if err := s.app.ShutdownWithContext(shutdownCtx); err != nil {
			return fmt.Errorf("server shutdown failed: %w", err)
		}

		s.logger.Info("shutdown completed")
		return nil
	case err := <-errCh:
		if err != nil {
			return fmt.Errorf("server listen failed: %w", err)
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

	apiGroup := s.app.Group(strings.TrimSpace(cfg.Application.APIPrefix))
	if s.moduleDeps.AuthService != nil {
		authGroup := apiGroup.Group("/auth")
		authmodule.RegisterCoreRoutes(authGroup, s.moduleDeps.AuthService)
		if s.moduleDeps.UsersService != nil {
			usersmodule.RegisterOnboardingRoute(authGroup, s.moduleDeps.AuthService, s.moduleDeps.UsersService)
		}
	}

	if s.moduleDeps.UsersService != nil && s.moduleDeps.AuthService != nil {
		usersGroup := apiGroup.Group("/users")
		usersmodule.RegisterUserRoutes(usersGroup, s.moduleDeps.AuthService, s.moduleDeps.UsersService)
	}

	if s.moduleDeps.RoutineService != nil && s.moduleDeps.AuthService != nil {
		routineGroup := apiGroup.Group("/routine")
		routinemodule.RegisterRoutes(routineGroup, s.moduleDeps.AuthService, s.moduleDeps.RoutineService)
	}

	if s.moduleDeps.JournalsService != nil && s.moduleDeps.AuthService != nil {
		journalsGroup := apiGroup.Group("/journals")
		journalsmodule.RegisterRoutes(journalsGroup, s.moduleDeps.AuthService, s.moduleDeps.JournalsService)
	}

	if s.moduleDeps.CommunityService != nil && s.moduleDeps.AuthService != nil {
		communityGroup := apiGroup.Group("/community")
		communitymodule.RegisterRoutes(
			communityGroup,
			s.moduleDeps.AuthService,
			s.moduleDeps.CommunityService,
			communityWriteLimiter(cfg),
		)
	}

	if s.moduleDeps.EducationService != nil && s.moduleDeps.AuthService != nil {
		educationGroup := apiGroup.Group("/education")
		educationmodule.RegisterRoutes(educationGroup, s.moduleDeps.AuthService, s.moduleDeps.EducationService)
	}

	if s.moduleDeps.ContentService != nil && s.moduleDeps.AuthService != nil {
		contentGroup := apiGroup.Group("/content")
		contentmodule.RegisterRoutes(contentGroup, s.moduleDeps.AuthService, s.moduleDeps.ContentService)
	}
}

func (s *Server) registerFallbackRoutes() {
	s.app.Use(func(c fiber.Ctx) error {
		return errs.New(errs.CodeNotFound, "Rute tidak ditemukan", nil, nil)
	})
}

func communityWriteLimiter(cfg config.Config) fiber.Handler {
	window := time.Duration(cfg.Security.RateLimit.WindowMs) * time.Millisecond
	if window <= 0 {
		window = time.Minute
	}

	maxWrite := cfg.Security.RateLimit.AuthMax
	if maxWrite <= 0 {
		maxWrite = 1
	}
	if cfg.Security.RateLimit.Max > 0 && maxWrite > cfg.Security.RateLimit.Max {
		maxWrite = cfg.Security.RateLimit.Max
	}

	return limiter.New(limiter.Config{
		Max:        maxWrite,
		Expiration: window,
		KeyGenerator: func(c fiber.Ctx) string {
			principal, ok := authmodule.PrincipalFromContext(c)
			if ok && strings.TrimSpace(principal.UserID) != "" {
				return strings.TrimSpace(principal.UserID)
			}
			return strings.TrimSpace(c.IP())
		},
		LimitReached: func(_ fiber.Ctx) error {
			return errs.New(errs.CodeRateLimited, "Terlalu banyak permintaan komunitas, coba lagi sebentar", nil, nil)
		},
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

		s.logger.Error("request failed",
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
