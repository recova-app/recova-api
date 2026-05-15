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
		Type:     "artikel",
		IsActive: true,
	}).Error; err != nil {
		t.Fatalf("create active content: %v", err)
	}
	if err := client.Gorm().WithContext(ctx).Create(&models.EducationContent{
		Title:    "inactive",
		URL:      "https://example.test/inactive",
		Category: "mindset",
		Type:     "video",
		IsActive: false,
	}).Error; err != nil {
		t.Fatalf("create inactive content: %v", err)
	}

	rows, err := repo.ListActiveContents(ctx)
	if err != nil {
		t.Fatalf("list active contents: %v", err)
	}
	found := false
	contentType := ""
	for _, row := range rows {
		if row.Title == "active" {
			found = true
			contentType = row.Type
			break
		}
	}
	if !found {
		t.Fatalf("expected inserted active content present, got rows=%+v", rows)
	}
	if contentType != "artikel" {
		t.Fatalf("expected active content type artikel, got: %q", contentType)
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
