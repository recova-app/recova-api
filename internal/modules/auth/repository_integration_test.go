package auth

import (
	"context"
	"testing"

	"github.com/recova-app/backend-v2/internal/platform/config"
	"github.com/recova-app/backend-v2/internal/platform/database"
	databaseharness "github.com/recova-app/backend-v2/test/harness/database"
)

func TestIntegration_Repository_ManualAuthUniqueAndLookup(t *testing.T) {
	databaseURL := databaseharness.RequireDatabaseURLFromEnv(t, "RECOVA_DB_INTEGRATION_URL")
	databaseharness.ResetMigrations(t, databaseURL)

	client, err := database.Connect(integrationConfigAuth(databaseURL))
	if err != nil {
		t.Fatalf("connect db: %v", err)
	}
	t.Cleanup(func() {
		_ = client.Close()
	})

	repo := NewRepository(client.Gorm())
	ctx := context.Background()

	passwordHash, err := hashPassword("password123")
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}

	created, err := repo.CreateManualUser(ctx, "manual@example.test", "manualuser", "manualuser", passwordHash)
	if err != nil {
		t.Fatalf("create manual user: %v", err)
	}
	if created.PasswordHash == nil || *created.PasswordHash != passwordHash {
		t.Fatalf("unexpected password hash stored: %#v", created.PasswordHash)
	}

	byEmail, err := repo.FindUserByLoginIdentifier(ctx, "manual@example.test")
	if err != nil {
		t.Fatalf("find by email: %v", err)
	}
	if byEmail.ID != created.ID {
		t.Fatalf("unexpected user by email: %s", byEmail.ID)
	}

	byUsername, err := repo.FindUserByLoginIdentifier(ctx, "manualuser")
	if err != nil {
		t.Fatalf("find by username: %v", err)
	}
	if byUsername.ID != created.ID {
		t.Fatalf("unexpected user by username: %s", byUsername.ID)
	}

	_, err = repo.CreateManualUser(ctx, "manual@example.test", "manualuser2", "manualuser2", passwordHash)
	if err == nil || !IsUniqueViolation(err) {
		t.Fatalf("expected unique violation for duplicate email, got: %v", err)
	}

	_, err = repo.CreateManualUser(ctx, "manual2@example.test", "manualuser", "manualuser", passwordHash)
	if err == nil || !IsUniqueViolation(err) {
		t.Fatalf("expected unique violation for duplicate username, got: %v", err)
	}
}

func integrationConfigAuth(databaseURL string) config.Config {
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
