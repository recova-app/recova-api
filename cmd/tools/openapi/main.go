package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	apphttp "github.com/recova-app/backend-v2/internal/app/http"
	authmodule "github.com/recova-app/backend-v2/internal/modules/auth"
	communitymodule "github.com/recova-app/backend-v2/internal/modules/community"
	contentmodule "github.com/recova-app/backend-v2/internal/modules/content"
	educationmodule "github.com/recova-app/backend-v2/internal/modules/education"
	journalsmodule "github.com/recova-app/backend-v2/internal/modules/journals"
	routinemodule "github.com/recova-app/backend-v2/internal/modules/routine"
	usersmodule "github.com/recova-app/backend-v2/internal/modules/users"
	"github.com/recova-app/backend-v2/internal/platform/config"
	contractopenapi "github.com/recova-app/backend-v2/internal/platform/openapi"
)

const (
	openAPISourcePath    = "api/openapi/openapi.yaml"
	openAPIGeneratedPath = "docs/generated/openapi.yaml"
	routeInventoryPath   = "docs/generated/routes.md"
)

func main() {
	cmd := "check"
	if len(os.Args) > 1 {
		cmd = strings.ToLower(strings.TrimSpace(os.Args[1]))
	}

	var err error
	switch cmd {
	case "generate":
		err = runGenerate()
	case "check":
		err = runCheck()
	default:
		err = fmt.Errorf("unknown command %q (supported: generate, check)", cmd)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "openapi %s failed: %v\n", cmd, err)
		os.Exit(1)
	}

	fmt.Printf("openapi %s succeeded\n", cmd)
}

func runGenerate() error {
	if _, err := contractopenapi.LoadAndValidate(openAPISourcePath); err != nil {
		return fmt.Errorf("source spec invalid: %w", err)
	}

	sourceBytes, err := contractopenapi.ReadSpecBytes(openAPISourcePath)
	if err != nil {
		return err
	}

	if err := writeFile(openAPIGeneratedPath, sourceBytes); err != nil {
		return fmt.Errorf("write generated spec: %w", err)
	}

	runtimeRouteSet, err := runtimeRouteSet()
	if err != nil {
		return err
	}

	routesDoc := renderRouteInventory(runtimeRouteSet)
	if err := writeFile(routeInventoryPath, []byte(routesDoc)); err != nil {
		return fmt.Errorf("write route inventory: %w", err)
	}

	return nil
}

func runCheck() error {
	sourceDoc, err := contractopenapi.LoadAndValidate(openAPISourcePath)
	if err != nil {
		return fmt.Errorf("source spec invalid: %w", err)
	}
	generatedDoc, err := contractopenapi.LoadAndValidate(openAPIGeneratedPath)
	if err != nil {
		return fmt.Errorf("generated spec invalid: %w", err)
	}

	sourceBytes, err := contractopenapi.ReadSpecBytes(openAPISourcePath)
	if err != nil {
		return err
	}
	generatedBytes, err := contractopenapi.ReadSpecBytes(openAPIGeneratedPath)
	if err != nil {
		return err
	}

	if !bytes.Equal(sourceBytes, generatedBytes) {
		return errors.New("generated OpenAPI spec is out of sync with source; run `go run ./cmd/tools/openapi generate`")
	}

	runtimeRouteSet, err := runtimeRouteSet()
	if err != nil {
		return err
	}
	specRouteSet := contractopenapi.SpecRouteSet(generatedDoc)
	drift := contractopenapi.CompareRouteSets(runtimeRouteSet, specRouteSet)
	if drift.HasDrift() {
		return fmt.Errorf("route-spec drift detected\n%s", formatDrift(drift))
	}

	if sourceDoc.OpenAPI == "" {
		return errors.New("OpenAPI source must include the `openapi` field")
	}

	expectedRoutesDoc := renderRouteInventory(runtimeRouteSet)
	actualRoutesDoc, err := os.ReadFile(routeInventoryPath)
	if err != nil {
		return fmt.Errorf("read route inventory: %w", err)
	}
	if !bytes.Equal([]byte(expectedRoutesDoc), actualRoutesDoc) {
		return errors.New("route inventory is out of sync; run `go run ./cmd/tools/openapi generate`")
	}

	return nil
}

func runtimeRouteSet() (map[contractopenapi.RouteKey]struct{}, error) {
	cfg := testRuntimeConfig()
	authService := authmodule.NewService(
		authmodule.NewRepository(nil),
		&noopGoogleVerifier{},
		authmodule.NewTokenManager(cfg),
	)
	usersService := usersmodule.NewService(usersmodule.NewRepository(nil), cfg.Application.AppEnv, cfg.Application.NodeEnv)
	routineService := routinemodule.NewService(routinemodule.NewRepository(nil))
	journalsService := journalsmodule.NewService(journalsmodule.NewRepository(nil))
	communityService := communitymodule.NewService(communitymodule.NewRepository(nil))
	educationService := educationmodule.NewService(educationmodule.NewRepository(nil))
	contentService := contentmodule.NewService(contentmodule.NewRepository(nil))

	srv, err := apphttp.NewServer(cfg, slog.New(slog.NewTextHandler(io.Discard, nil)), apphttp.WithModuleDependencies(apphttp.ModuleDependencies{
		AuthService:      authService,
		UsersService:     usersService,
		RoutineService:   routineService,
		JournalsService:  journalsService,
		CommunityService: communityService,
		EducationService: educationService,
		ContentService:   contentService,
	}))
	if err != nil {
		return nil, fmt.Errorf("build runtime server: %w", err)
	}

	routes := srv.Routes(true)
	return contractopenapi.RuntimeRouteSet(routes), nil
}

