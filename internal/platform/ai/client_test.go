package ai

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
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
