package achievements

import (
	"context"
	"testing"
	"time"

	"github.com/recova-app/backend-v2/internal/platform/config"
	"github.com/recova-app/backend-v2/internal/platform/database"
	"github.com/recova-app/backend-v2/internal/platform/database/models"
	databaseharness "github.com/recova-app/backend-v2/test/harness/database"
)

func TestIntegration_Repository_UpsertProgressIdempotentNoDoubleUnlock(t *testing.T) {
	databaseURL := databaseharness.RequireDatabaseURLFromEnv(t, "RECOVA_DB_INTEGRATION_URL")
	databaseharness.ResetMigrations(t, databaseURL)

	client, err := database.Connect(integrationConfigAchievements(databaseURL))
	if err != nil {
		t.Fatalf("connect db: %v", err)
	}
	t.Cleanup(func() {
		_ = client.Close()
	})

	repo := NewRepository(client.Gorm())
	ctx := context.Background()

	user := models.User{
		GoogleID: "google-achievements-it-1",
		Email:    "achievements-it-1@example.test",
		Nickname: "achievements-it-1",
	}
	if err := client.Gorm().WithContext(ctx).Create(&user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}

	achievement := models.Achievement{
		Code:        "streak_7_days",
		Title:       "7 Hari",
		Description: "desc",
		Category:    categoryStreakMilestone,
		Threshold:   7,
		IsActive:    true,
	}
	if err := client.Gorm().WithContext(ctx).Create(&achievement).Error; err != nil {
		t.Fatalf("create achievement: %v", err)
	}

	firstEval := time.Date(2026, 5, 9, 7, 0, 0, 0, time.UTC)
	if err := repo.UpsertProgress(ctx, user.ID, []progressUpsert{{
		AchievementID: achievement.ID,
		ProgressValue: 2,
		UnlockedAt:    nil,
	}}, firstEval); err != nil {
		t.Fatalf("first upsert: %v", err)
	}

	secondEval := time.Date(2026, 5, 10, 7, 0, 0, 0, time.UTC)
	if err := repo.UpsertProgress(ctx, user.ID, []progressUpsert{{
		AchievementID: achievement.ID,
		ProgressValue: 8,
		UnlockedAt:    &secondEval,
	}}, secondEval); err != nil {
		t.Fatalf("second upsert: %v", err)
	}

	thirdEval := time.Date(2026, 5, 11, 7, 0, 0, 0, time.UTC)
	if err := repo.UpsertProgress(ctx, user.ID, []progressUpsert{{
		AchievementID: achievement.ID,
		ProgressValue: 6,
		UnlockedAt:    nil,
	}}, thirdEval); err != nil {
		t.Fatalf("third upsert: %v", err)
	}

	var rows []models.UserAchievementProgress
	if err := client.Gorm().WithContext(ctx).
		Where("user_id = ?", user.ID).
		Where("achievement_id = ?", achievement.ID).
		Find(&rows).Error; err != nil {
		t.Fatalf("query rows: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("expected one progress row, got %d", len(rows))
	}
	if rows[0].ProgressValue != 8 {
		t.Fatalf("expected progress_value=8, got %v", rows[0].ProgressValue)
	}
	if rows[0].UnlockedAt == nil {
		t.Fatal("expected unlocked_at not nil")
	}
	if !rows[0].UnlockedAt.UTC().Equal(secondEval.UTC()) {
		t.Fatalf("expected unlocked_at=%s, got %s", secondEval.UTC(), rows[0].UnlockedAt.UTC())
	}
}

func TestIntegration_Repository_UniqueConstraintUserAchievementProgress(t *testing.T) {
	databaseURL := databaseharness.RequireDatabaseURLFromEnv(t, "RECOVA_DB_INTEGRATION_URL")
	databaseharness.ResetMigrations(t, databaseURL)

	client, err := database.Connect(integrationConfigAchievements(databaseURL))
	if err != nil {
		t.Fatalf("connect db: %v", err)
	}
	t.Cleanup(func() {
		_ = client.Close()
	})

	ctx := context.Background()
	user := models.User{
		GoogleID: "google-achievements-it-2",
		Email:    "achievements-it-2@example.test",
		Nickname: "achievements-it-2",
	}
	if err := client.Gorm().WithContext(ctx).Create(&user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}

	achievement := models.Achievement{
		Code:        "onboarding_complete",
		Title:       "Onboarding",
		Description: "desc",
		Category:    categoryOnboardingCompletion,
		Threshold:   1,
		IsActive:    true,
	}
	if err := client.Gorm().WithContext(ctx).Create(&achievement).Error; err != nil {
		t.Fatalf("create achievement: %v", err)
	}

	first := models.UserAchievementProgress{
		UserID:        user.ID,
		AchievementID: achievement.ID,
		ProgressValue: 1,
	}
	if err := client.Gorm().WithContext(ctx).Create(&first).Error; err != nil {
		t.Fatalf("create first progress: %v", err)
	}

	second := models.UserAchievementProgress{
		UserID:        user.ID,
		AchievementID: achievement.ID,
		ProgressValue: 1,
	}
	if err := client.Gorm().WithContext(ctx).Create(&second).Error; err == nil {
		t.Fatal("expected unique constraint error for duplicate user-achievement progress")
	}
}

func integrationConfigAchievements(databaseURL string) config.Config {
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
