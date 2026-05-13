// Package bootstrap assembles runtime dependencies for the application.
package bootstrap

import (
	"context"
	"log/slog"

	apphttp "github.com/recova-app/backend-v2/internal/app/http"
	achievementsmodule "github.com/recova-app/backend-v2/internal/modules/achievements"
	aimodule "github.com/recova-app/backend-v2/internal/modules/ai"
	authmodule "github.com/recova-app/backend-v2/internal/modules/auth"
	communitymodule "github.com/recova-app/backend-v2/internal/modules/community"
	contentmodule "github.com/recova-app/backend-v2/internal/modules/content"
	educationmodule "github.com/recova-app/backend-v2/internal/modules/education"
	journalsmodule "github.com/recova-app/backend-v2/internal/modules/journals"
	routinemodule "github.com/recova-app/backend-v2/internal/modules/routine"
	usersmodule "github.com/recova-app/backend-v2/internal/modules/users"
	aiplatform "github.com/recova-app/backend-v2/internal/platform/ai"
	"github.com/recova-app/backend-v2/internal/platform/config"
	"github.com/recova-app/backend-v2/internal/platform/database"
	"github.com/recova-app/backend-v2/internal/platform/observability"
)

// Application coordinates the top-level runtime process for the API service.
type Application struct {
	server   *apphttp.Server
	logger   *slog.Logger
	database *database.Client
}

// NewApplication constructs an executable application instance from runtime dependencies.
func NewApplication(cfg config.Config, logger *slog.Logger) (*Application, error) {
	obsRecorder := observability.NewRecorder()

	dbClient, err := database.Connect(cfg)
	if err != nil {
		return nil, err
	}
	if err := observability.RegisterDatabaseMetrics(dbClient.Gorm(), obsRecorder); err != nil {
		_ = dbClient.Close()
		return nil, err
	}

	authService := authmodule.NewService(
		authmodule.NewRepository(dbClient.Gorm()),
		&authmodule.GoogleIDTokenVerifier{},
		authmodule.NewTokenManager(cfg),
	)
	routineService := routinemodule.NewService(routinemodule.NewRepository(dbClient.Gorm()))
	journalsService := journalsmodule.NewService(journalsmodule.NewRepository(dbClient.Gorm()))
	communityService := communitymodule.NewService(communitymodule.NewRepository(dbClient.Gorm()))
	achievementsService := achievementsmodule.NewService(achievementsmodule.NewRepository(dbClient.Gorm()))
	educationService := educationmodule.NewService(educationmodule.NewRepository(dbClient.Gorm()))
	contentService := contentmodule.NewService(contentmodule.NewRepository(dbClient.Gorm()))
	aiClient, err := aiplatform.NewClient(cfg.AI)
	if err != nil {
		_ = dbClient.Close()
		return nil, err
	}
	aiClient = observability.WrapAIClient(aiClient, obsRecorder)
	aiService := aimodule.NewService(aimodule.NewRepository(dbClient.Gorm()), aiClient)
	aiService.SetTelemetry(observability.NewAIPersonaTelemetry(obsRecorder))
	usersService := usersmodule.NewService(
		usersmodule.NewRepository(dbClient.Gorm()),
		cfg.Application.AppEnv,
		cfg.Application.NodeEnv,
		aiService,
	)

	server, err := apphttp.NewServer(cfg, logger, apphttp.WithReadinessChecks([]apphttp.ReadinessCheck{
		{
			Name:    "database",
			Mode:    apphttp.ReadinessModeRequired,
			Message: "Koneksi database tidak sehat",
			Probe:   dbClient.Ping,
		},
	}), apphttp.WithModuleDependencies(apphttp.ModuleDependencies{
		AuthService:         authService,
		UsersService:        usersService,
		RoutineService:      routineService,
		JournalsService:     journalsService,
		CommunityService:    communityService,
		AchievementsService: achievementsService,
		EducationService:    educationService,
		ContentService:      contentService,
		AIService:           aiService,
	}), apphttp.WithObservability(obsRecorder))
	if err != nil {
		_ = dbClient.Close()
		return nil, err
	}

	return &Application{
		server:   server,
		logger:   logger,
		database: dbClient,
	}, nil
}

// Run starts the application runtime and blocks until shutdown.
func (a *Application) Run(ctx context.Context) error {
	defer func() {
		if a.database == nil {
			return
		}
		if err := a.database.Close(); err != nil {
			a.logger.Error("failed to close database connection", "error", err)
		}
	}()

	a.logger.Info("starting application", "app", "api")
	return a.server.Start(ctx)
}
