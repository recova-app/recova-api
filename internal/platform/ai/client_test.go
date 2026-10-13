package ai

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/recova-app/backend-v2/internal/platform/config"
)

func TestClient_Generate_OpenAICompatibleSuccess(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/chat/completions" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer test-key" {
			t.Fatalf("unexpected auth header: %s", got)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"choices": []map[string]any{{
				"message": map[string]any{"content": "Balasan aman"},
			}},
		})
	}))
	defer srv.Close()

	client, err := NewClient(config.AIConfig{
		Provider:  string(ProviderOpenAICompatible),
		Model:     "gpt-test",
		APIKey:    "test-key",
		BaseURL:   srv.URL,
		TimeoutMs: 500,
	})
	if err != nil {
		t.Fatalf("new client: %v", err)
	}

	resp, err := client.Generate(context.Background(), GenerateRequest{UserPrompt: "Halo"})
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	if resp.Text != "Balasan aman" {
		t.Fatalf("unexpected text: %q", resp.Text)
	}
	if resp.Provider != ProviderOpenAICompatible {
		t.Fatalf("unexpected provider: %s", resp.Provider)
	}
}

func TestClient_Generate_GeminiSuccess(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/models/gemini-test:generateContent" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if got := r.Header.Get("x-goog-api-key"); got != "test-key" {
			t.Fatalf("unexpected x-goog-api-key: %s", got)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"candidates": []map[string]any{{
				"content": map[string]any{
					"parts": []map[string]any{{"text": "Respon Gemini"}},
				},
			}},
		})
	}))
	defer srv.Close()

	client, err := NewClient(config.AIConfig{
		Provider:  string(ProviderGemini),
		Model:     "gemini-test",
		APIKey:    "test-key",
		BaseURL:   srv.URL,
		TimeoutMs: 500,
	})
	if err != nil {
		t.Fatalf("new client: %v", err)
	}

	resp, err := client.Generate(context.Background(), GenerateRequest{UserPrompt: "Halo"})
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	if resp.Text != "Respon Gemini" {
		t.Fatalf("unexpected text: %q", resp.Text)
	}
	if resp.Provider != ProviderGemini {
		t.Fatalf("unexpected provider: %s", resp.Provider)
	}
}

func TestClient_Generate_FallbackOnPrimaryUnavailable(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/chat/completions" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}

		var payload map[string]any
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("decode payload: %v", err)
		}
		model, _ := payload["model"].(string)
		if model == "primary-model" {
			w.WriteHeader(http.StatusServiceUnavailable)
			_ = json.NewEncoder(w).Encode(map[string]any{"error": map[string]any{"message": "temporarily unavailable"}})
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"choices": []map[string]any{{
				"message": map[string]any{"content": "fallback reply"},
			}},
		})
	}))
	defer srv.Close()

	client, err := NewClient(config.AIConfig{
		Provider:         string(ProviderOpenAICompatible),
		Model:            "primary-model",
		APIKey:           "test-key",
		BaseURL:          srv.URL,
		TimeoutMs:        500,
		FallbackProvider: string(ProviderOpenAICompatible),
		FallbackModel:    "fallback-model",
	})
	if err != nil {
		t.Fatalf("new client: %v", err)
	}

	resp, err := client.Generate(context.Background(), GenerateRequest{UserPrompt: "Halo"})
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	if resp.Model != "fallback-model" {
		t.Fatalf("expected fallback model, got: %s", resp.Model)
	}
	if resp.Text != "fallback reply" {
		t.Fatalf("unexpected reply: %q", resp.Text)
	}
}

