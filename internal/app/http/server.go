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
	achievementsmodule "github.com/recova-app/backend-v2/internal/modules/achievements"
	aimodule "github.com/recova-app/backend-v2/internal/modules/ai"
	authmodule "github.com/recova-app/backend-v2/internal/modules/auth"
	communitymodule "github.com/recova-app/backend-v2/internal/modules/community"
	contentmodule "github.com/recova-app/backend-v2/internal/modules/content"
	educationmodule "github.com/recova-app/backend-v2/internal/modules/education"
	journalsmodule "github.com/recova-app/backend-v2/internal/modules/journals"
	routinemodule "github.com/recova-app/backend-v2/internal/modules/routine"
	usersmodule "github.com/recova-app/backend-v2/internal/modules/users"
	"github.com/recova-app/backend-v2/internal/platform/config"
	"github.com/recova-app/backend-v2/internal/platform/observability"
	"github.com/recova-app/backend-v2/internal/shared/errs"
	"github.com/recova-app/backend-v2/internal/shared/response"
)

const (
	defaultShutdownTimeout = 5 * time.Second
)

// Server wraps the Fiber application lifecycle.
type Server struct {
	app              *fiber.App
	addr             string
	logger           *slog.Logger
	appEnv           string
	readinessChecks  []ReadinessCheck
	readinessTimeout time.Duration
	moduleDeps       ModuleDependencies
	routeRegistrars  []RouteRegistrar
	observability    *observability.Recorder
}

// RouteRegistrar defines a typed contract for registering an API module's routes.
type RouteRegistrar interface {
	RegisterRoutes(router fiber.Router)
}

// RouteRegistrarFunc adapts an ordinary function to the RouteRegistrar interface.
type RouteRegistrarFunc func(router fiber.Router)

// RegisterRoutes calls the underlying function with the provided router.
func (f RouteRegistrarFunc) RegisterRoutes(router fiber.Router) {
	f(router)
}

// ModuleDependencies stores domain services required to register API routes.
type ModuleDependencies struct {
	AuthService         *authmodule.Service
	UsersService        *usersmodule.Service
	RoutineService      *routinemodule.Service
	JournalsService     *journalsmodule.Service
	CommunityService    *communitymodule.Service
	AchievementsService *achievementsmodule.Service
	EducationService    *educationmodule.Service
	ContentService      *contentmodule.Service
	AIService           *aimodule.Service
}

