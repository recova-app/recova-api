package routine

import "testing"

func TestNormalizeActivitySummaryWindow_Default(t *testing.T) {
	value, err := NormalizeActivitySummaryWindow(nil)
	if err != nil {
		t.Fatalf("normalize default window: %v", err)
	}
	if value != 30 {
		t.Fatalf("expected default window 30, got %d", value)
	}
}

func TestNormalizeActivitySummaryWindow_ValidationError(t *testing.T) {
	raw := 6
	_, err := NormalizeActivitySummaryWindow(&raw)
	if err == nil {
		t.Fatal("expected validation error for window_days < 7")
	}
}

func TestNormalizeDailyCheckInRequest_RelapseTriggerRules(t *testing.T) {
	isSuccess := false
	normalized, err := NormalizeDailyCheckInRequest(DailyCheckInRequest{
		Mood:           "cemas",
		IsSuccessful:   &isSuccess,
		RelapseTrigger: []string{" scrolling malam ", "  "},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(normalized.RelapseTrigger) != 1 || normalized.RelapseTrigger[0] != "scrolling malam" {
		t.Fatalf("expected relapse trigger normalized, got %+v", normalized.RelapseTrigger)
	}

	success := true
	_, err = NormalizeDailyCheckInRequest(DailyCheckInRequest{
		Mood:           "tenang",
		IsSuccessful:   &success,
		RelapseTrigger: []string{"trigger"},
	})
	if err == nil {
		t.Fatal("expected validation error when relapse trigger on successful check-in")
	}
}
