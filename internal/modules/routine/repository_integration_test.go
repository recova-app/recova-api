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
		GoogleID: models.StringPtr("google-routine-it-1"),
		Email:    "routine-it@example.test",
		Nickname: "routine-it",
	}
	if err := client.Gorm().WithContext(ctx).Create(&user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}

	check_in_date := time.Date(2026, 5, 8, 0, 0, 0, 0, time.UTC)
	err = repo.CreateCheckIn(ctx, models.CheckIn{
		UserID:       user.ID,
		CheckInDate:  check_in_date,
		Mood:         "fokus",
		IsSuccessful: true,
	})
	if err != nil {
		t.Fatalf("create first checkin: %v", err)
	}

	err = repo.CreateCheckIn(ctx, models.CheckIn{
		UserID:       user.ID,
		CheckInDate:  check_in_date,
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

func TestIntegration_Repository_ListActivityWithinRange(t *testing.T) {
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
		GoogleID: models.StringPtr("google-routine-it-2"),
		Email:    "routine-it-2@example.test",
		Nickname: "routine-it-2",
	}
	if err := client.Gorm().WithContext(ctx).Create(&user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}

	checkIn1 := models.CheckIn{
		UserID:       user.ID,
		CheckInDate:  time.Date(2026, 5, 7, 0, 0, 0, 0, time.UTC),
		Mood:         "fokus",
		IsSuccessful: true,
		CreatedAt:    time.Date(2026, 5, 7, 9, 0, 0, 0, time.UTC),
	}
	checkIn2 := models.CheckIn{
		UserID:       user.ID,
		CheckInDate:  time.Date(2026, 5, 8, 0, 0, 0, 0, time.UTC),
		Mood:         "cemas",
		IsSuccessful: false,
		CreatedAt:    time.Date(2026, 5, 8, 9, 0, 0, 0, time.UTC),
	}
	if err := repo.CreateCheckIn(ctx, checkIn1); err != nil {
		t.Fatalf("create checkin 1: %v", err)
	}
	if err := repo.CreateCheckIn(ctx, checkIn2); err != nil {
		t.Fatalf("create checkin 2: %v", err)
	}

	storedCheckIn2, err := repo.FindCheckInByUserAndDate(ctx, user.ID, checkIn2.CheckInDate)
	if err != nil {
		t.Fatalf("find checkin 2: %v", err)
	}
	if err := repo.CreateJournal(ctx, models.Journal{
		UserID:    user.ID,
		CheckInID: &storedCheckIn2.ID,
		Content:   "catatan aktivitas",
		CreatedAt: time.Date(2026, 5, 8, 10, 0, 0, 0, time.UTC),
	}); err != nil {
		t.Fatalf("create journal: %v", err)
	}

	rows, err := repo.ListCheckInsByUserWithinDateRange(
		ctx,
		user.ID,
		time.Date(2026, 5, 8, 0, 0, 0, 0, time.UTC),
		time.Date(2026, 5, 8, 0, 0, 0, 0, time.UTC),
	)
	if err != nil {
		t.Fatalf("list checkins by range: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("expected 1 checkin in range, got %d", len(rows))
	}
	if rows[0].Mood != "cemas" {
		t.Fatalf("expected mood cemas, got %s", rows[0].Mood)
	}

	journals, err := repo.ListJournalsByUserWithinTimeRange(
		ctx,
		user.ID,
		time.Date(2026, 5, 8, 0, 0, 0, 0, time.UTC),
		time.Date(2026, 5, 9, 0, 0, 0, 0, time.UTC),
	)
	if err != nil {
		t.Fatalf("list journals by range: %v", err)
	}
	if len(journals) != 1 {
		t.Fatalf("expected 1 journal in range, got %d", len(journals))
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
