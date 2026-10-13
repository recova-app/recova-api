package journals

import "testing"

func TestNormalizeCreateJournalRequest_EmptyContent(t *testing.T) {
	_, err := NormalizeCreateJournalRequest(CreateJournalRequest{Content: "   "})
	if err == nil {
		t.Fatal("expected validation error")
	}
}

func TestNormalizeCreateJournalRequest_ContentTooLong(t *testing.T) {
	tooLong := make([]rune, maxJournalLength+1)
	for i := range tooLong {
		tooLong[i] = 'a'
	}

	_, err := NormalizeCreateJournalRequest(CreateJournalRequest{Content: string(tooLong)})
	if err == nil {
		t.Fatal("expected validation error")
	}
}

func TestNormalizeCreateJournalRequest_Success(t *testing.T) {
	input, err := NormalizeCreateJournalRequest(CreateJournalRequest{Content: "  catatan hari ini  "})
	if err != nil {
		t.Fatalf("normalize journal: %v", err)
	}
	if input.Content != "catatan hari ini" {
		t.Fatalf("unexpected normalized content: %q", input.Content)
	}
}

func TestNormalizeListJournalsQuery_DefaultAndBounds(t *testing.T) {
	limit, cursor, err := NormalizeListJournalsQuery(ListJournalsQuery{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if limit != defaultJournalLimit {
		t.Fatalf("expected default limit %d, got %d", defaultJournalLimit, limit)
	}
	if cursor != nil {
		t.Fatalf("expected nil cursor, got %v", cursor)
	}

	validLimit := 50
	cursorStr := "2026-05-10T15:30:00Z"
	limit, cursor, err = NormalizeListJournalsQuery(ListJournalsQuery{
		Limit:  &validLimit,
		Cursor: &cursorStr,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if limit != 50 {
		t.Fatalf("expected limit 50, got %d", limit)
	}
	if cursor == nil || cursor.Format("2006-01-02T15:04:05Z07:00") != cursorStr {
		t.Fatalf("expected parsed cursor %s, got %v", cursorStr, cursor)
	}

	invalidLimitHigh := 101
	_, _, err = NormalizeListJournalsQuery(ListJournalsQuery{Limit: &invalidLimitHigh})
	if err == nil {
		t.Fatal("expected error for limit > 100")
	}

	invalidLimitLow := 0
	_, _, err = NormalizeListJournalsQuery(ListJournalsQuery{Limit: &invalidLimitLow})
	if err == nil {
		t.Fatal("expected error for limit < 1")
	}

	invalidCursor := "not-a-timestamp"
	_, _, err = NormalizeListJournalsQuery(ListJournalsQuery{Cursor: &invalidCursor})
	if err == nil {
		t.Fatal("expected error for invalid cursor")
	}
}