// Registrars converts ModuleDependencies into typed route registrars for configured modules.
func (d ModuleDependencies) Registrars(cfg config.Config) []RouteRegistrar {
	var registrars []RouteRegistrar

	if d.AuthService != nil {
		registrars = append(registrars, RouteRegistrarFunc(func(api fiber.Router) {
			authGroup := api.Group("/auth")
			authGroup.Use(authRequestLimiter(cfg))
			authmodule.RegisterCoreRoutes(authGroup, d.AuthService)
			if d.UsersService != nil {
				usersmodule.RegisterOnboardingRoute(authGroup, d.AuthService, d.UsersService)
			}
		}))
	}

	if d.UsersService != nil && d.AuthService != nil {
		registrars = append(registrars, RouteRegistrarFunc(func(api fiber.Router) {
			usersGroup := api.Group("/users")
			usersmodule.RegisterUserRoutes(usersGroup, d.AuthService, d.UsersService)
		}))
	}

	if d.RoutineService != nil && d.AuthService != nil {
		registrars = append(registrars, RouteRegistrarFunc(func(api fiber.Router) {
			routineGroup := api.Group("/routine")
			routinemodule.RegisterRoutes(routineGroup, d.AuthService, d.RoutineService)
		}))
	}

	if d.JournalsService != nil && d.AuthService != nil {
		registrars = append(registrars, RouteRegistrarFunc(func(api fiber.Router) {
			journalsGroup := api.Group("/journals")
			journalsmodule.RegisterRoutes(journalsGroup, d.AuthService, d.JournalsService)
		}))
	}

	if d.CommunityService != nil && d.AuthService != nil {
		registrars = append(registrars, RouteRegistrarFunc(func(api fiber.Router) {
			communityGroup := api.Group("/community")
			communitymodule.RegisterRoutes(
				communityGroup,
				d.AuthService,
				d.CommunityService,
				communityWriteLimiter(cfg),
			)
		}))
	}

	if d.AchievementsService != nil && d.AuthService != nil {
		registrars = append(registrars, RouteRegistrarFunc(func(api fiber.Router) {
			achievementsGroup := api.Group("/achievements")
			achievementsmodule.RegisterRoutes(achievementsGroup, d.AuthService, d.AchievementsService)
		}))
	}

	if d.EducationService != nil && d.AuthService != nil {
		registrars = append(registrars, RouteRegistrarFunc(func(api fiber.Router) {
			educationGroup := api.Group("/education")
			educationmodule.RegisterRoutes(educationGroup, d.AuthService, d.EducationService)
		}))
	}

	if d.ContentService != nil && d.AuthService != nil {
		registrars = append(registrars, RouteRegistrarFunc(func(api fiber.Router) {
			contentGroup := api.Group("/content")
			contentmodule.RegisterRoutes(contentGroup, d.AuthService, d.ContentService)
		}))
	}

	if d.AIService != nil && d.AuthService != nil {
		registrars = append(registrars, RouteRegistrarFunc(func(api fiber.Router) {
			aiGroup := api.Group("/ai")
			aimodule.RegisterRoutes(aiGroup, d.AuthService, d.AIService, aiRequestLimiter(cfg))
		}))
	}

	return registrars
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

// WithObservability configures shared observability recorder for middleware and routes.
func WithObservability(recorder *observability.Recorder) ServerOption {
	return func(s *Server) {
		s.observability = recorder
	}
}

// WithRouteRegistrars appends explicit route registrars to the runtime assembly.
func WithRouteRegistrars(registrars ...RouteRegistrar) ServerOption {
	return func(s *Server) {
		for _, reg := range registrars {
			if reg != nil {
				s.routeRegistrars = append(s.routeRegistrars, reg)
			}
		}
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
		appEnv:           cfg.Application.AppEnv,
		readinessChecks:  defaultReadinessChecks(),
		readinessTimeout: time.Duration(cfg.Observability.HealthCheckTimeoutMs) * time.Millisecond,
	}
	for _, opt := range opts {
		if opt == nil {
			continue
		}
		opt(srv)
	}

	if srv.appEnv == "production" {
		if err := srv.validateProductionReadiness(); err != nil {
			return nil, err
		}
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

func (s *Server) validateProductionReadiness() error {
	hasRequiredDatabaseProbe := false
	for _, check := range s.readinessChecks {
		if check.Mode == ReadinessModePlaceholder {
			return fmt.Errorf("production environment cannot use placeholder readiness check: %s", check.Name)
		}
		if check.Name == "database" && check.Mode == ReadinessModeRequired && check.Probe != nil {
			hasRequiredDatabaseProbe = true
		}
	}
	if !hasRequiredDatabaseProbe {
		return fmt.Errorf("production environment requires an active database readiness probe")
	}
	return nil
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
	s.app.Use(observability.NewRequestTelemetryMiddleware(
		s.logger,
		s.observability,
		cfg.Observability.RequestIDHeader,
		observability.RequestHooks{
			OnComplete: func(ctx observability.RequestContext) {
				observability.RecordAuditEvent(s.logger, s.observability, ctx)
			},
		},
	))
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
	if s.observability != nil {
		s.app.Get("/metrics", s.observability.MetricsHandler())
	}

	s.app.Get("/health/live", func(c fiber.Ctx) error {
		payload := response.Success("Layanan aktif", fiber.Map{"status": "ok"}, nil)
		return c.Status(fiber.StatusOK).JSON(payload)
	})

	s.app.Get("/health/ready", func(c fiber.Ctx) error {
		reqID := strings.TrimSpace(requestid.FromContext(c))
		if reqID == "" {
			reqID = strings.TrimSpace(c.Get(cfg.Observability.RequestIDHeader))
		}
		checkSummary, ready := s.evaluateReadiness(c.Context(), reqID)
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
	for _, registrar := range s.moduleDeps.Registrars(cfg) {
		registrar.RegisterRoutes(apiGroup)
	}
	for _, registrar := range s.routeRegistrars {
		registrar.RegisterRoutes(apiGroup)
	}

	s.registerAPIDocsRoutes()
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
			return limiterKeyFromContext(c)
		},
		LimitReached: func(_ fiber.Ctx) error {
			return errs.New(errs.CodeRateLimited, "Terlalu banyak permintaan komunitas, coba lagi sebentar", nil, nil)
		},
	})
}

func aiRequestLimiter(cfg config.Config) fiber.Handler {
	window := time.Duration(cfg.Security.RateLimit.WindowMs) * time.Millisecond
	if window <= 0 {
		window = time.Minute
	}

	maxAI := cfg.Security.RateLimit.AIMax
	if maxAI <= 0 {
		maxAI = 1
	}
	if cfg.Security.RateLimit.Max > 0 && maxAI > cfg.Security.RateLimit.Max {
		maxAI = cfg.Security.RateLimit.Max
	}

	return limiter.New(limiter.Config{
		Max:        maxAI,
		Expiration: window,
		KeyGenerator: func(c fiber.Ctx) string {
			return limiterKeyFromContext(c)
		},
		LimitReached: func(_ fiber.Ctx) error {
			return errs.New(errs.CodeRateLimited, "Terlalu banyak permintaan AI, coba lagi sebentar", nil, nil)
		},
	})
}

func authRequestLimiter(cfg config.Config) fiber.Handler {
	window := time.Duration(cfg.Security.RateLimit.WindowMs) * time.Millisecond
	if window <= 0 {
		window = time.Minute
	}

	maxAuth := cfg.Security.RateLimit.AuthMax
	if maxAuth <= 0 {
		maxAuth = 1
	}
	if cfg.Security.RateLimit.Max > 0 && maxAuth > cfg.Security.RateLimit.Max {
		maxAuth = cfg.Security.RateLimit.Max
	}

	return limiter.New(limiter.Config{
		Max:        maxAuth,
		Expiration: window,
		KeyGenerator: func(c fiber.Ctx) string {
			return limiterKeyFromContext(c)
		},
		LimitReached: func(_ fiber.Ctx) error {
			return errs.New(errs.CodeRateLimited, "Terlalu banyak permintaan autentikasi, coba lagi sebentar", nil, nil)
		},
	})
}

func limiterKeyFromContext(c fiber.Ctx) string {
	principal, ok := authmodule.PrincipalFromContext(c)
	if ok && strings.TrimSpace(principal.UserID) != "" {
		return strings.TrimSpace(principal.UserID)
	}
	return strings.TrimSpace(c.IP())
}

func (s *Server) errorHandler(requestIDHeader string) fiber.ErrorHandler {
	return func(c fiber.Ctx, err error) error {
		mapped := errs.Map(err)
		reqID := strings.TrimSpace(requestid.FromContext(c))
		if reqID == "" {
			reqID = strings.TrimSpace(c.Get(requestIDHeader))
		}

		s.logger.Error("request failed",
			"request_id", reqID,
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