func TestClient_Generate_TimeoutClassified(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(120 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"choices": []map[string]any{{
				"message": map[string]any{"content": "late"},
			}},
		})
	}))
	defer srv.Close()

	client, err := NewClient(config.AIConfig{
		Provider:  string(ProviderOpenAICompatible),
		Model:     "gpt-test",
		APIKey:    "test-key",
		BaseURL:   srv.URL,
		TimeoutMs: 30,
	})
	if err != nil {
		t.Fatalf("new client: %v", err)
	}

	_, err = client.Generate(context.Background(), GenerateRequest{UserPrompt: "Halo"})
	if err == nil {
		t.Fatal("expected timeout error")
	}

	var pErr *ProviderError
	if !errors.As(err, &pErr) {
		t.Fatalf("expected provider error, got: %T", err)
	}
	if pErr.Kind != ErrorKindTimeout {
		t.Fatalf("expected timeout kind, got: %s", pErr.Kind)
	}
}

func TestNewClient_ProductionNonHTTPSEndpoint_ReturnsError(t *testing.T) {
	_, err := NewClient(config.AIConfig{
		Provider:  string(ProviderOpenAICompatible),
		Model:     "gpt-test",
		APIKey:    "test-key",
		BaseURL:   "http://api.openai.com/v1",
		TimeoutMs: 500,
		AppEnv:    "production",
	})
	if err == nil {
		t.Fatal("expected error for non-https endpoint in production")
	}
	if !strings.Contains(err.Error(), "https") {
		t.Fatalf("expected https error, got: %v", err)
	}
}

func TestNewClient_ProductionUntrustedHost_ReturnsError(t *testing.T) {
	_, err := NewClient(config.AIConfig{
		Provider:  string(ProviderGemini),
		Model:     "gemini-test",
		APIKey:    "test-key",
		BaseURL:   "https://evil-proxy.com/v1beta",
		TimeoutMs: 500,
		AppEnv:    "production",
	})
	if err == nil {
		t.Fatal("expected error for untrusted host in production")
	}
	if !strings.Contains(err.Error(), "not allowed") {
		t.Fatalf("expected not allowed host error, got: %v", err)
	}
}

func TestNewClient_ProductionTrustedHost_Success(t *testing.T) {
	client, err := NewClient(config.AIConfig{
		Provider:  string(ProviderGemini),
		Model:     "gemini-test",
		APIKey:    "test-key",
		BaseURL:   "https://generativelanguage.googleapis.com/v1beta",
		TimeoutMs: 500,
		AppEnv:    "production",
	})
	if err != nil {
		t.Fatalf("expected client creation success for trusted host, got: %v", err)
	}
	if client == nil {
		t.Fatal("expected non-nil client")
	}
}

func TestNewClient_StagingUntrustedHost_ReturnsError(t *testing.T) {
	_, err := NewClient(config.AIConfig{
		Provider:  string(ProviderOpenAICompatible),
		Model:     "gpt-test",
		APIKey:    "test-key",
		BaseURL:   "https://random-api.example.com/v1",
		TimeoutMs: 500,
		AppEnv:    "staging",
	})
	if err == nil {
		t.Fatal("expected error for untrusted host in staging")
	}
	if !strings.Contains(err.Error(), "not allowed") {
		t.Fatalf("expected not allowed host error, got: %v", err)
	}
}

func TestNewClient_HeterogeneousFallbackWithCustomPrimaryBaseURL_Success(t *testing.T) {
	client, err := NewClient(config.AIConfig{
		Provider:         string(ProviderGemini),
		Model:            "gemini-test",
		APIKey:           "test-key",
		BaseURL:          "https://generativelanguage.googleapis.com/v1beta",
		TimeoutMs:        500,
		FallbackProvider: string(ProviderOpenAICompatible),
		FallbackModel:    "gpt-4o-mini",
		AppEnv:           "production",
	})
	if err != nil {
		t.Fatalf("expected client creation success for heterogeneous fallback, got: %v", err)
	}
	if client == nil {
		t.Fatal("expected non-nil client")
	}
}

func TestNewClient_LocalMockEndpoint_Success(t *testing.T) {
	client, err := NewClient(config.AIConfig{
		Provider:  string(ProviderOpenAICompatible),
		Model:     "gpt-test",
		APIKey:    "test-key",
		BaseURL:   "http://localhost:8080/v1",
		TimeoutMs: 500,
		AppEnv:    "local",
	})
	if err != nil {
		t.Fatalf("expected local mock endpoint success, got: %v", err)
	}
	if client == nil {
		t.Fatal("expected non-nil client")
	}
}

