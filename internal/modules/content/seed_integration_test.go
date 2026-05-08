package content

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/recova-app/backend-v2/internal/platform/config"
	"github.com/recova-app/backend-v2/internal/platform/database"
	databaseharness "github.com/recova-app/backend-v2/test/harness/database"
)

func TestIntegration_SeedSQL_Idempotent(t *testing.T) {
	databaseURL := databaseharness.RequireDatabaseURLFromEnv(t, "RECOVA_DB_INTEGRATION_URL")
	databaseharness.ResetMigrations(t, databaseURL)

	client, err := database.Connect(integrationConfigContent(databaseURL))
	if err != nil {
		t.Fatalf("connect db: %v", err)
	}
	t.Cleanup(func() {
		_ = client.Close()
	})

	seedPath := filepath.Join(databaseharness.ProjectRoot(t), "migrations", "seeds", "000001_baseline_seed.sql")
	seedSQL, err := os.ReadFile(seedPath)
	if err != nil {
		t.Fatalf("read seed sql: %v", err)
	}

	ctx := context.Background()
	if err := client.Gorm().WithContext(ctx).Exec(string(seedSQL)).Error; err != nil {
		t.Fatalf("run first seed: %v", err)
	}
	firstEducation := countRows(t, client, "education_contents")
	firstMotivation := countRows(t, client, "daily_motivations")
	firstChallenge := countRows(t, client, "daily_challenges")

	if err := client.Gorm().WithContext(ctx).Exec(string(seedSQL)).Error; err != nil {
		t.Fatalf("run second seed: %v", err)
	}
	secondEducation := countRows(t, client, "education_contents")
	secondMotivation := countRows(t, client, "daily_motivations")
	secondChallenge := countRows(t, client, "daily_challenges")

	if firstEducation != secondEducation || firstMotivation != secondMotivation || firstChallenge != secondChallenge {
		t.Fatalf("seed is not idempotent: first=(%d,%d,%d) second=(%d,%d,%d)", firstEducation, firstMotivation, firstChallenge, secondEducation, secondMotivation, secondChallenge)
	}
}

func integrationConfigContent(databaseURL string) config.Config {
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

func countRows(t testing.TB, client *database.Client, tableName string) int64 {
	t.Helper()
	var count int64
	if err := client.Gorm().WithContext(context.Background()).Table(tableName).Count(&count).Error; err != nil {
		t.Fatalf("count %s: %v", tableName, err)
	}
	return count
}
