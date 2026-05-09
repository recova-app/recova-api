package community

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/limiter"
	"github.com/golang-jwt/jwt/v5"
	authmodule "github.com/recova-app/backend-v2/internal/modules/auth"
	"github.com/recova-app/backend-v2/internal/platform/database/models"
	"github.com/recova-app/backend-v2/internal/shared/errs"
	"github.com/recova-app/backend-v2/internal/shared/response"
	httpharness "github.com/recova-app/backend-v2/test/harness/http"
	"gorm.io/gorm"
)

func TestRegisterRoutes_Unauthenticated(t *testing.T) {
	authService := buildCommunityAuthService(t, "user-1")
	service := NewService(&communityRouteRepo{})

	app := newCommunityTestApp()
	RegisterRoutes(app.Group("/api/v1/community"), authService, service, nil)

	resp := httpharness.JSONRequest(t, app, fiber.MethodGet, "/api/v1/community", nil, nil)
	httpharness.RequireStatus(t, resp.StatusCode, fiber.StatusUnauthorized)
	httpharness.RequireErrorEnvelope(t, resp.JSON, "UNAUTHENTICATED")
}

func TestRegisterRoutes_CreateValidationError(t *testing.T) {
	authService := buildCommunityAuthService(t, "user-1")
	service := NewService(&communityRouteRepo{})

	app := newCommunityTestApp()
	RegisterRoutes(app.Group("/api/v1/community"), authService, service, nil)

	resp := httpharness.JSONRequest(t, app, fiber.MethodPost, "/api/v1/community", map[string]any{
		"content": "pendek",
	}, map[string]string{
		"Authorization": "Bearer access-token",
	})
	httpharness.RequireStatus(t, resp.StatusCode, fiber.StatusUnprocessableEntity)
	httpharness.RequireErrorEnvelope(t, resp.JSON, "VALIDATION_ERROR")
}

func TestRegisterRoutes_ListSuccess(t *testing.T) {
	authService := buildCommunityAuthService(t, "user-1")
	service := NewService(&communityRouteRepo{
		posts: []communityPostListRow{{
			ID:             "post-1",
			Content:        "konten komunitas contoh yang panjang",
			Category:       "motivation",
			CommentCount:   1,
			LikeCount:      1,
			CreatedAt:      time.Date(2026, 5, 8, 10, 0, 0, 0, time.UTC),
			AuthorNickname: "tester",
		}},
	})

	app := newCommunityTestApp()
	RegisterRoutes(app.Group("/api/v1/community"), authService, service, nil)

	resp := httpharness.JSONRequest(t, app, fiber.MethodGet, "/api/v1/community", nil, map[string]string{
		"Authorization": "Bearer access-token",
	})
	httpharness.RequireStatus(t, resp.StatusCode, fiber.StatusOK)
	httpharness.RequireSuccessEnvelope(t, resp.JSON)
}

func TestRegisterRoutes_CommentSuccess(t *testing.T) {
	authService := buildCommunityAuthService(t, "user-1")
	service := NewService(&communityRouteRepo{})

	app := newCommunityTestApp()
	RegisterRoutes(app.Group("/api/v1/community"), authService, service, nil)

	resp := httpharness.JSONRequest(t, app, fiber.MethodPost, "/api/v1/community/post-1/comments", map[string]any{
		"content": "komentar valid",
	}, map[string]string{
		"Authorization": "Bearer access-token",
	})
	httpharness.RequireStatus(t, resp.StatusCode, fiber.StatusCreated)
	httpharness.RequireSuccessEnvelope(t, resp.JSON)
}

func TestRegisterRoutes_CommentThreadSuccess(t *testing.T) {
	authService := buildCommunityAuthService(t, "user-1")
	service := NewService(&communityRouteRepo{
		threadRows: []models.CommunityComment{
			{
				ID:         "comment-1",
				PostID:     "post-1",
				UserID:     "user-1",
				Content:    "root",
				Depth:      0,
				ReplyCount: 1,
				CreatedAt:  time.Date(2026, 5, 8, 10, 0, 0, 0, time.UTC),
			},
		},
	})

	app := newCommunityTestApp()
	RegisterRoutes(app.Group("/api/v1/community"), authService, service, nil)

	resp := httpharness.JSONRequest(t, app, fiber.MethodGet, "/api/v1/community/post-1/comments", nil, map[string]string{
		"Authorization": "Bearer access-token",
	})
	httpharness.RequireStatus(t, resp.StatusCode, fiber.StatusOK)
	httpharness.RequireSuccessEnvelope(t, resp.JSON)
}

