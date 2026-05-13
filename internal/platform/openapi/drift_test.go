package openapi

import "testing"

func TestCompareRouteSets_DetectsMissingRoutes(t *testing.T) {
	runtime := map[RouteKey]struct{}{
		{Method: "GET", Path: "/health/live"}:  {},
		{Method: "GET", Path: "/health/ready"}: {},
	}
	spec := map[RouteKey]struct{}{
		{Method: "GET", Path: "/health/live"}: {},
	}

	result := CompareRouteSets(runtime, spec)
	if !result.HasDrift() {
		t.Fatal("expected drift detected")
	}
	if len(result.MissingInSpec) != 1 {
		t.Fatalf("expected 1 missing-in-spec route, got %d", len(result.MissingInSpec))
	}
	if len(result.MissingInRuntime) != 0 {
		t.Fatalf("expected 0 missing-in-runtime route, got %d", len(result.MissingInRuntime))
	}
}