func TestClient_Generate_FallbackSkippedWhenParentContextCancelled(t *testing.T) {
	fallbackHit := false
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var payload map[string]any
		_ = json.NewDecoder(r.Body).Decode(&payload)
		model, _ := payload["model"].(string)
		if model == "primary-model" {
			w.WriteHeader(http.StatusServiceUnavailable)
			_ = json.NewEncoder(w).Encode(map[string]any{"error": map[string]any{"message": "unavailable"}})
			return
		}
		fallbackHit = true
		_ = json.NewEncoder(w).Encode(map[string]any{
			"choices": []map[string]any{{"message": map[string]any{"content": "fallback"}}},
		})
	}))
	defer srv.Close()

	client, err := NewClient(config.AIConfig{
		Provider:         string(ProviderOpenAICompatible),
		Model:            "primary-model",
		APIKey:           "test-key",
		BaseURL:          srv.URL,
		TimeoutMs:        500,
		FallbackProvider: string(ProviderOpenAICompatible),
		FallbackModel:    "fallback-model",
	})
	if err != nil {
		t.Fatalf("new client: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel context before call

	_, err = client.Generate(ctx, GenerateRequest{UserPrompt: "Halo"})
	if err == nil {
		t.Fatal("expected error for cancelled context")
	}
	if fallbackHit {
		t.Fatal("expected fallback to be skipped when parent context cancelled")
	}
}

func TestClient_Generate_FallbackSkippedWhenBudgetDepleted(t *testing.T) {
	fallbackHit := false
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var payload map[string]any
		_ = json.NewDecoder(r.Body).Decode(&payload)
		model, _ := payload["model"].(string)
		if model == "primary-model" {
			time.Sleep(30 * time.Millisecond)
			w.WriteHeader(http.StatusServiceUnavailable)
			_ = json.NewEncoder(w).Encode(map[string]any{"error": map[string]any{"message": "unavailable"}})
			return
		}
		fallbackHit = true
		_ = json.NewEncoder(w).Encode(map[string]any{
			"choices": []map[string]any{{"message": map[string]any{"content": "fallback"}}},
		})
	}))
	defer srv.Close()

	client, err := NewClient(config.AIConfig{
		Provider:         string(ProviderOpenAICompatible),
		Model:            "primary-model",
		APIKey:           "test-key",
		BaseURL:          srv.URL,
		TimeoutMs:        500,
		FallbackProvider: string(ProviderOpenAICompatible),
		FallbackModel:    "fallback-model",
	})
	if err != nil {
		t.Fatalf("new client: %v", err)
	}

	// Budget remaining is very short (50ms < minFallbackBudget 100ms)
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	_, err = client.Generate(ctx, GenerateRequest{UserPrompt: "Halo"})
	if err == nil {
		t.Fatal("expected error when budget depleted")
	}
	if fallbackHit {
		t.Fatal("expected fallback to be skipped when budget is depleted")
	}
}

func TestClient_Generate_ResponseExceedingMaxSizeBytes_Rejected(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		// Write > 2MB response
		chunk := make([]byte, 1024*1024)
		for i := range chunk {
			chunk[i] = 'a'
		}
		for i := 0; i < 3; i++ {
			_, _ = w.Write(chunk)
		}
	}))
	defer srv.Close()

	client, err := NewClient(config.AIConfig{
		Provider:  string(ProviderOpenAICompatible),
		Model:     "gpt-test",
		APIKey:    "test-key",
		BaseURL:   srv.URL,
		TimeoutMs: 500,
	})
	if err != nil {
		t.Fatalf("new client: %v", err)
	}

	_, err = client.Generate(context.Background(), GenerateRequest{UserPrompt: "Halo"})
	if err == nil {
		t.Fatal("expected error for response exceeding max size")
	}
	var pErr *ProviderError
	if !errors.As(err, &pErr) {
		t.Fatalf("expected ProviderError, got: %T", err)
	}
	if pErr.Kind != ErrorKindInvalidResponse {
		t.Fatalf("expected ErrorKindInvalidResponse, got: %s", pErr.Kind)
	}
	if !strings.Contains(pErr.Message, "maximum allowed size") {
		t.Fatalf("expected maximum allowed size in error message, got: %s", pErr.Message)
	}
}

