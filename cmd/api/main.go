// Package main starts the Recova Backend API process.
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/recova-app/backend-v2/internal/app/bootstrap"
	"github.com/recova-app/backend-v2/internal/app/lifecycle"
	"github.com/recova-app/backend-v2/internal/platform/config"
	platformlogger "github.com/recova-app/backend-v2/internal/platform/logger"
)

// main loads runtime dependencies and runs the API process lifecycle.
func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "config error: %v\n", err)
		os.Exit(1)
	}

	logger := platformlogger.New(cfg)
	logger.Info("configuration loaded", "config", cfg.RedactedSummary())
	app, err := bootstrap.NewApplication(cfg, logger)
	if err != nil {
		logger.Error("application bootstrap failed", "error", err)
		os.Exit(1)
	}

	ctx, stop := lifecycle.WithShutdownSignal(context.Background())
	defer stop()

	if err := app.Run(ctx); err != nil {
		logger.Error("application stopped with error", "error", err)
		os.Exit(1)
	}
}