func TestRegisterRoutes_ReplySuccess(t *testing.T) {
	authService := buildCommunityAuthService(t, "user-1")
	service := NewService(&communityRouteRepo{
		parentComment: models.CommunityComment{
			ID:      "comment-1",
			PostID:  "post-1",
			UserID:  "user-2",
			Content: "parent",
			Depth:   0,
		},
	})

	app := newCommunityTestApp()
	RegisterRoutes(app.Group("/api/v1/community"), authService, service, nil)

	resp := httpharness.JSONRequest(t, app, fiber.MethodPost, "/api/v1/community/post-1/comments/comment-1/replies", map[string]any{
		"content": "balasan valid",
	}, map[string]string{
		"Authorization": "Bearer access-token",
	})
	httpharness.RequireStatus(t, resp.StatusCode, fiber.StatusCreated)
	httpharness.RequireSuccessEnvelope(t, resp.JSON)
}

func TestRegisterRoutes_ToggleLikeSuccess(t *testing.T) {
	authService := buildCommunityAuthService(t, "user-1")
	service := NewService(&communityRouteRepo{})

	app := newCommunityTestApp()
	RegisterRoutes(app.Group("/api/v1/community"), authService, service, nil)

	resp := httpharness.JSONRequest(t, app, fiber.MethodPost, "/api/v1/community/post-1/like", nil, map[string]string{
		"Authorization": "Bearer access-token",
	})
	httpharness.RequireStatus(t, resp.StatusCode, fiber.StatusOK)
	httpharness.RequireSuccessEnvelope(t, resp.JSON)
}

func TestRegisterRoutes_WriteRateLimited(t *testing.T) {
	authService := buildCommunityAuthService(t, "user-1")
	service := NewService(&communityRouteRepo{})

	app := newCommunityTestApp()
	writeLimiter := limiter.New(limiter.Config{
		Max:        1,
		Expiration: time.Minute,
		KeyGenerator: func(_ fiber.Ctx) string {
			return "user-1"
		},
		LimitReached: func(_ fiber.Ctx) error {
			return errs.New(errs.CodeRateLimited, "Terlalu banyak permintaan komunitas, coba lagi sebentar", nil, nil)
		},
	})
	RegisterRoutes(app.Group("/api/v1/community"), authService, service, writeLimiter)

	headers := map[string]string{"Authorization": "Bearer access-token"}
	first := httpharness.JSONRequest(t, app, fiber.MethodPost, "/api/v1/community", map[string]any{
		"content":  "konten komunitas valid minimal sepuluh",
		"category": "advice",
	}, headers)
	httpharness.RequireStatus(t, first.StatusCode, fiber.StatusCreated)

	second := httpharness.JSONRequest(t, app, fiber.MethodPost, "/api/v1/community", map[string]any{
		"content":  "konten komunitas valid minimal sepuluh",
		"category": "advice",
	}, headers)
	httpharness.RequireStatus(t, second.StatusCode, fiber.StatusTooManyRequests)
	httpharness.RequireErrorEnvelope(t, second.JSON, "RATE_LIMITED")
}

func buildCommunityAuthService(t testing.TB, userID string) *authmodule.Service {
	t.Helper()

	repo := &communityAuthRepo{user: models.User{ID: userID, Email: "user@example.test", Nickname: "tester"}}
	tokens := &communityAuthTokenProvider{
		claims: authmodule.SessionClaims{RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(5 * time.Minute)),
		}},
	}
	return authmodule.NewService(repo, &communityAuthVerifier{}, tokens)
}

type communityRouteRepo struct {
	posts         []communityPostListRow
	threadRows    []models.CommunityComment
	parentComment models.CommunityComment
	toggleLike    bool
}

func (r *communityRouteRepo) FindUserByID(_ context.Context, _ string) (models.User, error) {
	return models.User{ID: "user-1", Email: "user@example.test", Nickname: "tester"}, nil
}

func (r *communityRouteRepo) CreatePost(_ context.Context, row models.CommunityPost) (models.CommunityPost, error) {
	row.ID = "post-1"
	row.CreatedAt = time.Date(2026, 5, 8, 10, 0, 0, 0, time.UTC)
	return row, nil
}

func (r *communityRouteRepo) ListPosts(_ context.Context, _ *PostCategory) ([]communityPostListRow, error) {
	return r.posts, nil
}

func (r *communityRouteRepo) CreateCommentAndIncrement(_ context.Context, userID string, postID string, content string) (models.CommunityComment, error) {
	return models.CommunityComment{
		ID:              "comment-1",
		UserID:          userID,
		PostID:          postID,
		ParentCommentID: nil,
		Content:         content,
		Depth:           0,
		ReplyCount:      0,
		CreatedAt:       time.Date(2026, 5, 8, 10, 0, 0, 0, time.UTC),
	}, nil
}