func testRuntimeConfig() config.Config {
	return config.Config{
		Application: config.ApplicationConfig{
			AppName:   "recova-openapi-check",
			AppEnv:    "test",
			NodeEnv:   "test",
			Port:      "3000",
			APIPrefix: "/api/v1",
		},
		Auth: config.AuthConfig{
			JWTSecret:     "openapi-check-jwt-secret-123456",
			JWTAccessTTL:  15 * time.Minute,
			JWTRefreshTTL: 24 * time.Hour,
			GoogleClient:  "openapi-check-google-client-id",
			Cookie: config.CookieConfig{
				Name:     "recova_refresh_openapi",
				Secure:   false,
				SameSite: "lax",
			},
		},
		Security: config.SecurityConfig{
			CORSOrigins:      []string{"http://localhost:5173"},
			RequestBodyLimit: "1mb",
		},
		Observability: config.ObservabilityConfig{
			RequestIDHeader:      "x-request-id",
			HealthCheckTimeoutMs: 2000,
		},
	}
}

func writeFile(path string, content []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, content, 0o644)
}

func formatDrift(drift contractopenapi.DriftResult) string {
	lines := []string{}
	if len(drift.MissingInSpec) > 0 {
		lines = append(lines, "present in runtime, missing in spec:")
		for _, route := range drift.MissingInSpec {
			lines = append(lines, "- "+route.String())
		}
	}
	if len(drift.MissingInRuntime) > 0 {
		lines = append(lines, "present in spec, missing in runtime:")
		for _, route := range drift.MissingInRuntime {
			lines = append(lines, "- "+route.String())
		}
	}
	return strings.Join(lines, "\n")
}

func renderRouteInventory(routeSet map[contractopenapi.RouteKey]struct{}) string {
	routes := contractopenapi.SortedRouteKeys(routeSet)
	now := time.Date(2026, time.May, 8, 0, 0, 0, 0, time.UTC)

	builder := &strings.Builder{}
	builder.WriteString("---\n")
	builder.WriteString("title: Recova Backend Route Inventory\n")
	builder.WriteString("description: Runtime API route inventory used for contract coverage verification and documentation drift detection.\n")
	builder.WriteString("owner: backend-owner\n")
	builder.WriteString("reviewers:\n")
	builder.WriteString("  - engineering-lead\n")
	builder.WriteString("  - platform-docs-maintainer\n")
	builder.WriteString("doc_status: draft\n")
	builder.WriteString("source_repo: recova-backend-v2\n")
	builder.WriteString("source_path: docs/generated/routes.md\n")
	builder.WriteString("last_reviewed: 2026-05-08\n")
	builder.WriteString("generated_by: cmd-tools-openapi\n")
	builder.WriteString("generated_at: " + now.Format(time.RFC3339) + "\n")
	builder.WriteString("---\n\n")
	builder.WriteString("# Recova Backend Route Inventory\n\n")
	builder.WriteString("This document lists active routes from the current Go Fiber runtime.\n\n")
	builder.WriteString("## Summary\n\n")
	builder.WriteString("| Metric | Value |\n")
	builder.WriteString("| --- | --- |\n")
	builder.WriteString(fmt.Sprintf("| Total routes | %d |\n", len(routes)))
	builder.WriteString("| API prefix | `/api/v1` |\n")
	builder.WriteString("| Last verified | `2026-05-08` |\n\n")
	builder.WriteString("## Registered Routes\n\n")
	builder.WriteString("| Method | Path | Module |\n")
	builder.WriteString("| --- | --- | --- |\n")
	for _, route := range routes {
		builder.WriteString(fmt.Sprintf("| `%s` | `%s` | `%s` |\n", route.Method, route.Path, inferModule(route.Path)))
	}
	builder.WriteString("\n## Drift Check Use\n\n")
	builder.WriteString("Use this file to validate runtime route and OpenAPI contract synchronization in review and CI flows.\n\n")
	builder.WriteString("## Known Gap\n\n")
	builder.WriteString("This route inventory is auto-synced from runtime. Any mismatch with the OpenAPI contract is treated as drift and must be fixed before merge.\n\n")
	builder.WriteString("## Related Documents\n\n")
	builder.WriteString("- [OpenAPI Standard](/Users/macbookpro/Development/recova-backend-v2/docs/standards/openapi.md)\n")
	builder.WriteString("- [API Reference](/Users/macbookpro/Development/recova-backend-v2/docs/api-reference.md)\n")
	builder.WriteString("- [API Docs Generation](/Users/macbookpro/Development/recova-backend-v2/docs/operations/api-docs-generation.md)\n\n")
	builder.WriteString("## Source Reference\n\n")
	builder.WriteString("- [Fiber App API `GetRoutes`](https://docs.gofiber.io/next/api/app/)\n")
	builder.WriteString("- [OpenAPI Specification](https://spec.openapis.org/oas/latest)\n")
	return builder.String()
}

func inferModule(path string) string {
	switch {
	case strings.HasPrefix(path, "/health/"):
		return "health"
	case strings.HasPrefix(path, "/api/v1/"):
		return "api-v1"
	default:
		return "platform"
	}
}

type noopGoogleVerifier struct{}

func (v *noopGoogleVerifier) Verify(_ context.Context, _, _ string) (authmodule.GoogleIdentity, error) {
	return authmodule.GoogleIdentity{}, errors.New("noop verifier")
}
