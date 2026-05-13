// Package httpharness provides reusable helpers for Fiber handler and contract tests.
package httpharness

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v3"
)

// Response captures HTTP testing result and optional parsed JSON payload.
type Response struct {
	StatusCode int
	Header     http.Header
	Body       []byte
	JSON       map[string]any
}

// JSONRequest executes one HTTP request to Fiber app and parses JSON response payload.
func JSONRequest(t testing.TB, app *fiber.App, method string, path string, body any, headers map[string]string) Response {
	t.Helper()

	var bodyReader io.Reader
	if body != nil {
		encoded, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal request body: %v", err)
		}
		bodyReader = bytes.NewReader(encoded)
	}

	req := httptest.NewRequest(method, path, bodyReader)
	for key, value := range headers {
		req.Header.Set(key, value)
	}
	if body != nil && strings.TrimSpace(req.Header.Get("Content-Type")) == "" {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read response body: %v", err)
	}

	result := Response{
		StatusCode: resp.StatusCode,
		Header:     resp.Header.Clone(),
		Body:       bodyBytes,
	}

	contentType := strings.ToLower(strings.TrimSpace(resp.Header.Get("Content-Type")))
	if strings.HasPrefix(contentType, "application/json") {
		var payload map[string]any
		if err := json.Unmarshal(bodyBytes, &payload); err != nil {
			t.Fatalf("parse json response: %v\nbody=%s", err, string(bodyBytes))
		}
		result.JSON = payload
	}

	return result
}

// RequireSuccessEnvelope validates minimal success envelope fields.
func RequireSuccessEnvelope(t testing.TB, payload map[string]any) {
	t.Helper()

	if payload == nil {
		t.Fatal("expected json payload, got nil")
	}

	ok, castOK := payload["success"].(bool)
	if !castOK || !ok {
		t.Fatalf("expected success=true, got: %#v", payload["success"])
	}

	if _, exists := payload["message"]; !exists {
		t.Fatalf("expected field 'message' in payload: %#v", payload)
	}

	if _, exists := payload["data"]; !exists {
		t.Fatalf("expected field 'data' in payload: %#v", payload)
	}
}

// RequireErrorEnvelope validates minimal error envelope fields and optionally asserts error code.
func RequireErrorEnvelope(t testing.TB, payload map[string]any, expectedCode string) {
	t.Helper()

	if payload == nil {
		t.Fatal("expected json payload, got nil")
	}

	ok, castOK := payload["success"].(bool)
	if !castOK || ok {
		t.Fatalf("expected success=false, got: %#v", payload["success"])
	}

	errMap, mapOK := payload["error"].(map[string]any)
	if !mapOK {
		t.Fatalf("expected error object, got: %#v", payload["error"])
	}

	if strings.TrimSpace(expectedCode) == "" {
		return
	}

	code, codeOK := errMap["code"].(string)
	if !codeOK {
		t.Fatalf("expected error.code string, got: %#v", errMap["code"])
	}

	if strings.TrimSpace(code) != strings.TrimSpace(expectedCode) {
		t.Fatalf("unexpected error code: want=%q got=%q", expectedCode, code)
	}
}

// RequireStatus asserts exact status code.
func RequireStatus(t testing.TB, got int, want int) {
	t.Helper()
	if got != want {
		t.Fatalf("unexpected status code: want=%d got=%d", want, got)
	}
}

// MustPath normalizes and validates non-empty route path input.
func MustPath(t testing.TB, path string) string {
	t.Helper()
	trimmed := strings.TrimSpace(path)
	if trimmed == "" {
		t.Fatal("path must not be empty")
	}
	if !strings.HasPrefix(trimmed, "/") {
		t.Fatalf("path must start with '/': %s", trimmed)
	}
	return trimmed
}

// DebugBody returns readable body string for diagnostics.
func DebugBody(resp Response) string {
	return fmt.Sprintf("status=%d body=%s", resp.StatusCode, string(resp.Body))
}
