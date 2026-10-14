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

	userA := models.User{GoogleID: models.StringPtr("google-ai-it-a"), Email: "ai-a@example.test", Nickname: "ai-a"}
	userB := models.User{GoogleID: models.StringPtr("google-ai-it-b"), Email: "ai-b@example.test", Nickname: "ai-b"}
	if err := client.Gorm().WithContext(ctx).Create(&userA).Error; err != nil {
		t.Fatalf("create userA: %v", err)
	}
	if err := client.Gorm().WithContext(ctx).Create(&userB).Error; err != nil {
		t.Fatalf("create userB: %v", err)
	}

	baseTime := time.Date(2026, 5, 8, 10, 0, 0, 0, time.UTC)
	if err := repo.CreateChatMessages(ctx, []models.AIChat{
		{UserID: userA.ID, Role: "user", Content: "message-a1", CreatedAt: baseTime.Add(1 * time.Minute)},
		{UserID: userA.ID, Role: "model", Content: "message-a2", CreatedAt: baseTime.Add(2 * time.Minute)},
		{UserID: userB.ID, Role: "user", Content: "message-b1", CreatedAt: baseTime.Add(3 * time.Minute)},
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
		t.Fatalf("expected ascending created_at order, got %s then %s", rows[0].CreatedAt, rows[1].CreatedAt)
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

	user := models.User{GoogleID: models.StringPtr("google-ai-it-persona"), Email: "ai-persona@example.test", Nickname: "ai-persona"}
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

func TestIntegration_Repository_GetCoachContext_Consolidated(t *testing.T) {
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

	// 1. Non-existent user
	_, err = repo.GetCoachContext(ctx, "00000000-0000-0000-0000-000000000099", 10)
	if !IsRecordNotFound(err) {
		t.Fatalf("expected not found for non-existent user, got: %v", err)
	}

	// 2. User with full metadata (profile, streak, persona, chats)
	userWhy := "ingin hidup lebih tenang"
	user := models.User{
		GoogleID: models.StringPtr("google-ai-full"),
		Email:    "ai-full@example.test",
		Nickname: "Budi",
		UserWhy:  &userWhy,
	}
	if err := client.Gorm().WithContext(ctx).Create(&user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}

	profile := models.Profile{
		UserID:          user.ID,
		Answers:         []byte(`{"goal":"recovery"}`),
		DependencyLevel: models.StringPtr("sedang"),
		AISummary:       models.StringPtr("Semangat terus"),
	}
	if err := client.Gorm().WithContext(ctx).Create(&profile).Error; err != nil {
		t.Fatalf("create profile: %v", err)
	}

	baseDate := time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)
	streak := models.Streak{
		UserID:    user.ID,
		StartDate: baseDate,
		IsActive:  true,
	}
	if err := client.Gorm().WithContext(ctx).Create(&streak).Error; err != nil {
		t.Fatalf("create streak: %v", err)
	}

	if err := repo.UpsertPersonaPreference(ctx, user.ID, "friendly", time.Now().UTC()); err != nil {
		t.Fatalf("upsert persona: %v", err)
	}

	if err := repo.CreateChatMessages(ctx, []models.AIChat{
		{UserID: user.ID, Role: "user", Content: "Halo coach", CreatedAt: time.Now().UTC().Add(-1 * time.Minute)},
		{UserID: user.ID, Role: "assistant", Content: "Halo Budi!", CreatedAt: time.Now().UTC()},
	}); err != nil {
		t.Fatalf("create chats: %v", err)
	}

	coachCtx, err := repo.GetCoachContext(ctx, user.ID, 10)
	if err != nil {
		t.Fatalf("get coach context: %v", err)
	}
	if coachCtx.User.Nickname != "Budi" {
		t.Fatalf("expected nickname Budi, got %s", coachCtx.User.Nickname)
	}
	if !coachCtx.HasProfile || coachCtx.Profile.DependencyLevel == nil || *coachCtx.Profile.DependencyLevel != "sedang" {
		t.Fatalf("expected profile with dependency level sedang, got %+v", coachCtx.Profile)
	}
	if coachCtx.ActiveStreak == nil || !coachCtx.ActiveStreak.StartDate.Equal(baseDate) {
		t.Fatalf("expected active streak starting %v, got %+v", baseDate, coachCtx.ActiveStreak)
	}
	if coachCtx.Persona != "friendly" {
		t.Fatalf("expected persona friendly, got %s", coachCtx.Persona)
	}
	if len(coachCtx.RecentChats) != 2 {
		t.Fatalf("expected 2 recent chats, got %d", len(coachCtx.RecentChats))
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
