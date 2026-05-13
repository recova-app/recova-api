package auth

import (
	"strings"
	"testing"
)

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

func TestNormalizeAndValidateManualRegisterRequest_PasswordMismatch(t *testing.T) {
	_, err := NormalizeAndValidateManualRegisterRequest(ManualRegisterRequest{
		Email:           "manual@example.com",
		Username:        "manual_user",
		Password:        "password123",
		ConfirmPassword: "password456",
	})
	if err == nil {
		t.Fatal("expected validation error")
	}
}

func TestNormalizeAndValidateManualRegisterRequest_InvalidEmail(t *testing.T) {
	_, err := NormalizeAndValidateManualRegisterRequest(ManualRegisterRequest{
		Email:           "not-an-email",
		Username:        "manual_user",
		Password:        "password123",
		ConfirmPassword: "password123",
	})
	if err == nil {
		t.Fatal("expected validation error")
	}
}

func TestNormalizeAndValidateManualRegisterRequest_InvalidUsername(t *testing.T) {
	_, err := NormalizeAndValidateManualRegisterRequest(ManualRegisterRequest{
		Email:           "manual@example.com",
		Username:        "Manual-User",
		Password:        "password123",
		ConfirmPassword: "password123",
	})
	if err == nil {
		t.Fatal("expected validation error")
	}
}

func TestNormalizeAndValidateManualRegisterRequest_WeakPasswordTooShort(t *testing.T) {
	_, err := NormalizeAndValidateManualRegisterRequest(ManualRegisterRequest{
		Email:           "manual@example.com",
		Username:        "manual_user",
		Password:        "short",
		ConfirmPassword: "short",
	})
	if err == nil {
		t.Fatal("expected validation error")
	}
}

func TestNormalizeAndValidateManualRegisterRequest_PasswordTooLong(t *testing.T) {
	password := strings.Repeat("a", 73)
	_, err := NormalizeAndValidateManualRegisterRequest(ManualRegisterRequest{
		Email:           "manual@example.com",
		Username:        "manual_user",
		Password:        password,
		ConfirmPassword: password,
	})
	if err == nil {
		t.Fatal("expected validation error")
	}
}

func TestNormalizeAndValidateManualRegisterRequest_Success(t *testing.T) {
	input, err := NormalizeAndValidateManualRegisterRequest(ManualRegisterRequest{
		Email:           "Manual@Example.com",
		Username:        "Manual_User",
		Password:        "password123",
		ConfirmPassword: "password123",
	})
	if err != nil {
		t.Fatalf("normalize register: %v", err)
	}
	if input.Email != "manual@example.com" {
		t.Fatalf("unexpected normalized email: %s", input.Email)
	}
	if input.Username != "manual_user" {
		t.Fatalf("unexpected normalized username: %s", input.Username)
	}
}

func TestNormalizeAndValidateManualLoginRequest_IdentifierFallback(t *testing.T) {
	input, err := NormalizeAndValidateManualLoginRequest(ManualLoginRequest{
		Email:    "Manual@Example.com",
		Password: "password123",
	})
	if err != nil {
		t.Fatalf("normalize login: %v", err)
	}
	if input.Identifier != "manual@example.com" {
		t.Fatalf("unexpected identifier: %s", input.Identifier)
	}
}

func TestNormalizeAndValidateManualLoginRequest_EmptyIdentifier(t *testing.T) {
	_, err := NormalizeAndValidateManualLoginRequest(ManualLoginRequest{
		Password: "password123",
	})
	if err == nil {
		t.Fatal("expected validation error")
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

func TestNormalizeAndValidateOnboardingRequest_UsesSnakeCaseFields(t *testing.T) {
	input, err := NormalizeAndValidateOnboardingRequest(OnboardingRequest{
		Nickname:         "tester",
		RecoveryReason:   "fokus sehat",
		DailyCheckInTime: "07:30",
		PornFreeGoal:     intPointer(7),
		DependencyLevel:  "medium",
		Answers:          map[string]any{"q1": "a1"},
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
	if input.PornFreeGoal != 7 {
		t.Fatalf("unexpected porn free goal: %d", input.PornFreeGoal)
	}
}

func intPointer(v int) *int {
	return &v
}
