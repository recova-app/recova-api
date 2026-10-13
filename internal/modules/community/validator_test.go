package community

import (
	"testing"
	"time"

	"github.com/recova-app/backend-v2/internal/shared/errs"
)

func TestNormalizeCreatePostRequest_InvalidCategory(t *testing.T) {
	_, err := NormalizeCreatePostRequest(CreatePostRequest{
		Content:  "content that is long enough",
		Category: "invalid",
	})
	if err == nil {
		t.Fatal("expected validation error")
	}
	if errs.Map(err).Code != errs.CodeValidationError {
		t.Fatalf("expected validation error code, got: %s", errs.Map(err).Code)
	}
}

func TestNormalizeCreatePostRequest_LegacyEnglishCategoryRejected(t *testing.T) {
	_, err := NormalizeCreatePostRequest(CreatePostRequest{
		Content:  "content that is long enough",
		Category: "motivation",
	})
	if err == nil {
		t.Fatal("expected validation error")
	}
	if errs.Map(err).Code != errs.CodeValidationError {
		t.Fatalf("expected validation error code, got: %s", errs.Map(err).Code)
	}
}

func TestNormalizeListPostsQuery_ValidCategory(t *testing.T) {
	category := "motivasi"
	value, limit, cursor, err := NormalizeListPostsQuery(ListPostsQuery{Category: &category})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if value == nil || *value != PostCategoryMotivasi {
		t.Fatalf("expected motivasi category, got: %#v", value)
	}
	if limit != defaultPostLimit {
		t.Fatalf("expected default limit %d, got %d", defaultPostLimit, limit)
	}
	if cursor != nil {
		t.Fatalf("expected nil cursor, got %v", cursor)
	}
}

func TestNormalizeListPostsQuery_LimitAndCursor(t *testing.T) {
	validLimit := 50
	cursorStr := "2026-05-10T12:00:00Z"
	_, limit, cursor, err := NormalizeListPostsQuery(ListPostsQuery{
		Limit:  &validLimit,
		Cursor: &cursorStr,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if limit != 50 {
		t.Fatalf("expected limit 50, got %d", limit)
	}
	if cursor == nil || cursor.Format(time.RFC3339) != cursorStr {
		t.Fatalf("expected parsed cursor %s, got %v", cursorStr, cursor)
	}

	invalidLimitHigh := 101
	_, _, _, err = NormalizeListPostsQuery(ListPostsQuery{Limit: &invalidLimitHigh})
	if err == nil {
		t.Fatal("expected error for limit > 100")
	}

	invalidLimitLow := 0
	_, _, _, err = NormalizeListPostsQuery(ListPostsQuery{Limit: &invalidLimitLow})
	if err == nil {
		t.Fatal("expected error for limit < 1")
	}

	invalidCursor := "invalid-date"
	_, _, _, err = NormalizeListPostsQuery(ListPostsQuery{Cursor: &invalidCursor})
	if err == nil {
		t.Fatal("expected error for malformed cursor")
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

func TestNormalizeListCommentThreadQuery_InvalidLimit(t *testing.T) {
	limit := 201
	_, err := NormalizeListCommentThreadQuery(ListCommentThreadQuery{Limit: &limit})
	if err == nil {
		t.Fatal("expected validation error")
	}
	if errs.Map(err).Code != errs.CodeValidationError {
		t.Fatalf("expected validation error code, got: %s", errs.Map(err).Code)
	}
}

func TestValidateReplyDepth_MaxDepthExceeded(t *testing.T) {
	_, err := validateReplyDepth(2)
	if err == nil {
		t.Fatal("expected validation error")
	}
	if errs.Map(err).Code != errs.CodeValidationError {
		t.Fatalf("expected validation error code, got: %s", errs.Map(err).Code)
	}
}
