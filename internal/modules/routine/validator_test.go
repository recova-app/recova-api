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
	success := true
	_, err := NormalizeDailyCheckInRequest(DailyCheckInRequest{
		Mood:           "tenang",
		IsSuccessful:   &success,
		RelapseTrigger: []string{"trigger"},
	})
	if err == nil {
		t.Fatal("expected validation error when relapse trigger sent on check-in endpoint")
	}

	failure := false
	_, err = NormalizeDailyCheckInRequest(DailyCheckInRequest{
		Mood:         "cemas",
		IsSuccessful: &failure,
	})
	if err == nil {
		t.Fatal("expected validation error when check-in marked unsuccessful")
	}
}

func TestNormalizeRelapseRequest_ValidationRules(t *testing.T) {
	_, err := NormalizeRelapseRequest(RelapseRequest{
		Mood:           "",
		RelapseTrigger: []string{" ", ""},
	})
	if err == nil {
		t.Fatal("expected validation error when relapse trigger empty")
	}

	normalized, err := NormalizeRelapseRequest(RelapseRequest{
		Mood:           "cemas",
		RelapseTrigger: []string{" stres kerja ", "sendiri malam"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(normalized.RelapseTrigger) != 2 {
		t.Fatalf("expected 2 relapse triggers, got %d", len(normalized.RelapseTrigger))
	}
	if normalized.RelapseTrigger[0] != "stres kerja" {
		t.Fatalf("unexpected first relapse trigger: %s", normalized.RelapseTrigger[0])
	}
	if normalized.Mood != "cemas" {
		t.Fatalf("expected mood cemas, got %s", normalized.Mood)
	}
}
