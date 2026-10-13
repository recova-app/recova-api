// Package bootstrap assembles runtime dependencies for the application.
package bootstrap

import (
	"context"
	"log/slog"

	"gorm.io/gorm"

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

type identityServices struct {
	authService  *authmodule.Service
	usersService *usersmodule.Service
}

type recoveryServices struct {
	routineService      *routinemodule.Service
	journalsService     *journalsmodule.Service
	achievementsService *achievementsmodule.Service
}

type contentServices struct {
	communityService *communitymodule.Service
	educationService *educationmodule.Service
	contentService   *contentmodule.Service
}

func assembleAIService(
	cfg config.Config,
	db *gorm.DB,
	obsRecorder *observability.Recorder,
) (*aimodule.Service, error) {
	aiClient, err := aiplatform.NewClient(cfg.AI)
	if err != nil {
		return nil, err
	}
	aiClient = observability.WrapAIClient(aiClient, obsRecorder)
	aiService := aimodule.NewService(aimodule.NewRepository(db), aiClient)
	aiService.SetTelemetry(observability.NewAIPersonaTelemetry(obsRecorder))
	return aiService, nil
}

func assembleIdentityServices(
	cfg config.Config,
	db *gorm.DB,
	aiService *aimodule.Service,
) identityServices {
	authService := authmodule.NewService(
		authmodule.NewRepository(db),
		&authmodule.GoogleIDTokenVerifier{},
		authmodule.NewTokenManager(cfg),
	)
	usersService := usersmodule.NewService(
		usersmodule.NewRepository(db),
		cfg.Application.AppEnv,
		cfg.Application.NodeEnv,
		aiService,
	)
	return identityServices{
		authService:  authService,
		usersService: usersService,
	}
}

func assembleRecoveryServices(
	db *gorm.DB,
	aiService *aimodule.Service,
) recoveryServices {
	return recoveryServices{
		routineService:      routinemodule.NewService(routinemodule.NewRepository(db), aiService),
		journalsService:     journalsmodule.NewService(journalsmodule.NewRepository(db)),
		achievementsService: achievementsmodule.NewService(achievementsmodule.NewRepository(db)),
	}
}

func assembleContentServices(db *gorm.DB) contentServices {
	return contentServices{
		communityService: communitymodule.NewService(communitymodule.NewRepository(db)),
		educationService: educationmodule.NewService(educationmodule.NewRepository(db)),
		contentService:   contentmodule.NewService(contentmodule.NewRepository(db)),
	}
}

func assembleModuleDependencies(
	cfg config.Config,
	db *gorm.DB,
	obsRecorder *observability.Recorder,
) (apphttp.ModuleDependencies, error) {
	aiService, err := assembleAIService(cfg, db, obsRecorder)
	if err != nil {
		return apphttp.ModuleDependencies{}, err
	}

	identity := assembleIdentityServices(cfg, db, aiService)
	recovery := assembleRecoveryServices(db, aiService)
	content := assembleContentServices(db)

	return apphttp.ModuleDependencies{
		AuthService:         identity.authService,
		UsersService:        identity.usersService,
		RoutineService:      recovery.routineService,
		JournalsService:     recovery.journalsService,
		AchievementsService: recovery.achievementsService,
		CommunityService:    content.communityService,
		EducationService:    content.educationService,
		ContentService:      content.contentService,
		AIService:           aiService,
	}, nil
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

	moduleDeps, err := assembleModuleDependencies(cfg, dbClient.Gorm(), obsRecorder)
	if err != nil {
		_ = dbClient.Close()
		return nil, err
	}

	server, err := apphttp.NewServer(cfg, logger, apphttp.WithReadinessChecks([]apphttp.ReadinessCheck{
		{
			Name:    "database",
			Mode:    apphttp.ReadinessModeRequired,
			Message: "Koneksi database tidak sehat",
			Probe:   dbClient.Ping,
		},
	}), apphttp.WithModuleDependencies(moduleDeps), apphttp.WithObservability(obsRecorder))
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