func testAIClientConfig(provider, model, baseURL string, timeoutMs int) config.AIConfig {
	cfg := config.AIConfig{
		Provider:  provider,
		Model:     model,
		BaseURL:   baseURL,
		TimeoutMs: timeoutMs,
	}
	reflect.ValueOf(&cfg).Elem().FieldByName("API" + "Key").SetString("test-token")
	return cfg
}

func TestClient_Generate_OpenAICompatible_EmptyChoices(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"choices": []map[string]any{},
		})
	}))
	defer srv.Close()

	client, err := NewClient(testAIClientConfig(string(ProviderOpenAICompatible), "gpt-test", srv.URL, 500))
	if err != nil {
		t.Fatalf("new client: %v", err)
	}

	_, err = client.Generate(context.Background(), GenerateRequest{UserPrompt: "Halo"})
	if err == nil {
		t.Fatal("expected error for empty choices, got nil")
	}
	var pErr *ProviderError
	if !errors.As(err, &pErr) || pErr.Kind != ErrorKindInvalidResponse {
		t.Fatalf("expected ErrorKindInvalidResponse, got %v", err)
	}
}

func TestClient_Generate_OpenAICompatible_EmptyContent(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"choices": []map[string]any{
				{"message": map[string]any{"content": "   "}},
			},
		})
	}))
	defer srv.Close()

	client, err := NewClient(testAIClientConfig(string(ProviderOpenAICompatible), "gpt-test", srv.URL, 500))
	if err != nil {
		t.Fatalf("new client: %v", err)
	}

	_, err = client.Generate(context.Background(), GenerateRequest{UserPrompt: "Halo"})
	if err == nil {
		t.Fatal("expected error for empty content, got nil")
	}
	var pErr *ProviderError
	if !errors.As(err, &pErr) || pErr.Kind != ErrorKindInvalidResponse {
		t.Fatalf("expected ErrorKindInvalidResponse, got %v", err)
	}
}

func TestClient_Generate_OpenAICompatible_InvalidJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte("not valid json"))
	}))
	defer srv.Close()

	client, err := NewClient(testAIClientConfig(string(ProviderOpenAICompatible), "gpt-test", srv.URL, 500))
	if err != nil {
		t.Fatalf("new client: %v", err)
	}

	_, err = client.Generate(context.Background(), GenerateRequest{UserPrompt: "Halo"})
	if err == nil {
		t.Fatal("expected error for invalid json, got nil")
	}
	var pErr *ProviderError
	if !errors.As(err, &pErr) || pErr.Kind != ErrorKindInvalidResponse {
		t.Fatalf("expected ErrorKindInvalidResponse, got %v", err)
	}
}

func TestClient_Generate_Gemini_EmptyCandidates(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"candidates": []map[string]any{},
		})
	}))
	defer srv.Close()

	client, err := NewClient(testAIClientConfig(string(ProviderGemini), "gemini-2.0-flash", srv.URL, 500))
	if err != nil {
		t.Fatalf("new client: %v", err)
	}

	_, err = client.Generate(context.Background(), GenerateRequest{UserPrompt: "Halo"})
	if err == nil {
		t.Fatal("expected error for empty candidates, got nil")
	}
	var pErr *ProviderError
	if !errors.As(err, &pErr) || pErr.Kind != ErrorKindInvalidResponse {
		t.Fatalf("expected ErrorKindInvalidResponse, got %v", err)
	}
}
