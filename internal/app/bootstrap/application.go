// Package bootstrap assembles runtime dependencies for the application.
package bootstrap

import (
	"context"
	"log/slog"

	apphttp "github.com/recova-app/backend-v2/internal/app/http"
	"github.com/recova-app/backend-v2/internal/platform/config"
	"github.com/recova-app/backend-v2/internal/platform/database"
)

// Application coordinates the top-level runtime process for the API service.
type Application struct {
	server   *apphttp.Server
	logger   *slog.Logger
	database *database.Client
}

// NewApplication constructs an executable application instance from runtime dependencies.
func NewApplication(cfg config.Config, logger *slog.Logger) (*Application, error) {
	dbClient, err := database.Connect(cfg)
	if err != nil {
		return nil, err
	}

	server, err := apphttp.NewServer(cfg, logger, apphttp.WithReadinessChecks([]apphttp.ReadinessCheck{
		{
			Name:    "database",
			Mode:    apphttp.ReadinessModeRequired,
			Message: "Koneksi database tidak sehat",
			Probe:   dbClient.Ping,
		},
	}))
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
			a.logger.Error("gagal menutup koneksi database", "error", err)
		}
	}()

	a.logger.Info("menjalankan aplikasi", "app", "api")
	return a.server.Start(ctx)
}
