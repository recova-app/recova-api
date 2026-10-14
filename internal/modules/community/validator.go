package community

import (
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/recova-app/backend-v2/internal/shared/errs"
)

const (
	minPostContentLength = 10
	maxPostTitleLength   = 120
	maxPostContentLength = 5000
	maxCommentLength     = 2000
	minThreadLimit       = 1
	maxThreadLimit       = 200
	maxThreadDepth       = 2
	minPostLimit         = 1
	maxPostLimit         = 100
	defaultPostLimit     = 20
)

var allowedCategories = map[PostCategory]struct{}{
	PostCategorySaran:      {},
	PostCategoryMotivasi:   {},
	PostCategoryCerita:     {},
	PostCategoryPertanyaan: {},
	PostCategoryBantuan:    {},
}

// NormalizeListPostsQuery validates list-post query and returns normalized category, limit, and cursor.
func NormalizeListPostsQuery(query ListPostsQuery) (*PostCategory, int, *PostCursor, error) {
	var category *PostCategory
	if query.Category != nil {
		cat := PostCategory(strings.ToLower(strings.TrimSpace(*query.Category)))
		if cat != "" {
			if !isAllowedCategory(cat) {
				return nil, 0, nil, errs.New(errs.CodeValidationError, "Kategori postingan tidak valid", []map[string]string{
					{"field": "category", "message": "Kategori harus salah satu dari saran, motivasi, cerita, pertanyaan, bantuan"},
				}, nil)
			}
			category = &cat
		}
	}

	limit := defaultPostLimit
	if query.Limit != nil {
		if *query.Limit < minPostLimit || *query.Limit > maxPostLimit {
			return nil, 0, nil, errs.New(errs.CodeValidationError, "Nilai limit postingan tidak valid", []map[string]string{
				{"field": "limit", "message": "Limit postingan harus di antara 1 sampai 100"},
			}, nil)
		}
		limit = *query.Limit
	}

	var cursor *PostCursor
	if query.Cursor != nil {
		trimmed := strings.TrimSpace(*query.Cursor)
		if trimmed != "" {
			parsedCursor, err := parsePostCursor(trimmed)
			if err != nil {
				return nil, 0, nil, errs.New(errs.CodeValidationError, "Format cursor tidak valid", []map[string]string{
					{"field": "cursor", "message": "Format cursor harus RFC3339 valid atau format <RFC3339>_<id>"},
				}, nil)
			}
			cursor = parsedCursor
		}
	}

	return category, limit, cursor, nil
}

func parsePostCursor(raw string) (*PostCursor, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}

	var timePart, idPart string
	if idx := strings.LastIndex(raw, "_"); idx != -1 {
		timePart = strings.TrimSpace(raw[:idx])
		idPart = strings.TrimSpace(raw[idx+1:])
	} else {
		timePart = raw
	}

	parsed, err := time.Parse(time.RFC3339Nano, timePart)
	if err != nil {
		parsed, err = time.Parse(time.RFC3339, timePart)
	}
	if err != nil {
		return nil, err
	}

	if idPart != "" {
		if _, err := uuid.Parse(idPart); err != nil {
			return nil, err
		}
	}

	return &PostCursor{
		CreatedAt: parsed.UTC(),
		ID:        idPart,
	}, nil
}

