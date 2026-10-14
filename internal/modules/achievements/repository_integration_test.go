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
		GoogleID: models.StringPtr("google-achievements-it-1"),
		Email:    "achievements-it-1@example.test",
		Nickname: "achievements-it-1",
	}
	if err := client.Gorm().WithContext(ctx).Create(&user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}

	achievement := models.Achievement{
		Code:        "streak_7_days",
		Title:       "7 Days",
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
		GoogleID: models.StringPtr("google-achievements-it-2"),
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

func TestIntegration_Repository_ComputeEvaluationMetrics_Batched(t *testing.T) {
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
		GoogleID: models.StringPtr("google-achievements-eval"),
		Email:    "achievements-eval@example.test",
		Nickname: "eval-user",
	}
	if err := client.Gorm().WithContext(ctx).Create(&user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}

	now := time.Date(2026, 5, 10, 12, 0, 0, 0, time.UTC)

	// 1. Zero state
	metrics, err := repo.ComputeEvaluationMetrics(ctx, user.ID, now)
	if err != nil {
		t.Fatalf("compute evaluation metrics: %v", err)
	}
	if metrics.StreakMilestone != 0 {
		t.Fatalf("expected 0 streak, got %f", metrics.StreakMilestone)
	}
	if metrics.OnboardingCompletion != 0 {
		t.Fatalf("expected 0 onboarding, got %f", metrics.OnboardingCompletion)
	}

	// 2. Populated activity state
	// Profile
	profile := models.Profile{UserID: user.ID, Answers: []byte(`{"q":"a"}`)}
	if err := client.Gorm().WithContext(ctx).Create(&profile).Error; err != nil {
		t.Fatalf("create profile: %v", err)
	}

	// Streak (started 4 days ago -> 5 days streak inclusive)
	streak := models.Streak{UserID: user.ID, StartDate: now.AddDate(0, 0, -4), IsActive: true}
	if err := client.Gorm().WithContext(ctx).Create(&streak).Error; err != nil {
		t.Fatalf("create streak: %v", err)
	}

	// Relapse 2 days ago, then success 1 day ago
	relapseDate := now.AddDate(0, 0, -2)
	successDate := now.AddDate(0, 0, -1)
	if err := client.Gorm().WithContext(ctx).Create(&models.CheckIn{
		UserID: user.ID, CheckInDate: relapseDate, Mood: "bad", IsSuccessful: false,
	}).Error; err != nil {
		t.Fatalf("create relapse checkin: %v", err)
	}
	if err := client.Gorm().WithContext(ctx).Create(&models.CheckIn{
		UserID: user.ID, CheckInDate: successDate, Mood: "good", IsSuccessful: true,
	}).Error; err != nil {
		t.Fatalf("create success checkin: %v", err)
	}

	// Journal
	if err := client.Gorm().WithContext(ctx).Create(&models.Journal{
		UserID: user.ID, Content: "today was great", CreatedAt: now.Add(-2 * time.Hour),
	}).Error; err != nil {
		t.Fatalf("create journal: %v", err)
	}

	// Community: 1 post, 1 comment, 1 like = 3
	post := models.CommunityPost{UserID: user.ID, Content: "community test content", Category: "motivasi"}
	if err := client.Gorm().WithContext(ctx).Create(&post).Error; err != nil {
		t.Fatalf("create post: %v", err)
	}
	comment := models.CommunityComment{UserID: user.ID, PostID: post.ID, Content: "comment content"}
	if err := client.Gorm().WithContext(ctx).Create(&comment).Error; err != nil {
		t.Fatalf("create comment: %v", err)
	}
	like := models.CommunityPostLike{UserID: user.ID, PostID: post.ID}
	if err := client.Gorm().WithContext(ctx).Create(&like).Error; err != nil {
		t.Fatalf("create like: %v", err)
	}

	// Evaluate populated metrics
	populated, err := repo.ComputeEvaluationMetrics(ctx, user.ID, now)
	if err != nil {
		t.Fatalf("compute populated metrics: %v", err)
	}
	if populated.OnboardingCompletion != 1 {
		t.Fatalf("expected onboarding 1, got %f", populated.OnboardingCompletion)
	}
	if populated.StreakMilestone < 4 {
		t.Fatalf("expected streak >= 4, got %f", populated.StreakMilestone)
	}
	if populated.CheckinConsistency != 1 {
		t.Fatalf("expected checkin consistency 1 (successful checkin), got %f", populated.CheckinConsistency)
	}
	if populated.JournalConsistency != 1 {
		t.Fatalf("expected journal consistency 1, got %f", populated.JournalConsistency)
	}
	if populated.RelapseRecovery != 1 {
		t.Fatalf("expected relapse recovery 1 (success after relapse), got %f", populated.RelapseRecovery)
	}
	if populated.CommunityParticipation != 3 {
		t.Fatalf("expected community participation 3 (1 post + 1 comment + 1 like), got %f", populated.CommunityParticipation)
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
