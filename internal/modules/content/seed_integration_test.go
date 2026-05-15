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
	firstCounts := captureSeedCounts(t, client)

	if err := client.Gorm().WithContext(ctx).Exec(string(seedSQL)).Error; err != nil {
		t.Fatalf("run second seed: %v", err)
	}
	secondCounts := captureSeedCounts(t, client)

	for table, first := range firstCounts {
		second := secondCounts[table]
		if first != second {
			t.Fatalf("seed is not idempotent for %s: first=%d second=%d", table, first, second)
		}
	}

	minimum := map[string]int64{
		"users":                       6,
		"profiles":                    6,
		"streaks":                     11,
		"check_ins":                   84,
		"journals":                    84,
		"community_posts":             12,
		"community_comments":          25,
		"community_post_likes":        20,
		"education_contents":          23,
		"daily_motivations":           35,
		"daily_challenges":            35,
		"daily_physical_challenges":   20,
		"achievements":                15,
		"user_achievement_progress":   24,
		"user_ai_persona_preferences": 6,
		"ai_chats":                    18,
	}

	for table, minCount := range minimum {
		got := secondCounts[table]
		if got < minCount {
			t.Fatalf("seed minimum size not met for %s: got=%d min=%d", table, got, minCount)
		}
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

func captureSeedCounts(t testing.TB, client *database.Client) map[string]int64 {
	t.Helper()
	tables := []string{
		"users",
		"profiles",
		"streaks",
		"check_ins",
		"journals",
		"community_posts",
		"community_comments",
		"community_post_likes",
		"education_contents",
		"daily_motivations",
		"daily_challenges",
		"daily_physical_challenges",
		"achievements",
		"user_achievement_progress",
		"user_ai_persona_preferences",
		"ai_chats",
	}

	counts := make(map[string]int64, len(tables))
	for _, table := range tables {
		counts[table] = countRows(t, client, table)
	}
	return counts
}
