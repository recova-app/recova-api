package contract

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/getkin/kin-openapi/openapi3filter"
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

	doc, err := contractopenapi.LoadAndValidate(generatedOpenAPIPath)
	if err != nil {
		t.Fatalf("load openapi: %v", err)
	}

	router, err := legacy.NewRouter(doc)
	if err != nil {
		t.Fatalf("build openapi router: %v", err)
	}

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

			route, pathParams, err := router.FindRoute(req)
			if err != nil {
				t.Fatalf("find route in openapi spec: %v", err)
			}

			requestValidation := &openapi3filter.RequestValidationInput{
				Request:    req,
				PathParams: pathParams,
				Route:      route,
			}

			if err := openapi3filter.ValidateRequest(context.Background(), requestValidation); err != nil {
				t.Fatalf("request contract invalid: %v", err)
			}

			responseValidation := &openapi3filter.ResponseValidationInput{
				RequestValidationInput: requestValidation,
				Status:                 resp.StatusCode,
				Header:                 resp.Header,
			}
			responseValidation.SetBodyBytes(bodyBytes)

			if err := openapi3filter.ValidateResponse(context.Background(), responseValidation); err != nil {
				t.Fatalf("response contract invalid: %v", err)
			}
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
