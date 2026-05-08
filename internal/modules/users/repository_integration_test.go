package users

import (
	"context"
	"testing"

	"github.com/recova-app/backend-v2/internal/platform/config"
	"github.com/recova-app/backend-v2/internal/platform/database"
	"github.com/recova-app/backend-v2/internal/platform/database/models"
	databaseharness "github.com/recova-app/backend-v2/test/harness/database"
)

func TestIntegration_Repository_ResetUserDataForTesting_RemovesAIChatsScopedUser(t *testing.T) {
	databaseURL := databaseharness.RequireDatabaseURLFromEnv(t, "RECOVA_DB_INTEGRATION_URL")
	databaseharness.ResetMigrations(t, databaseURL)

	client, err := database.Connect(integrationConfigUsers(databaseURL))
	if err != nil {
		t.Fatalf("connect db: %v", err)
	}
	t.Cleanup(func() {
		_ = client.Close()
	})

	repo := NewRepository(client.Gorm())
	ctx := context.Background()

	userA := models.User{GoogleID: "google-reset-ai-a", Email: "reset-ai-a@example.test", Nickname: "reset-a"}
	userB := models.User{GoogleID: "google-reset-ai-b", Email: "reset-ai-b@example.test", Nickname: "reset-b"}
	if err := client.Gorm().WithContext(ctx).Create(&userA).Error; err != nil {
		t.Fatalf("create userA: %v", err)
	}
	if err := client.Gorm().WithContext(ctx).Create(&userB).Error; err != nil {
		t.Fatalf("create userB: %v", err)
	}

	rows := []models.AIChat{
		{UserID: userA.ID, Role: "user", Content: "pesan a"},
		{UserID: userB.ID, Role: "user", Content: "pesan b"},
	}
	if err := client.Gorm().WithContext(ctx).Create(&rows).Error; err != nil {
		t.Fatalf("create ai chats: %v", err)
	}

	if err := repo.ResetUserDataForTesting(ctx, userA.ID); err != nil {
		t.Fatalf("reset user data: %v", err)
	}

	var countUserA int64
	if err := client.Gorm().WithContext(ctx).
		Model(&models.AIChat{}).
		Where("user_id = ?", userA.ID).
		Count(&countUserA).Error; err != nil {
		t.Fatalf("count userA ai chats: %v", err)
	}
	if countUserA != 0 {
		t.Fatalf("expected userA ai chats removed, got %d", countUserA)
	}

	var countUserB int64
	if err := client.Gorm().WithContext(ctx).
		Model(&models.AIChat{}).
		Where("user_id = ?", userB.ID).
		Count(&countUserB).Error; err != nil {
		t.Fatalf("count userB ai chats: %v", err)
	}
	if countUserB != 1 {
		t.Fatalf("expected userB ai chats kept, got %d", countUserB)
	}
}

func integrationConfigUsers(databaseURL string) config.Config {
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