// NormalizeCreatePostRequest validates and normalizes create-post payload.
func NormalizeCreatePostRequest(req CreatePostRequest) (CreatePostInput, error) {
	content := strings.TrimSpace(req.Content)
	if content == "" {
		return CreatePostInput{}, errs.New(errs.CodeValidationError, "Konten postingan wajib diisi", []map[string]string{
			{"field": "content", "message": "Konten postingan wajib diisi"},
		}, nil)
	}
	if len([]rune(content)) < minPostContentLength {
		return CreatePostInput{}, errs.New(errs.CodeValidationError, "Konten postingan terlalu pendek", []map[string]string{
			{"field": "content", "message": "Konten postingan minimal 10 karakter"},
		}, nil)
	}
	if len([]rune(content)) > maxPostContentLength {
		return CreatePostInput{}, errs.New(errs.CodeValidationError, "Konten postingan terlalu panjang", []map[string]string{
			{"field": "content", "message": "Konten postingan maksimal 5000 karakter"},
		}, nil)
	}

	category := PostCategory(strings.ToLower(strings.TrimSpace(req.Category)))
	if category == "" {
		return CreatePostInput{}, errs.New(errs.CodeValidationError, "Kategori postingan wajib diisi", []map[string]string{
			{"field": "category", "message": "Kategori postingan wajib diisi"},
		}, nil)
	}
	if !isAllowedCategory(category) {
		return CreatePostInput{}, errs.New(errs.CodeValidationError, "Kategori postingan tidak valid", []map[string]string{
			{"field": "category", "message": "Kategori harus salah satu dari saran, motivasi, cerita, pertanyaan, bantuan"},
		}, nil)
	}

	var title *string
	if req.Title != nil {
		trimmed := strings.TrimSpace(*req.Title)
		if trimmed != "" {
			if len([]rune(trimmed)) > maxPostTitleLength {
				return CreatePostInput{}, errs.New(errs.CodeValidationError, "Judul postingan terlalu panjang", []map[string]string{
					{"field": "title", "message": "Judul postingan maksimal 120 karakter"},
				}, nil)
			}
			title = &trimmed
		}
	}

	return CreatePostInput{
		Title:    title,
		Content:  content,
		Category: category,
	}, nil
}

// NormalizeCreateCommentRequest validates and normalizes create-comment payload.
func NormalizeCreateCommentRequest(req CreateCommentRequest) (CreateCommentInput, error) {
	content := strings.TrimSpace(req.Content)
	if content == "" {
		return CreateCommentInput{}, errs.New(errs.CodeValidationError, "Komentar wajib diisi", []map[string]string{
			{"field": "content", "message": "Komentar wajib diisi"},
		}, nil)
	}
	if len([]rune(content)) > maxCommentLength {
		return CreateCommentInput{}, errs.New(errs.CodeValidationError, "Komentar terlalu panjang", []map[string]string{
			{"field": "content", "message": "Komentar maksimal 2000 karakter"},
		}, nil)
	}

	return CreateCommentInput{Content: content}, nil
}

// NormalizeCreateReplyRequest validates and normalizes create-reply payload.
func NormalizeCreateReplyRequest(req CreateReplyRequest) (CreateReplyInput, error) {
	input, err := NormalizeCreateCommentRequest(CreateCommentRequest{Content: req.Content})
	if err != nil {
		return CreateReplyInput{}, err
	}
	return CreateReplyInput{Content: input.Content}, nil
}

// NormalizeListCommentThreadQuery validates thread query params and returns effective limit.
func NormalizeListCommentThreadQuery(query ListCommentThreadQuery) (int, error) {
	if query.Limit == nil {
		return maxThreadLimit, nil
	}
	limit := *query.Limit
	if limit < minThreadLimit || limit > maxThreadLimit {
		return 0, errs.New(errs.CodeValidationError, "Nilai limit komentar tidak valid", []map[string]string{
			{"field": "limit", "message": "Limit komentar harus di antara 1 sampai 200"},
		}, nil)
	}
	return limit, nil
}

func validateReplyDepth(parentDepth int16) (int16, error) {
	nextDepth := parentDepth + 1
	if int(nextDepth) > maxThreadDepth {
		return 0, errs.New(errs.CodeValidationError, "Balasan komentar melewati batas kedalaman thread", []map[string]string{
			{"field": "comment_id", "message": "Kedalaman balasan komentar maksimal 2"},
		}, nil)
	}
	return nextDepth, nil
}

func isAllowedCategory(category PostCategory) bool {
	_, ok := allowedCategories[category]
	return ok
}
