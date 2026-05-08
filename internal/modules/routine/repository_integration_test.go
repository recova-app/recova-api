package routine

import (
	"context"
	"testing"
	"time"

	"github.com/recova-app/backend-v2/internal/platform/config"
	"github.com/recova-app/backend-v2/internal/platform/database"
	"github.com/recova-app/backend-v2/internal/platform/database/models"
	databaseharness "github.com/recova-app/backend-v2/test/harness/database"
)

func TestIntegration_Repository_DuplicateCheckInReturnsUniqueViolation(t *testing.T) {
	databaseURL := databaseharness.RequireDatabaseURLFromEnv(t, "RECOVA_DB_INTEGRATION_URL")
	databaseharness.ResetMigrations(t, databaseURL)

	client, err := database.Connect(integrationConfig(databaseURL))
	if err != nil {
		t.Fatalf("connect db: %v", err)
	}
	t.Cleanup(func() {
		_ = client.Close()
	})

	repo := NewRepository(client.Gorm())
	ctx := context.Background()

	user := models.User{
		GoogleID: "google-routine-it-1",
		Email:    "routine-it@example.test",
		Nickname: "routine-it",
	}
	if err := client.Gorm().WithContext(ctx).Create(&user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}

	checkInDate := time.Date(2026, 5, 8, 0, 0, 0, 0, time.UTC)
	err = repo.CreateCheckIn(ctx, models.CheckIn{
		UserID:       user.ID,
		CheckInDate:  checkInDate,
		Mood:         "fokus",
		IsSuccessful: true,
	})
	if err != nil {
		t.Fatalf("create first checkin: %v", err)
	}

	err = repo.CreateCheckIn(ctx, models.CheckIn{
		UserID:       user.ID,
		CheckInDate:  checkInDate,
		Mood:         "fokus",
		IsSuccessful: false,
	})
	if err == nil {
		t.Fatal("expected duplicate check-in unique violation")
	}
	if !IsUniqueViolation(err) {
		t.Fatalf("expected unique violation error, got: %v", err)
	}
}

func integrationConfig(databaseURL string) config.Config {
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
