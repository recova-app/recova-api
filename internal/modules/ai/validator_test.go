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

func TestNormalizePersonaPreferenceRequest(t *testing.T) {
	normalized, err := NormalizePersonaPreferenceRequest(PersonaPreferenceRequest{Persona: " Friendly "})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if normalized.Persona != "friendly" {
		t.Fatalf("expected friendly, got %q", normalized.Persona)
	}

	if _, err := NormalizePersonaPreferenceRequest(PersonaPreferenceRequest{Persona: "unknown"}); err == nil {
		t.Fatal("expected validation error for unknown persona")
	}
}

func TestResolvePersonaOrDefault(t *testing.T) {
	if got := ResolvePersonaOrDefault("direct"); got != "direct" {
		t.Fatalf("expected direct, got %q", got)
	}
	if got := ResolvePersonaOrDefault("invalid"); got != DefaultPersona {
		t.Fatalf("expected default persona, got %q", got)
	}
}

func TestNormalizeRelapseSolutionRequest(t *testing.T) {
	normalized, err := NormalizeRelapseSolutionRequest(RelapseSolutionRequest{
		Mood:           " cemas ",
		RelapseTrigger: []string{" sosmed malam ", "  "},
		Commitment:     ptr(" lanjut recovery "),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if normalized.Mood != "cemas" {
		t.Fatalf("expected mood normalized, got %q", normalized.Mood)
	}
	if len(normalized.RelapseTrigger) != 1 || normalized.RelapseTrigger[0] != "sosmed malam" {
		t.Fatalf("expected relapse trigger normalized, got %+v", normalized.RelapseTrigger)
	}

	if _, err := NormalizeRelapseSolutionRequest(RelapseSolutionRequest{Mood: " "}); err == nil {
		t.Fatal("expected validation error for empty mood")
	}
}

func ptr(v string) *string {
	return &v
}
