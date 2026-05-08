package community

import (
	"strings"

	"github.com/recova-app/backend-v2/internal/shared/errs"
)

const (
	minPostContentLength = 10
	maxPostTitleLength   = 120
	maxPostContentLength = 5000
	maxCommentLength     = 2000
)

var allowedCategories = map[PostCategory]struct{}{
	PostCategoryAdvice:     {},
	PostCategoryMotivation: {},
	PostCategoryStory:      {},
	PostCategoryQuestion:   {},
	PostCategoryAssistance: {},
}

// NormalizeListPostsQuery validates list-post query and returns normalized category.
func NormalizeListPostsQuery(query ListPostsQuery) (*PostCategory, error) {
	if query.Category == nil {
		return nil, nil
	}

	category := PostCategory(strings.ToLower(strings.TrimSpace(*query.Category)))
	if category == "" {
		return nil, nil
	}
	if !isAllowedCategory(category) {
		return nil, errs.New(errs.CodeValidationError, "Kategori postingan tidak valid", []map[string]string{
			{"field": "category", "message": "Kategori harus salah satu dari advice, motivation, story, question, assistance"},
		}, nil)
	}

	return &category, nil
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
			{"field": "category", "message": "Kategori harus salah satu dari advice, motivation, story, question, assistance"},
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

func isAllowedCategory(category PostCategory) bool {
	_, ok := allowedCategories[category]
	return ok
}
