package auth

import "testing"

func TestValidateGoogleLoginRequest_EmptyToken(t *testing.T) {
	err := ValidateGoogleLoginRequest(GoogleLoginRequest{})
	if err == nil {
		t.Fatal("expected validation error")
	}
}

func TestValidateGoogleLoginRequest_Success(t *testing.T) {
	err := ValidateGoogleLoginRequest(GoogleLoginRequest{Token: "google-token"})
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}
}

func TestNormalizeAndValidateOnboardingRequest_EmptyNickname(t *testing.T) {
	_, err := NormalizeAndValidateOnboardingRequest(OnboardingRequest{
		Nickname:         "  ",
		RecoveryReason:   "ingin pulih",
		DailyCheckInTime: "09:00",
	})
	if err == nil {
		t.Fatal("expected validation error")
	}
}

func TestNormalizeAndValidateOnboardingRequest_UsesLegacyFields(t *testing.T) {
	input, err := NormalizeAndValidateOnboardingRequest(OnboardingRequest{
		Nickname:               "tester",
		RecoveryReasonLegacy:   "fokus sehat",
		DailyCheckInTimeLegacy: "07:30",
		DependencyLevelLegacy:  "medium",
		Answers:                map[string]any{"q1": "a1"},
	})
	if err != nil {
		t.Fatalf("normalize onboarding: %v", err)
	}
	if input.RecoveryReason != "fokus sehat" {
		t.Fatalf("unexpected recovery reason: %s", input.RecoveryReason)
	}
	if input.DailyCheckInRaw != "07:30" {
		t.Fatalf("unexpected check-in raw: %s", input.DailyCheckInRaw)
	}
	if input.DependencyLevel == nil || *input.DependencyLevel != "medium" {
		t.Fatalf("unexpected dependency level: %#v", input.DependencyLevel)
	}
}
