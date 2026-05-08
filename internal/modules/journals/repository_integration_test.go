package journals

import (
	"context"
	"testing"
	"time"

	"github.com/recova-app/backend-v2/internal/platform/config"
	"github.com/recova-app/backend-v2/internal/platform/database"
	"github.com/recova-app/backend-v2/internal/platform/database/models"
	databaseharness "github.com/recova-app/backend-v2/test/harness/database"
)

func TestIntegration_Repository_ListJournals_UserScoped(t *testing.T) {
	databaseURL := databaseharness.RequireDatabaseURLFromEnv(t, "RECOVA_DB_INTEGRATION_URL")
	databaseharness.ResetMigrations(t, databaseURL)

	client, err := database.Connect(integrationConfigJournals(databaseURL))
	if err != nil {
		t.Fatalf("connect db: %v", err)
	}
	t.Cleanup(func() {
		_ = client.Close()
	})

	repo := NewRepository(client.Gorm())
	ctx := context.Background()

	userA := models.User{
		GoogleID: "google-journal-it-a",
		Email:    "journal-a@example.test",
		Nickname: "journal-a",
	}
	userB := models.User{
		GoogleID: "google-journal-it-b",
		Email:    "journal-b@example.test",
		Nickname: "journal-b",
	}
	if err := client.Gorm().WithContext(ctx).Create(&userA).Error; err != nil {
		t.Fatalf("create userA: %v", err)
	}
	if err := client.Gorm().WithContext(ctx).Create(&userB).Error; err != nil {
		t.Fatalf("create userB: %v", err)
	}

	if _, err := repo.CreateJournal(ctx, userA.ID, "catatan a1"); err != nil {
		t.Fatalf("create journal a1: %v", err)
	}
	if _, err := repo.CreateJournal(ctx, userA.ID, "catatan a2"); err != nil {
		t.Fatalf("create journal a2: %v", err)
	}
	if _, err := repo.CreateJournal(ctx, userB.ID, "catatan b1"); err != nil {
		t.Fatalf("create journal b1: %v", err)
	}

	rows, err := repo.ListJournalsByUserID(ctx, userA.ID)
	if err != nil {
		t.Fatalf("list journals userA: %v", err)
	}
	if len(rows) != 2 {
		t.Fatalf("expected 2 journals for userA, got %d", len(rows))
	}
	for _, row := range rows {
		if row.UserID != userA.ID {
			t.Fatalf("expected only userA journals, got user_id=%s", row.UserID)
		}
		if row.CreatedAt.After(time.Now().Add(1 * time.Minute)) {
			t.Fatalf("unexpected created_at: %s", row.CreatedAt)
		}
	}
}

func integrationConfigJournals(databaseURL string) config.Config {
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
