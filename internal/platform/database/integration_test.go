package database

import (
	"errors"
	"testing"

	"github.com/recova-app/backend-v2/internal/platform/config"
	databaseharness "github.com/recova-app/backend-v2/test/harness/database"
	"gorm.io/gorm"
)

func TestIntegration_ConnectAndMigrationRoundTrip(t *testing.T) {
	databaseURL := databaseharness.RequireDatabaseURLFromEnv(t, "RECOVA_DB_INTEGRATION_URL")

	cfg := config.Config{
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

	client, err := Connect(cfg)
	if err != nil {
		t.Fatalf("connect error: %v", err)
	}
	t.Cleanup(func() {
		_ = client.Close()
	})

	if err := client.Ping(nil); err != nil {
		t.Fatalf("ping error: %v", err)
	}

	databaseharness.ResetMigrations(t, databaseURL)

	if err := assertTableExists(client.Gorm(), "users"); err != nil {
		t.Fatalf("expected users table exists after migration reset: %v", err)
	}
}

func assertTableExists(gormDB *gorm.DB, tableName string) error {
	const query = `
SELECT EXISTS (
	SELECT 1
	FROM information_schema.tables
	WHERE table_schema = 'public' AND table_name = $1
)`

	var exists bool
	if err := gormDB.Raw(query, tableName).Scan(&exists).Error; err != nil {
		return err
	}

	if !exists {
		return errors.New("table not found")
	}

	return nil
}
