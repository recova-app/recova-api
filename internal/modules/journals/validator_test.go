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
