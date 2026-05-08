package journals

import (
	"strings"

	"github.com/recova-app/backend-v2/internal/shared/errs"
)

const maxJournalLength = 5000

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
