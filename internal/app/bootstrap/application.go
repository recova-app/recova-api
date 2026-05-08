// Package bootstrap assembles runtime dependencies for the application.
package bootstrap

import (
	"context"
	"log/slog"

	apphttp "github.com/recova-app/backend-v2/internal/app/http"
	"github.com/recova-app/backend-v2/internal/platform/config"
)

// Application coordinates the top-level runtime process for the API service.
type Application struct {
	server *apphttp.Server
	logger *slog.Logger
}

// NewApplication constructs an executable application instance from runtime dependencies.
func NewApplication(cfg config.Config, logger *slog.Logger) (*Application, error) {
	server, err := apphttp.NewServer(cfg, logger)
	if err != nil {
		return nil, err
	}

	return &Application{
		server: server,
		logger: logger,
	}, nil
}

// Run starts the application runtime and blocks until shutdown.
func (a *Application) Run(ctx context.Context) error {
	a.logger.Info("menjalankan aplikasi", "app", "api")
	return a.server.Start(ctx)
}
