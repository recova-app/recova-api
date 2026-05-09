package contract

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/getkin/kin-openapi/routers"
	"github.com/getkin/kin-openapi/routers/legacy"
	"github.com/gofiber/fiber/v3"
	contractopenapi "github.com/recova-app/backend-v2/internal/platform/openapi"
	contractharness "github.com/recova-app/backend-v2/test/harness/contract"
	httpharness "github.com/recova-app/backend-v2/test/harness/http"
)

const generatedOpenAPIPath = "../../docs/generated/openapi.yaml"

func TestContract_RuntimeRoutes_MatchOpenAPISpec(t *testing.T) {
	srv := contractharness.BuildServer(t)

	doc, err := contractopenapi.LoadAndValidate(generatedOpenAPIPath)
	if err != nil {
		t.Fatalf("load openapi: %v", err)
	}

	runtimeSet := contractopenapi.RuntimeRouteSet(srv.Routes(true))
	specSet := contractopenapi.SpecRouteSet(doc)
	drift := contractopenapi.CompareRouteSets(runtimeSet, specSet)
	if drift.HasDrift() {
		t.Fatalf("route drift detected: runtimeMissing=%v specMissing=%v", drift.MissingInRuntime, drift.MissingInSpec)
	}
}

func TestContract_HealthResponses_ValidAgainstOpenAPI(t *testing.T) {
	srv := contractharness.BuildServer(t)

	router := buildOpenAPIRouter(t)

	tests := []struct {
		name       string
		path       string
		expectCode int
	}{
		{name: "live", path: "/health/live", expectCode: fiber.StatusOK},
		{name: "ready", path: "/health/ready", expectCode: fiber.StatusOK},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tc.path, nil)
			req.Header.Set("Accept", "application/json")

			resp, err := srv.Test(req)
			if err != nil {
				t.Fatalf("execute request: %v", err)
			}
			defer resp.Body.Close()

			bodyBytes, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Fatalf("read body: %v", err)
			}

			if resp.StatusCode != tc.expectCode {
				t.Fatalf("unexpected status code: %d", resp.StatusCode)
			}

			validateHTTPContract(t, router, req, resp.StatusCode, resp.Header, bodyBytes, true)
		})
	}
}

func TestContract_ProtectedRoutes_Unauthenticated_ValidAgainstOpenAPI(t *testing.T) {
	srv := contractharness.BuildServer(t)
	router := buildOpenAPIRouter(t)

	tests := []struct {
		name   string
		method string
		path   string
		body   any
	}{
		{
			name:   "auth onboarding",
			method: http.MethodPost,
			path:   "/api/v1/auth/onboarding",
			body: map[string]any{
				"nickname":           "uji-user",
				"recovery_reason":    "ingin pulih",
				"daily_checkin_time": "09:00",
			},
		},
		{name: "auth logout", method: http.MethodPost, path: "/api/v1/auth/logout"},
		{name: "users me", method: http.MethodGet, path: "/api/v1/users/me"},
		{name: "users settings", method: http.MethodPut, path: "/api/v1/users/settings", body: map[string]any{"nickname": "uji"}},
		{name: "users reset", method: http.MethodDelete, path: "/api/v1/users/me/reset-data"},
		{
			name:   "routine checkin",
			method: http.MethodPost,
			path:   "/api/v1/routine/checkin",
			body: map[string]any{
				"mood":         "tenang",
				"isSuccessful": true,
				"commitment":   "fokus",
			},
		},
		{name: "routine statistics", method: http.MethodGet, path: "/api/v1/routine/statistics"},
		{name: "routine activity summary", method: http.MethodGet, path: "/api/v1/routine/statistics/activity-summary"},
		{name: "routine relapses", method: http.MethodGet, path: "/api/v1/routine/relapses"},
		{name: "journals list", method: http.MethodGet, path: "/api/v1/journals"},
		{name: "journals create", method: http.MethodPost, path: "/api/v1/journals", body: map[string]any{"content": "uji kontrak"}},
		{name: "community list", method: http.MethodGet, path: "/api/v1/community"},
		{
			name:   "community create",
			method: http.MethodPost,
			path:   "/api/v1/community",
			body: map[string]any{
				"content":  "konten komunitas uji valid",
				"category": "motivation",
			},
		},
		{name: "community comment", method: http.MethodPost, path: "/api/v1/community/post-1/comments", body: map[string]any{"content": "uji"}},
		{name: "community comment thread", method: http.MethodGet, path: "/api/v1/community/post-1/comments"},
		{name: "community comment reply", method: http.MethodPost, path: "/api/v1/community/post-1/comments/comment-1/replies", body: map[string]any{"content": "uji"}},
		{name: "community like", method: http.MethodPost, path: "/api/v1/community/post-1/like"},
		{name: "achievements catalog", method: http.MethodGet, path: "/api/v1/achievements/catalog"},
		{name: "achievements progress", method: http.MethodGet, path: "/api/v1/achievements/progress"},
		{name: "achievements unlocked", method: http.MethodGet, path: "/api/v1/achievements/unlocked"},
		{name: "education list", method: http.MethodGet, path: "/api/v1/education"},
		{name: "daily content", method: http.MethodGet, path: "/api/v1/content/daily"},
		{name: "ai ask coach", method: http.MethodPost, path: "/api/v1/ai/ask-coach", body: map[string]any{"message": "halo"}},
		{name: "ai chat history", method: http.MethodGet, path: "/api/v1/ai/chat-history"},
		{name: "ai summary", method: http.MethodGet, path: "/api/v1/ai/summary"},
		{
			name:   "ai onboarding analysis",
			method: http.MethodPost,
			path:   "/api/v1/ai/onboarding-analysis",
			body: map[string]any{
				"answers": map[string]any{
					"q1": "jawaban uji",
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := newJSONRequest(t, tc.method, tc.path, tc.body)
			req.Header.Set("Accept", "application/json")

			resp, err := srv.Test(req)
			if err != nil {
				t.Fatalf("execute request: %v", err)
			}
			defer resp.Body.Close()

			bodyBytes, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Fatalf("read body: %v", err)
			}

			if resp.StatusCode != fiber.StatusUnauthorized {
				t.Fatalf("unexpected status code %d body=%s", resp.StatusCode, string(bodyBytes))
			}

			validateHTTPContract(t, router, req, resp.StatusCode, resp.Header, bodyBytes, false)
		})
	}
}

