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
		t.Fatal("expected validation error for windowDays < 7")
	}
}
