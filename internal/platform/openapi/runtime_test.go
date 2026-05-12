package openapi

import (
	"testing"

	"github.com/gofiber/fiber/v3"
)

func TestRuntimeRouteSet_FiltersAndNormalizes(t *testing.T) {
	routes := []fiber.Route{
		{Method: fiber.MethodGet, Path: "/health/ready"},
		{Method: "TRACE", Path: "/ignored"},
		{Method: fiber.MethodGet, Path: "api/v1/users/:id"},
		{Method: fiber.MethodPost, Path: "/api/v1/auth/login"},
	}

	set := RuntimeRouteSet(routes)
	if len(set) < 2 {
		t.Fatalf("expected runtime route set contains >= 2 entries, got=%d", len(set))
	}

	if _, ok := set[RouteKey{Method: fiber.MethodGet, Path: "/health/ready"}]; !ok {
		t.Fatalf("expected health route in set")
	}
	if _, ok := set[RouteKey{Method: fiber.MethodGet, Path: "/api/v1/users/{id}"}]; !ok {
		t.Fatalf("expected param route normalized in set")
	}
	if _, ok := set[RouteKey{Method: "TRACE", Path: "/ignored"}]; ok {
		t.Fatalf("expected disallowed method filtered out")
	}
}
