package auth

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/recova-app/backend-v2/internal/platform/config"
	"github.com/recova-app/backend-v2/internal/platform/database"
	"github.com/recova-app/backend-v2/internal/platform/database/models"
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

func TestIntegration_Repository_RotateRefreshToken_Atomic(t *testing.T) {
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

	user, err := repo.CreateManualUser(ctx, "session-tester@example.test", "sessiontester", "sessiontester", passwordHash)
	if err != nil {
		t.Fatalf("create user: %v", err)
	}

	initialToken := models.AuthRefreshToken{
		UserID:    user.ID,
		TokenHash: "initial-token-hash",
		ExpiresAt: time.Now().Add(time.Hour),
	}
	if err := repo.CreateRefreshToken(ctx, initialToken); err != nil {
		t.Fatalf("create refresh token: %v", err)
	}

	active, err := repo.GetActiveRefreshTokenByHash(ctx, "initial-token-hash")
	if err != nil {
		t.Fatalf("get active refresh token: %v", err)
	}

	now := time.Now().UTC()
	rotatedFromID := active.ID
	newToken1 := models.AuthRefreshToken{
		UserID:        user.ID,
		TokenHash:     "rotated-token-hash-1",
		ExpiresAt:     now.Add(time.Hour),
		RotatedFromID: &rotatedFromID,
	}

	// First rotation succeeds
	if err := repo.RotateRefreshToken(ctx, active.ID, now, newToken1); err != nil {
		t.Fatalf("first rotate refresh token failed: %v", err)
	}

	// Second rotation (replay attempt) with same old token ID must fail with ErrRefreshTokenAlreadyRevoked
	newToken2 := models.AuthRefreshToken{
		UserID:        user.ID,
		TokenHash:     "replay-token-hash-2",
		ExpiresAt:     now.Add(time.Hour),
		RotatedFromID: &rotatedFromID,
	}
	err = repo.RotateRefreshToken(ctx, active.ID, now, newToken2)
	if !errors.Is(err, ErrRefreshTokenAlreadyRevoked) {
		t.Fatalf("expected ErrRefreshTokenAlreadyRevoked on replay, got: %v", err)
	}

	// newToken2 should NOT exist as active token
	if _, err := repo.GetActiveRefreshTokenByHash(ctx, "replay-token-hash-2"); !IsRecordNotFound(err) {
		t.Fatalf("replay token should not exist, got: %v", err)
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
