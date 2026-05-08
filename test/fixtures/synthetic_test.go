package fixtures

import (
	"strings"
	"testing"
)

func TestSyntheticUser_NoSecretFields(t *testing.T) {
	user := SyntheticUser()
	if user.Email == "" || !strings.HasSuffix(user.Email, ".test") {
		t.Fatalf("expected synthetic test email domain, got %q", user.Email)
	}

	if strings.Contains(strings.ToLower(user.GoogleID), "prod") {
		t.Fatalf("fixture must not look like production identity: %q", user.GoogleID)
	}
}

func TestSyntheticFixtures_Deterministic(t *testing.T) {
	journal := SyntheticJournal()
	routine := SyntheticRoutineCheckIn()
	if journal.UserID != routine.UserID {
		t.Fatalf("expected fixture user id alignment, got journal=%q routine=%q", journal.UserID, routine.UserID)
	}
}
