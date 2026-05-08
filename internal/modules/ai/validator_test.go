package ai

import "testing"

func TestNormalizeAskCoachRequest_Validation(t *testing.T) {
	_, err := NormalizeAskCoachRequest(AskCoachRequest{Message: "   "})
	if err == nil {
		t.Fatal("expected validation error")
	}
}

func TestNormalizeOnboardingAnalysisRequest_Validation(t *testing.T) {
	_, err := NormalizeOnboardingAnalysisRequest(OnboardingAnalysisRequest{Answers: map[string]any{}})
	if err == nil {
		t.Fatal("expected validation error")
	}
}

func TestNormalizeChatHistoryLimit_DefaultAndBounds(t *testing.T) {
	limit, err := NormalizeChatHistoryLimit(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if limit != defaultHistoryLimit {
		t.Fatalf("expected default limit %d, got %d", defaultHistoryLimit, limit)
	}

	invalid := 0
	if _, err := NormalizeChatHistoryLimit(&invalid); err == nil {
		t.Fatal("expected validation error for zero limit")
	}

	valid := 25
	limit, err = NormalizeChatHistoryLimit(&valid)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if limit != valid {
		t.Fatalf("expected limit %d, got %d", valid, limit)
	}
}
