package education

import (
	"context"
	"testing"

	"github.com/recova-app/backend-v2/internal/platform/config"
	"github.com/recova-app/backend-v2/internal/platform/database"
	"github.com/recova-app/backend-v2/internal/platform/database/models"
	databaseharness "github.com/recova-app/backend-v2/test/harness/database"
)

func TestIntegration_Repository_ListActiveContents(t *testing.T) {
	databaseURL := databaseharness.RequireDatabaseURLFromEnv(t, "RECOVA_DB_INTEGRATION_URL")
	databaseharness.ResetMigrations(t, databaseURL)

	client, err := database.Connect(integrationConfigEducation(databaseURL))
	if err != nil {
		t.Fatalf("connect db: %v", err)
	}
	t.Cleanup(func() {
		_ = client.Close()
	})

	repo := NewRepository(client.Gorm())
	ctx := context.Background()

	if err := client.Gorm().WithContext(ctx).Create(&models.EducationContent{
		Title:    "active",
		URL:      "https://example.test/active",
		Category: "mindset",
		IsActive: true,
	}).Error; err != nil {
		t.Fatalf("create active content: %v", err)
	}
	if err := client.Gorm().WithContext(ctx).Create(&models.EducationContent{
		Title:    "inactive",
		URL:      "https://example.test/inactive",
		Category: "mindset",
		IsActive: false,
	}).Error; err != nil {
		t.Fatalf("create inactive content: %v", err)
	}

	rows, err := repo.ListActiveContents(ctx)
	if err != nil {
		t.Fatalf("list active contents: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("expected 1 active content, got %d", len(rows))
	}
	if rows[0].Title != "active" {
		t.Fatalf("expected active content title, got %s", rows[0].Title)
	}
}

func integrationConfigEducation(databaseURL string) config.Config {
	return config.Config{
		Database: config.DatabaseConfig{
			URL:                databaseURL,
			MaxOpenConns:       10,
			MaxIdleConns:       5,
			ConnMaxLifetimeSec: 300,
		},
		Observability: config.ObservabilityConfig{
			HealthCheckTimeoutMs: 3000,
		},
	}
}
