package observability

import (
	"errors"
	"io"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestRecorder_MetricsHandler_ExposesRecordedMetrics(t *testing.T) {
	recorder := NewRecorder()
	recorder.RecordHTTPRequest("GET", "/health/live", 200, 15*time.Millisecond)
	recorder.RecordHTTPRequest("POST", "/api/v1/auth/google", 401, 12*time.Millisecond)
	recorder.RecordDBOperation("query", "users", 4*time.Millisecond, nil)
	recorder.RecordDBOperation("update", "profiles", 7*time.Millisecond, errors.New("failed"))
	recorder.RecordAIRequest("gemini", "gemini-2.0-flash", 82*time.Millisecond, nil)
	recorder.RecordAuditEvent("auth.login", "succeeded")

	req := httptest.NewRequest("GET", "/metrics", nil)
	w := httptest.NewRecorder()
	recorder.MetricsHandler().ServeHTTP(w, req)

	body, err := io.ReadAll(w.Result().Body)
	if err != nil {
		t.Fatalf("read metrics body: %v", err)
	}

	text := string(body)
	assertContains(t, text, "recova_http_requests_total")
	assertContains(t, text, "recova_http_request_duration_seconds")
	assertContains(t, text, "recova_http_errors_total")
	assertContains(t, text, "recova_db_operation_duration_seconds")
	assertContains(t, text, "recova_ai_request_duration_seconds")
	assertContains(t, text, "recova_audit_events_total")
	assertContains(t, text, "route=\"/health/live\"")
	assertContains(t, text, "route=\"/api/v1/auth/google\"")
}

func TestRegisterDatabaseMetrics_NilSafe(t *testing.T) {
	if err := RegisterDatabaseMetrics(nil, NewRecorder()); err != nil {
		t.Fatalf("expected nil-safe registration, got: %v", err)
	}
	if err := RegisterDatabaseMetrics(nil, nil); err != nil {
		t.Fatalf("expected nil-safe registration with nil recorder, got: %v", err)
	}
}

func TestAuditAction_KnownRoutes(t *testing.T) {
	cases := []struct {
		method string
		path   string
		action string
	}{
		{method: "POST", path: "/api/v1/auth/google", action: "auth.login"},
		{method: "POST", path: "/api/v1/auth/refresh", action: "auth.refresh"},
		{method: "PUT", path: "/api/v1/users/settings", action: "users.settings.update"},
	}

	for _, tc := range cases {
		got, ok := auditAction(tc.method, tc.path)
		if !ok {
			t.Fatalf("expected audit action for %s %s", tc.method, tc.path)
		}
		if got != tc.action {
			t.Fatalf("unexpected action for %s %s: want=%s got=%s", tc.method, tc.path, tc.action, got)
		}
	}

	if _, ok := auditAction("GET", "/api/v1/community"); ok {
		t.Fatal("expected non-audited route to return false")
	}
}

func assertContains(t *testing.T, text string, pattern string) {
	t.Helper()
	if !strings.Contains(text, pattern) {
		t.Fatalf("expected metrics output contains %q", pattern)
	}
}
