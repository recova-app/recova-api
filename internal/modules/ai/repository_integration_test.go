package ai

import (
	"context"
	"testing"
	"time"

	"github.com/recova-app/backend-v2/internal/platform/config"
	"github.com/recova-app/backend-v2/internal/platform/database"
	"github.com/recova-app/backend-v2/internal/platform/database/models"
	databaseharness "github.com/recova-app/backend-v2/test/harness/database"
)

func TestIntegration_Repository_ListRecentChats_UserScopedAndAscending(t *testing.T) {
	databaseURL := databaseharness.RequireDatabaseURLFromEnv(t, "RECOVA_DB_INTEGRATION_URL")
	databaseharness.ResetMigrations(t, databaseURL)

	client, err := database.Connect(integrationConfigAI(databaseURL))
	if err != nil {
		t.Fatalf("connect db: %v", err)
	}
	t.Cleanup(func() {
		_ = client.Close()
	})

	repo := NewRepository(client.Gorm())
	ctx := context.Background()

	userA := models.User{GoogleID: "google-ai-it-a", Email: "ai-a@example.test", Nickname: "ai-a"}
	userB := models.User{GoogleID: "google-ai-it-b", Email: "ai-b@example.test", Nickname: "ai-b"}
	if err := client.Gorm().WithContext(ctx).Create(&userA).Error; err != nil {
		t.Fatalf("create userA: %v", err)
	}
	if err := client.Gorm().WithContext(ctx).Create(&userB).Error; err != nil {
		t.Fatalf("create userB: %v", err)
	}

	baseTime := time.Date(2026, 5, 8, 10, 0, 0, 0, time.UTC)
	if err := repo.CreateChatMessages(ctx, []models.AIChat{
		{UserID: userA.ID, Role: "user", Content: "pesan-a1", CreatedAt: baseTime.Add(1 * time.Minute)},
		{UserID: userA.ID, Role: "model", Content: "pesan-a2", CreatedAt: baseTime.Add(2 * time.Minute)},
		{UserID: userB.ID, Role: "user", Content: "pesan-b1", CreatedAt: baseTime.Add(3 * time.Minute)},
	}); err != nil {
		t.Fatalf("create chat rows: %v", err)
	}

	rows, err := repo.ListRecentChatsByUserID(ctx, userA.ID, 10)
	if err != nil {
		t.Fatalf("list recent chats: %v", err)
	}
	if len(rows) != 2 {
		t.Fatalf("expected 2 rows for userA, got %d", len(rows))
	}
	if rows[0].UserID != userA.ID || rows[1].UserID != userA.ID {
		t.Fatalf("expected rows scoped to userA, got %+v", rows)
	}
	if rows[0].CreatedAt.After(rows[1].CreatedAt) {
		t.Fatalf("expected ascending createdAt order, got %s then %s", rows[0].CreatedAt, rows[1].CreatedAt)
	}
}

func TestIntegration_Repository_PersonaPreferenceUpsertAndRead(t *testing.T) {
	databaseURL := databaseharness.RequireDatabaseURLFromEnv(t, "RECOVA_DB_INTEGRATION_URL")
	databaseharness.ResetMigrations(t, databaseURL)

	client, err := database.Connect(integrationConfigAI(databaseURL))
	if err != nil {
		t.Fatalf("connect db: %v", err)
	}
	t.Cleanup(func() {
		_ = client.Close()
	})

	repo := NewRepository(client.Gorm())
	ctx := context.Background()

	user := models.User{GoogleID: "google-ai-it-persona", Email: "ai-persona@example.test", Nickname: "ai-persona"}
	if err := client.Gorm().WithContext(ctx).Create(&user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}

	firstUpdate := time.Date(2026, 5, 9, 10, 0, 0, 0, time.UTC)
	if err := repo.UpsertPersonaPreference(ctx, user.ID, "friendly", firstUpdate); err != nil {
		t.Fatalf("first upsert persona: %v", err)
	}

	secondUpdate := firstUpdate.Add(5 * time.Minute)
	if err := repo.UpsertPersonaPreference(ctx, user.ID, "direct", secondUpdate); err != nil {
		t.Fatalf("second upsert persona: %v", err)
	}

	row, err := repo.GetPersonaPreferenceByUserID(ctx, user.ID)
	if err != nil {
		t.Fatalf("get persona preference: %v", err)
	}
	if row.Persona != "direct" {
		t.Fatalf("expected persona direct, got %q", row.Persona)
	}
	if !row.UpdatedAt.UTC().Equal(secondUpdate.UTC()) {
		t.Fatalf("expected updated_at=%s, got %s", secondUpdate.UTC(), row.UpdatedAt.UTC())
	}
}

func integrationConfigAI(databaseURL string) config.Config {
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