func TestContract_AuthRouteParity_ValidAgainstOpenAPI(t *testing.T) {
	srv := contractharness.BuildServer(t)
	router := buildOpenAPIRouter(t)

	tests := []struct {
		name       string
		method     string
		path       string
		body       any
		expectCode int
	}{
		{
			name:       "google login invalid token",
			method:     http.MethodPost,
			path:       "/api/v1/auth/google",
			body:       map[string]any{"token": "invalid"},
			expectCode: fiber.StatusUnauthorized,
		},
		{
			name:       "refresh without cookie",
			method:     http.MethodPost,
			path:       "/api/v1/auth/refresh",
			expectCode: fiber.StatusUnauthorized,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := newJSONRequest(t, tc.method, tc.path, tc.body)
			req.Header.Set("Accept", "application/json")

			resp, err := srv.Test(req)
			if err != nil {
				t.Fatalf("execute request: %v", err)
			}
			defer resp.Body.Close()

			bodyBytes, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Fatalf("read body: %v", err)
			}

			if resp.StatusCode != tc.expectCode {
				t.Fatalf("unexpected status code %d body=%s", resp.StatusCode, string(bodyBytes))
			}

			validateHTTPContract(t, router, req, resp.StatusCode, resp.Header, bodyBytes, true)
		})
	}
}

func TestContract_APIV1UnknownRoute_UsesStandardEnvelope(t *testing.T) {
	srv := contractharness.BuildServer(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/not-implemented", nil)
	req.Header.Set("x-request-id", "req-contract-api-v1")

	httpResp, err := srv.Test(req)
	if err != nil {
		t.Fatalf("execute request: %v", err)
	}
	defer httpResp.Body.Close()

	bodyBytes, err := io.ReadAll(httpResp.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}

	var payload map[string]any
	if err := json.Unmarshal(bodyBytes, &payload); err != nil {
		t.Fatalf("parse payload: %v", err)
	}

	httpharness.RequireStatus(t, httpResp.StatusCode, fiber.StatusNotFound)
	httpharness.RequireErrorEnvelope(t, payload, "NOT_FOUND")
}

func buildOpenAPIRouter(t testing.TB) routers.Router {
	t.Helper()

	doc, err := contractopenapi.LoadAndValidate(generatedOpenAPIPath)
	if err != nil {
		t.Fatalf("load openapi: %v", err)
	}

	router, err := legacy.NewRouter(doc)
	if err != nil {
		t.Fatalf("build openapi router: %v", err)
	}
	return router
}

func validateHTTPContract(
	t testing.TB,
	router routers.Router,
	req *http.Request,
	statusCode int,
	header http.Header,
	body []byte,
	validateRequest bool,
) {
	t.Helper()

	route, pathParams, err := router.FindRoute(req)
	if err != nil {
		t.Fatalf("find route in openapi spec: %v", err)
	}

	requestValidation := &openapi3filter.RequestValidationInput{
		Request:    req,
		PathParams: pathParams,
		Route:      route,
	}
	requestValidation.Options = &openapi3filter.Options{
		AuthenticationFunc: func(_ context.Context, _ *openapi3filter.AuthenticationInput) error {
			return nil
		},
	}

	if validateRequest {
		if err := openapi3filter.ValidateRequest(context.Background(), requestValidation); err != nil {
			t.Fatalf("request contract invalid: %v", err)
		}
	}

	responseValidation := &openapi3filter.ResponseValidationInput{
		RequestValidationInput: requestValidation,
		Status:                 statusCode,
		Header:                 header,
	}
	responseValidation.SetBodyBytes(body)

	if err := openapi3filter.ValidateResponse(context.Background(), responseValidation); err != nil {
		t.Fatalf("response contract invalid: %v", err)
	}
}

func newJSONRequest(t testing.TB, method string, path string, body any) *http.Request {
	t.Helper()

	if body == nil {
		return httptest.NewRequest(method, path, nil)
	}

	encoded, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("marshal request body: %v", err)
	}

	req := httptest.NewRequest(method, path, bytes.NewReader(encoded))
	req.Header.Set("Content-Type", "application/json")
	return req
}
