package community

import (
	"testing"

	"github.com/recova-app/backend-v2/internal/shared/errs"
)

func TestNormalizeCreatePostRequest_InvalidCategory(t *testing.T) {
	_, err := NormalizeCreatePostRequest(CreatePostRequest{
		Content:  "konten yang cukup panjang",
		Category: "invalid",
	})
	if err == nil {
		t.Fatal("expected validation error")
	}
	if errs.Map(err).Code != errs.CodeValidationError {
		t.Fatalf("expected validation error code, got: %s", errs.Map(err).Code)
	}
}

func TestNormalizeListPostsQuery_ValidCategory(t *testing.T) {
	category := "motivation"
	value, err := NormalizeListPostsQuery(ListPostsQuery{Category: &category})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if value == nil || *value != PostCategoryMotivation {
		t.Fatalf("expected motivation category, got: %#v", value)
	}
}

func TestNormalizeCreateCommentRequest_Empty(t *testing.T) {
	_, err := NormalizeCreateCommentRequest(CreateCommentRequest{Content: "   "})
	if err == nil {
		t.Fatal("expected validation error")
	}
	if errs.Map(err).Code != errs.CodeValidationError {
		t.Fatalf("expected validation error code, got: %s", errs.Map(err).Code)
	}
}
