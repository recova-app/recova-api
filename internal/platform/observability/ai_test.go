package observability

import (
	"context"
	"errors"
	"io"
	"net/http/httptest"
	"strings"
	"testing"

	aiplatform "github.com/recova-app/backend-v2/internal/platform/ai"
)

func TestWrapAIClient_NilClient_ReturnsError(t *testing.T) {
	var client *instrumentedAIClient
	resp, err := client.Generate(context.Background(), aiplatform.GenerateRequest{UserPrompt: "test"})
	if err == nil {
		t.Fatal("expected error for nil client")
	}
	if resp.Text != "" {
		t.Fatalf("expected empty response, got: %v", resp)
	}
}

func TestWrapAIClient_RecordsMetrics(t *testing.T) {
	recorder := NewRecorder()
	client := WrapAIClient(&fakeAIClient{
		response: aiplatform.GenerateResponse{
			Provider: aiplatform.ProviderGemini,
			Model:    "gemini-2.0-flash",
			Text:     "ok",
		},
	}, recorder)

	if _, err := client.Generate(context.Background(), aiplatform.GenerateRequest{UserPrompt: "halo"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	client = WrapAIClient(&fakeAIClient{err: errors.New("timeout")}, recorder)
	if _, err := client.Generate(context.Background(), aiplatform.GenerateRequest{UserPrompt: "halo"}); err == nil {
		t.Fatal("expected error from fake ai client")
	}

	req := httptest.NewRequest("GET", "/metrics", nil)
	w := httptest.NewRecorder()
	recorder.MetricsHandler().ServeHTTP(w, req)
	body, err := io.ReadAll(w.Result().Body)
	if err != nil {
		t.Fatalf("read metrics: %v", err)
	}
	text := string(body)
	if !strings.Contains(text, "recova_ai_request_duration_seconds") {
		t.Fatalf("expected ai metric in output: %s", text)
	}
	if !strings.Contains(text, "provider=\"gemini\"") {
		t.Fatalf("expected gemini provider label in output: %s", text)
	}
}

type fakeAIClient struct {
	response aiplatform.GenerateResponse
	err      error
	observer aiplatform.FallbackObserver
}

func (f *fakeAIClient) SetFallbackObserver(observer aiplatform.FallbackObserver) {
	f.observer = observer
}

func (f *fakeAIClient) Generate(_ context.Context, _ aiplatform.GenerateRequest) (aiplatform.GenerateResponse, error) {
	if f.err != nil {
		return aiplatform.GenerateResponse{}, f.err
	}
	return f.response, nil
}

func TestWrapAIClient_HooksFallbackObserver(t *testing.T) {
	recorder := NewRecorder()
	fake := &fakeAIClient{}
	client := WrapAIClient(fake, recorder)

	if fake.observer == nil {
		t.Fatal("expected fallback observer to be hooked onto inner client")
	}

	fake.observer.RecordAIFallback("openai-compatible", nil)

	req := httptest.NewRequest("GET", "/metrics", nil)
	w := httptest.NewRecorder()
	recorder.MetricsHandler().ServeHTTP(w, req)
	body, err := io.ReadAll(w.Result().Body)
	if err != nil {
		t.Fatalf("read metrics: %v", err)
	}
	text := string(body)
	if !strings.Contains(text, "recova_ai_fallback_total") {
		t.Fatalf("expected recova_ai_fallback_total in output: %s", text)
	}
	if !strings.Contains(text, "provider=\"openai-compatible\",status=\"success\"") {
		t.Fatalf("expected provider/status labels in output: %s", text)
	}

	// Also verify delegation on instrumented client
	if delegator, ok := client.(interface {
		SetFallbackObserver(observer aiplatform.FallbackObserver)
	}); ok {
		delegator.SetFallbackObserver(nil)
		if fake.observer != nil {
			t.Fatal("expected observer to be cleared via delegator")
		}
	}
}
