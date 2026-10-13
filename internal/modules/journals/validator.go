package journals

import (
	"strings"
	"time"

	"github.com/recova-app/backend-v2/internal/shared/errs"
)

const (
	maxJournalLength    = 5000
	minJournalLimit     = 1
	maxJournalLimit     = 100
	defaultJournalLimit = 20
)

// NormalizeListJournalsQuery validates list-journals query and returns normalized limit and cursor.
func NormalizeListJournalsQuery(query ListJournalsQuery) (int, *time.Time, error) {
	limit := defaultJournalLimit
	if query.Limit != nil {
		if *query.Limit < minJournalLimit || *query.Limit > maxJournalLimit {
			return 0, nil, errs.New(errs.CodeValidationError, "Nilai limit jurnal tidak valid", []map[string]string{
				{"field": "limit", "message": "Limit jurnal harus di antara 1 sampai 100"},
			}, nil)
		}
		limit = *query.Limit
	}

	var cursorTime *time.Time
	if query.Cursor != nil {
		trimmed := strings.TrimSpace(*query.Cursor)
		if trimmed != "" {
			parsed, err := time.Parse(time.RFC3339Nano, trimmed)
			if err != nil {
				parsed, err = time.Parse(time.RFC3339, trimmed)
			}
			if err != nil {
				return 0, nil, errs.New(errs.CodeValidationError, "Format cursor tidak valid", []map[string]string{
					{"field": "cursor", "message": "Format cursor harus RFC3339 valid"},
				}, nil)
			}
			cursorUTC := parsed.UTC()
			cursorTime = &cursorUTC
		}
	}

	return limit, cursorTime, nil
}

// NormalizeCreateJournalRequest validates and normalizes create-journal payload.
func NormalizeCreateJournalRequest(req CreateJournalRequest) (CreateJournalInput, error) {
	content := strings.TrimSpace(req.Content)
	if content == "" {
		return CreateJournalInput{}, errs.New(errs.CodeValidationError, "Konten jurnal wajib diisi", []map[string]string{
			{"field": "content", "message": "Konten jurnal wajib diisi"},
		}, nil)
	}
	if len([]rune(content)) > maxJournalLength {
		return CreateJournalInput{}, errs.New(errs.CodeValidationError, "Konten jurnal terlalu panjang", []map[string]string{
			{"field": "content", "message": "Konten jurnal maksimal 5000 karakter"},
		}, nil)
	}

	return CreateJournalInput{Content: content}, nil
}
