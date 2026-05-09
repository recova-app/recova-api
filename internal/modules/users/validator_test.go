package users

import "testing"

func TestNormalizeSettingsUpdate_EmptyPayload(t *testing.T) {
	_, err := NormalizeSettingsUpdate(SettingsUpdateRequest{})
	if err == nil {
		t.Fatal("expected validation error")
	}
}

func TestNormalizeSettingsUpdate_Success(t *testing.T) {
	nickname := "tester"
	recoveryReason := "fokus pulih"
	checkIn := "08:15"

	updates, err := NormalizeSettingsUpdate(SettingsUpdateRequest{
		Nickname:         &nickname,
		RecoveryReason:   &recoveryReason,
		DailyCheckInTime: &checkIn,
	})
	if err != nil {
		t.Fatalf("normalize settings: %v", err)
	}
	if updates["nickname"] != "tester" {
		t.Fatalf("unexpected nickname update: %#v", updates["nickname"])
	}
	if updates["user_why"] != "fokus pulih" {
		t.Fatalf("unexpected recovery reason update: %#v", updates["user_why"])
	}
	if _, ok := updates["check_in_time"]; !ok {
		t.Fatal("expected check_in_time update key")
	}
}

func TestNormalizeOnboardingRequest_UsesLegacyFields(t *testing.T) {
	input, err := NormalizeOnboardingRequest(OnboardingRequest{
		Nickname:               "tester",
		RecoveryReasonLegacy:   "konsisten",
		DailyCheckInTimeLegacy: "06:45",
		DependencyLevelLegacy:  "low",
		Answers:                map[string]any{"q1": "a1"},
	})
	if err != nil {
		t.Fatalf("normalize onboarding: %v", err)
	}
	if input.RecoveryReason != "konsisten" {
		t.Fatalf("unexpected recovery reason: %s", input.RecoveryReason)
	}
	if input.DailyCheckInRaw != "06:45" {
		t.Fatalf("unexpected checkin raw: %s", input.DailyCheckInRaw)
	}
	if input.DependencyLevel == nil || *input.DependencyLevel != "low" {
		t.Fatalf("unexpected dependency level: %#v", input.DependencyLevel)
	}
}