func (r *communityRouteRepo) CreateReplyAndIncrement(_ context.Context, userID string, postID string, parentCommentID string, content string, depth int16) (models.CommunityComment, error) {
	parentID := parentCommentID
	return models.CommunityComment{
		ID:              "comment-2",
		UserID:          userID,
		PostID:          postID,
		ParentCommentID: &parentID,
		Content:         content,
		Depth:           depth,
		ReplyCount:      0,
		CreatedAt:       time.Date(2026, 5, 8, 10, 1, 0, 0, time.UTC),
	}, nil
}

func (r *communityRouteRepo) FindCommentByID(_ context.Context, _ string) (models.CommunityComment, error) {
	if strings.TrimSpace(r.parentComment.ID) == "" {
		return models.CommunityComment{}, gorm.ErrRecordNotFound
	}
	return r.parentComment, nil
}

func (r *communityRouteRepo) ListCommentThreadByPostID(_ context.Context, _ string, _ int) ([]models.CommunityComment, error) {
	return r.threadRows, nil
}

func (r *communityRouteRepo) ToggleLike(_ context.Context, _ string, _ string) (ToggleLikePayload, error) {
	r.toggleLike = !r.toggleLike
	if r.toggleLike {
		return ToggleLikePayload{LikedCount: 1, IsLiked: true}, nil
	}
	return ToggleLikePayload{LikedCount: 0, IsLiked: false}, nil
}

type communityAuthRepo struct {
	user models.User
}

func (r *communityAuthRepo) FindOrCreateUserByGoogleIdentity(_ context.Context, _ authmodule.GoogleIdentity) (models.User, error) {
	return r.user, nil
}
func (r *communityAuthRepo) FindUserByID(_ context.Context, _ string) (models.User, error) {
	return r.user, nil
}
func (r *communityAuthRepo) IsOnboardingCompleted(_ context.Context, _ string) (bool, error) {
	return false, nil
}
func (r *communityAuthRepo) CreateRefreshToken(_ context.Context, _ models.AuthRefreshToken) error {
	return nil
}
func (r *communityAuthRepo) GetActiveRefreshTokenByHash(_ context.Context, _ string) (models.AuthRefreshToken, error) {
	return models.AuthRefreshToken{}, nil
}
func (r *communityAuthRepo) RevokeRefreshTokenByID(_ context.Context, _ string, _ time.Time) error {
	return nil
}
func (r *communityAuthRepo) RevokeRefreshTokenByHash(_ context.Context, _ string, _ time.Time) error {
	return nil
}
func (r *communityAuthRepo) RotateRefreshToken(_ context.Context, _ string, _ time.Time, _ models.AuthRefreshToken) error {
	return nil
}

type communityAuthTokenProvider struct {
	claims authmodule.SessionClaims
}

func (p *communityAuthTokenProvider) GoogleAudience() string { return "google-client-id" }
func (p *communityAuthTokenProvider) IssueAccessToken(_ string) (string, authmodule.SessionPayload, error) {
	return "access-token", authmodule.SessionPayload{AccessToken: "access-token", TokenType: "Bearer", ExpiresIn: 900}, nil
}
func (p *communityAuthTokenProvider) IssueRefreshToken(_ string) (string, authmodule.SessionClaims, error) {
	return "refresh-token", authmodule.SessionClaims{RegisteredClaims: jwt.RegisteredClaims{Subject: "user-1", ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour))}}, nil
}
func (p *communityAuthTokenProvider) ParseAccessToken(_ string) (authmodule.SessionClaims, error) {
	return p.claims, nil
}
func (p *communityAuthTokenProvider) ParseRefreshToken(_ string) (authmodule.SessionClaims, error) {
	return authmodule.SessionClaims{RegisteredClaims: jwt.RegisteredClaims{Subject: "user-1", ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour))}}, nil
}
func (p *communityAuthTokenProvider) HashRefreshToken(_ string) string { return "hash-refresh" }
func (p *communityAuthTokenProvider) RefreshCookie(rawToken string) *fiber.Cookie {
	return &fiber.Cookie{Name: "refresh", Value: rawToken}
}
func (p *communityAuthTokenProvider) ExpiredRefreshCookie() *fiber.Cookie {
	return &fiber.Cookie{Name: "refresh", Value: "", MaxAge: -1}
}
func (p *communityAuthTokenProvider) RefreshCookieValue(_ fiber.Ctx) string { return "refresh-token" }

type communityAuthVerifier struct{}

func (v *communityAuthVerifier) Verify(_ context.Context, _, _ string) (authmodule.GoogleIdentity, error) {
	return authmodule.GoogleIdentity{}, errors.New("not implemented")
}

func newCommunityTestApp() *fiber.App {
	return fiber.New(fiber.Config{
		ErrorHandler: func(c fiber.Ctx, err error) error {
			mapped := errs.Map(err)
			return c.Status(mapped.Status).JSON(response.Error(mapped.Message, string(mapped.Code), mapped.Details, ""))
		},
	})
}
