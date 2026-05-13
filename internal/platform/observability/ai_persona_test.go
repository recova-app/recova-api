package observability

import (
	"io"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestNewAIPersonaTelemetry_NilRecorder_ReturnsNil(t *testing.T) {
	if NewAIPersonaTelemetry(nil) != nil {
		t.Fatal("expected nil telemetry when recorder nil")
	}
}

func TestAIPersonaTelemetry_RecordPersonaUsage_NilSafe(t *testing.T) {
	var tel *AIPersonaTelemetry
	tel.RecordPersonaUsage("generate", "direct", nil)
}

func TestAIPersonaTelemetry_RecordPersonaUsage_Records(t *testing.T) {
	rec := NewRecorder()
	tel := NewAIPersonaTelemetry(rec)
	tel.RecordPersonaUsage("generate", "direct", nil)

	req := httptest.NewRequest("GET", "/metrics", nil)
	resp := httptest.NewRecorder()
	rec.MetricsHandler().ServeHTTP(resp, req)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read metrics body: %v", err)
	}
	text := string(body)
	if !strings.Contains(text, "recova_ai_persona_usage_total") {
		t.Fatalf("expected persona usage metric present, got: %s", text)
	}
}
