package openapi

import "testing"

func TestFiberPathToOpenAPIPath_ConvertsNamedParams(t *testing.T) {
	got := FiberPathToOpenAPIPath("/api/v1/community/:postId/comments")
	if got != "/api/v1/community/{postId}/comments" {
		t.Fatalf("unexpected converted path: %s", got)
	}
}

func TestNormalizeOpenAPIPath_StripsTrailingSlash(t *testing.T) {
	got := NormalizeOpenAPIPath("/health/live/")
	if got != "/health/live" {
		t.Fatalf("unexpected normalized path: %s", got)
	}
}
