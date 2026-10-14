package community

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/recova-app/backend-v2/internal/platform/database/models"
	"github.com/recova-app/backend-v2/internal/shared/errs"
	"gorm.io/gorm"
)

func TestService_ListPosts_Success(t *testing.T) {
	now := time.Date(2026, 5, 8, 12, 0, 0, 0, time.UTC)
	streak := now.AddDate(0, 0, -2)

	repo := &fakeCommunityRepo{
		posts: []communityPostListRow{
			{
				ID:              "post-1",
				Content:         "long sample community content",
				Category:        "motivasi",
				CommentCount:    2,
				LikeCount:       3,
				CreatedAt:       now,
				AuthorNickname:  "user-a",
				StreakStartDate: &streak,
			},
		},
	}
	service := NewService(repo)
	service.now = func() time.Time { return now }

	category := "motivasi"
	result, err := service.ListPosts(context.Background(), ListPostsQuery{Category: &category})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 post, got %d", len(result))
	}
	if result[0].Author.CurrentStreak != 3 {
		t.Fatalf("expected current streak 3, got %d", result[0].Author.CurrentStreak)
	}
}

func TestService_ListPosts_EmptyAndBounded(t *testing.T) {
	repo := &fakeCommunityRepo{posts: []communityPostListRow{}}
	service := NewService(repo)

	limit := 10
	cursorStr := "2026-05-10T00:00:00Z"
	result, err := service.ListPosts(context.Background(), ListPostsQuery{
		Limit:  &limit,
		Cursor: &cursorStr,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 0 {
		t.Fatalf("expected empty array, got %d items", len(result))
	}
	if repo.recordedLimit != 10 {
		t.Fatalf("expected recorded limit 10, got %d", repo.recordedLimit)
	}
	if repo.recordedCursor == nil || repo.recordedCursor.CreatedAt.Format(time.RFC3339) != cursorStr {
		t.Fatalf("expected recorded cursor %s, got %v", cursorStr, repo.recordedCursor)
	}
}

func TestService_ListPosts_CompositeCursor(t *testing.T) {
	repo := &fakeCommunityRepo{posts: []communityPostListRow{}}
	service := NewService(repo)

	cursorStr := "2026-05-10T00:00:00Z_550e8400-e29b-41d4-a716-446655440000"
	_, err := service.ListPosts(context.Background(), ListPostsQuery{
		Cursor: &cursorStr,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.recordedCursor == nil {
		t.Fatal("expected non-nil recorded cursor")
	}
	if repo.recordedCursor.CreatedAt.Format(time.RFC3339) != "2026-05-10T00:00:00Z" {
		t.Fatalf("expected 2026-05-10T00:00:00Z, got %v", repo.recordedCursor.CreatedAt)
	}
	if repo.recordedCursor.ID != "550e8400-e29b-41d4-a716-446655440000" {
		t.Fatalf("expected cursor ID 550e8400-e29b-41d4-a716-446655440000, got %s", repo.recordedCursor.ID)
	}
}

func TestService_CreatePost_UserNotFound(t *testing.T) {
	repo := &fakeCommunityRepo{userErr: gorm.ErrRecordNotFound}
	service := NewService(repo)

	_, err := service.CreatePost(context.Background(), "user-1", CreatePostRequest{
		Content:  "valid community content with enough length",
		Category: "saran",
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if errs.Map(err).Code != errs.CodeNotFound {
		t.Fatalf("expected NOT_FOUND, got %s", errs.Map(err).Code)
	}
}

func TestService_ToggleLike_PostNotFound(t *testing.T) {
	repo := &fakeCommunityRepo{toggleErr: gorm.ErrRecordNotFound}
	service := NewService(repo)

	_, err := service.ToggleLike(context.Background(), "user-1", "post-1")
	if err == nil {
		t.Fatal("expected error")
	}
	if errs.Map(err).Code != errs.CodeNotFound {
		t.Fatalf("expected NOT_FOUND, got %s", errs.Map(err).Code)
	}
}

func TestService_CreateReply_DepthExceeded(t *testing.T) {
	repo := &fakeCommunityRepo{
		parentComment: models.CommunityComment{
			ID:      "comment-parent",
			PostID:  "post-1",
			UserID:  "user-2",
			Content: "parent",
			Depth:   2,
		},
	}
	service := NewService(repo)

	_, err := service.CreateReply(context.Background(), "user-1", "post-1", "comment-parent", CreateReplyRequest{
		Content: "reply",
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if errs.Map(err).Code != errs.CodeValidationError {
		t.Fatalf("expected VALIDATION_ERROR, got %s", errs.Map(err).Code)
	}
}

func TestService_ListCommentThread_Success(t *testing.T) {
	now := time.Date(2026, 5, 8, 12, 0, 0, 0, time.UTC)
	repo := &fakeCommunityRepo{
		threadRows: []models.CommunityComment{
			{
				ID:         "c1",
				PostID:     "post-1",
				UserID:     "user-1",
				Content:    "root",
				Depth:      0,
				ReplyCount: 1,
				CreatedAt:  now,
			},
			{
				ID:              "c2",
				PostID:          "post-1",
				UserID:          "user-2",
				ParentCommentID: ptrStringService("c1"),
				Content:         "reply",
				Depth:           1,
				ReplyCount:      0,
				CreatedAt:       now.Add(time.Minute),
			},
		},
	}
	service := NewService(repo)

	out, err := service.ListCommentThread(context.Background(), "user-1", "post-1", ListCommentThreadQuery{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.PostID != "post-1" {
		t.Fatalf("unexpected post id: %s", out.PostID)
	}
	if len(out.Comments) != 1 {
		t.Fatalf("expected 1 root comment, got %d", len(out.Comments))
	}
	if len(out.Comments[0].Replies) != 1 {
		t.Fatalf("expected 1 reply, got %d", len(out.Comments[0].Replies))
	}
}

type fakeCommunityRepo struct {
	userErr          error
	createErr        error
	createCommentErr error
	listErr          error
	toggleErr        error
	threadErr        error
	parentErr        error
	createReplyErr   error

	posts           []communityPostListRow
	threadRows      []models.CommunityComment
	parentComment   models.CommunityComment
	toggleResponse  ToggleLikePayload
	hasCustomToggle bool
	recordedLimit   int
	recordedCursor  *PostCursor
}

func (r *fakeCommunityRepo) FindUserByID(_ context.Context, _ string) (models.User, error) {
	if r.userErr != nil {
		return models.User{}, r.userErr
	}
	return models.User{ID: "user-1", Nickname: "user-a", Email: "user-a@example.test"}, nil
}

func (r *fakeCommunityRepo) CreatePost(_ context.Context, row models.CommunityPost) (models.CommunityPost, error) {
	if r.createErr != nil {
		return models.CommunityPost{}, r.createErr
	}
	row.ID = "post-1"
	row.CreatedAt = time.Date(2026, 5, 8, 12, 0, 0, 0, time.UTC)
	return row, nil
}

func (r *fakeCommunityRepo) ListPosts(_ context.Context, _ *PostCategory, limit int, cursor *PostCursor) ([]communityPostListRow, error) {
	r.recordedLimit = limit
	r.recordedCursor = cursor
	if r.listErr != nil {
		return nil, r.listErr
	}
	return r.posts, nil
}

func (r *fakeCommunityRepo) CreateCommentAndIncrement(_ context.Context, userID string, postID string, content string) (models.CommunityComment, error) {
	if r.createCommentErr != nil {
		return models.CommunityComment{}, r.createCommentErr
	}
	return models.CommunityComment{
		ID:              "comment-1",
		UserID:          userID,
		PostID:          postID,
		ParentCommentID: nil,
		Content:         content,
		Depth:           0,
		ReplyCount:      0,
		CreatedAt:       time.Date(2026, 5, 8, 12, 0, 0, 0, time.UTC),
	}, nil
}

func (r *fakeCommunityRepo) CreateReplyAndIncrement(_ context.Context, userID string, postID string, parentCommentID string, content string, depth int16) (models.CommunityComment, error) {
	if r.createReplyErr != nil {
		return models.CommunityComment{}, r.createReplyErr
	}
	parentID := parentCommentID
	return models.CommunityComment{
		ID:              "comment-reply-1",
		UserID:          userID,
		PostID:          postID,
		ParentCommentID: &parentID,
		Content:         content,
		Depth:           depth,
		ReplyCount:      0,
		CreatedAt:       time.Date(2026, 5, 8, 12, 0, 0, 0, time.UTC),
	}, nil
}

func (r *fakeCommunityRepo) FindCommentByID(_ context.Context, _ string) (models.CommunityComment, error) {
	if r.parentErr != nil {
		return models.CommunityComment{}, r.parentErr
	}
	if strings.TrimSpace(r.parentComment.ID) == "" {
		return models.CommunityComment{}, gorm.ErrRecordNotFound
	}
	return r.parentComment, nil
}

func (r *fakeCommunityRepo) ListCommentThreadByPostID(_ context.Context, _ string, _ int) ([]models.CommunityComment, error) {
	if r.threadErr != nil {
		return nil, r.threadErr
	}
	return r.threadRows, nil
}

func (r *fakeCommunityRepo) ToggleLike(_ context.Context, _ string, _ string) (ToggleLikePayload, error) {
	if r.toggleErr != nil {
		return ToggleLikePayload{}, r.toggleErr
	}
	if r.hasCustomToggle {
		return r.toggleResponse, nil
	}
	return ToggleLikePayload{LikedCount: 1, IsLiked: true}, nil
}

func ptrStringService(v string) *string {
	return &v
}

func TestService_CreatePost_RepoError(t *testing.T) {
	repo := &fakeCommunityRepo{
		createErr: errors.New("db write failed"),
	}
	service := NewService(repo)

	_, err := service.CreatePost(context.Background(), "user-1", CreatePostRequest{
		Content:  "valid community content with enough length",
		Category: "saran",
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if errs.Map(err).Code != errs.CodeInternalError {
		t.Fatalf("expected INTERNAL_ERROR, got %s", errs.Map(err).Code)
	}
}

func TestService_CreateComment_Success(t *testing.T) {
	repo := &fakeCommunityRepo{}
	service := NewService(repo)

	comment, err := service.CreateComment(context.Background(), "user-1", "post-1", CreateCommentRequest{
		Content: "valid comment content here",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if comment.Content != "valid comment content here" {
		t.Fatalf("unexpected content: %s", comment.Content)
	}
	if comment.PostID != "post-1" {
		t.Fatalf("unexpected post id: %s", comment.PostID)
	}
}

func TestService_CreateComment_UserNotFound(t *testing.T) {
	repo := &fakeCommunityRepo{userErr: gorm.ErrRecordNotFound}
	service := NewService(repo)

	_, err := service.CreateComment(context.Background(), "user-1", "post-1", CreateCommentRequest{
		Content: "valid comment content here",
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if errs.Map(err).Code != errs.CodeNotFound {
		t.Fatalf("expected NOT_FOUND, got %s", errs.Map(err).Code)
	}
}

func TestService_CreateComment_UserRepoError(t *testing.T) {
	repo := &fakeCommunityRepo{userErr: errors.New("db error")}
	service := NewService(repo)

	_, err := service.CreateComment(context.Background(), "user-1", "post-1", CreateCommentRequest{
		Content: "valid comment content here",
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if errs.Map(err).Code != errs.CodeInternalError {
		t.Fatalf("expected INTERNAL_ERROR, got %s", errs.Map(err).Code)
	}
}

func TestService_CreateComment_PostNotFound(t *testing.T) {
	repo := &fakeCommunityRepo{createCommentErr: gorm.ErrRecordNotFound}
	service := NewService(repo)

	_, err := service.CreateComment(context.Background(), "user-1", "post-1", CreateCommentRequest{
		Content: "valid comment content here",
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if errs.Map(err).Code != errs.CodeNotFound {
		t.Fatalf("expected NOT_FOUND, got %s", errs.Map(err).Code)
	}
}

func TestService_CreateComment_RepoError(t *testing.T) {
	repo := &fakeCommunityRepo{createCommentErr: errors.New("db write failed")}
	service := NewService(repo)

	_, err := service.CreateComment(context.Background(), "user-1", "post-1", CreateCommentRequest{
		Content: "valid comment content here",
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if errs.Map(err).Code != errs.CodeInternalError {
		t.Fatalf("expected INTERNAL_ERROR, got %s", errs.Map(err).Code)
	}
}

func TestService_CreateReply_ParentCommentNotFound(t *testing.T) {
	repo := &fakeCommunityRepo{parentErr: gorm.ErrRecordNotFound}
	service := NewService(repo)

	_, err := service.CreateReply(context.Background(), "user-1", "post-1", "parent-1", CreateReplyRequest{
		Content: "valid reply content here",
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if errs.Map(err).Code != errs.CodeNotFound {
		t.Fatalf("expected NOT_FOUND, got %s", errs.Map(err).Code)
	}
}

func TestService_CreateReply_ParentCommentRepoError(t *testing.T) {
	repo := &fakeCommunityRepo{parentErr: errors.New("db read failed")}
	service := NewService(repo)

	_, err := service.CreateReply(context.Background(), "user-1", "post-1", "parent-1", CreateReplyRequest{
		Content: "valid reply content here",
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if errs.Map(err).Code != errs.CodeInternalError {
		t.Fatalf("expected INTERNAL_ERROR, got %s", errs.Map(err).Code)
	}
}

func TestService_CreateReply_ParentPostMismatch(t *testing.T) {
	repo := &fakeCommunityRepo{
		parentComment: models.CommunityComment{
			ID:     "parent-1",
			PostID: "other-post-id",
			Depth:  0,
		},
	}
	service := NewService(repo)

	_, err := service.CreateReply(context.Background(), "user-1", "post-1", "parent-1", CreateReplyRequest{
		Content: "valid reply content here",
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if errs.Map(err).Code != errs.CodeValidationError {
		t.Fatalf("expected VALIDATION_ERROR, got %s", errs.Map(err).Code)
	}
}

func TestService_CreateReply_RepoNotFound(t *testing.T) {
	repo := &fakeCommunityRepo{
		parentComment: models.CommunityComment{
			ID:     "parent-1",
			PostID: "post-1",
			Depth:  0,
		},
		createReplyErr: gorm.ErrRecordNotFound,
	}
	service := NewService(repo)

	_, err := service.CreateReply(context.Background(), "user-1", "post-1", "parent-1", CreateReplyRequest{
		Content: "valid reply content here",
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if errs.Map(err).Code != errs.CodeNotFound {
		t.Fatalf("expected NOT_FOUND, got %s", errs.Map(err).Code)
	}
}

func TestService_CreateReply_RepoError(t *testing.T) {
	repo := &fakeCommunityRepo{
		parentComment: models.CommunityComment{
			ID:     "parent-1",
			PostID: "post-1",
			Depth:  0,
		},
		createReplyErr: errors.New("db error write reply"),
	}
	service := NewService(repo)

	_, err := service.CreateReply(context.Background(), "user-1", "post-1", "parent-1", CreateReplyRequest{
		Content: "valid reply content here",
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if errs.Map(err).Code != errs.CodeInternalError {
		t.Fatalf("expected INTERNAL_ERROR, got %s", errs.Map(err).Code)
	}
}

func TestService_ListCommentThread_UserNotFound(t *testing.T) {
	repo := &fakeCommunityRepo{userErr: gorm.ErrRecordNotFound}
	service := NewService(repo)

	_, err := service.ListCommentThread(context.Background(), "user-1", "post-1", ListCommentThreadQuery{})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if errs.Map(err).Code != errs.CodeNotFound {
		t.Fatalf("expected NOT_FOUND, got %s", errs.Map(err).Code)
	}
}

func TestService_ListCommentThread_UserRepoError(t *testing.T) {
	repo := &fakeCommunityRepo{userErr: errors.New("db error")}
	service := NewService(repo)

	_, err := service.ListCommentThread(context.Background(), "user-1", "post-1", ListCommentThreadQuery{})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if errs.Map(err).Code != errs.CodeInternalError {
		t.Fatalf("expected INTERNAL_ERROR, got %s", errs.Map(err).Code)
	}
}

func TestService_ListCommentThread_InvalidLimit(t *testing.T) {
	repo := &fakeCommunityRepo{}
	service := NewService(repo)

	invalidLimit := 201
	_, err := service.ListCommentThread(context.Background(), "user-1", "post-1", ListCommentThreadQuery{
		Limit: &invalidLimit,
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if errs.Map(err).Code != errs.CodeValidationError {
		t.Fatalf("expected VALIDATION_ERROR, got %s", errs.Map(err).Code)
	}
}

func TestService_ListCommentThread_RepoError(t *testing.T) {
	repo := &fakeCommunityRepo{threadErr: errors.New("db thread error")}
	service := NewService(repo)

	_, err := service.ListCommentThread(context.Background(), "user-1", "post-1", ListCommentThreadQuery{})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if errs.Map(err).Code != errs.CodeInternalError {
		t.Fatalf("expected INTERNAL_ERROR, got %s", errs.Map(err).Code)
	}
}

func TestService_ToggleLike_UserNotFound(t *testing.T) {
	repo := &fakeCommunityRepo{userErr: gorm.ErrRecordNotFound}
	service := NewService(repo)

	_, err := service.ToggleLike(context.Background(), "user-1", "post-1")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if errs.Map(err).Code != errs.CodeNotFound {
		t.Fatalf("expected NOT_FOUND, got %s", errs.Map(err).Code)
	}
}

func TestService_ToggleLike_UserRepoError(t *testing.T) {
	repo := &fakeCommunityRepo{userErr: errors.New("db error")}
	service := NewService(repo)

	_, err := service.ToggleLike(context.Background(), "user-1", "post-1")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if errs.Map(err).Code != errs.CodeInternalError {
		t.Fatalf("expected INTERNAL_ERROR, got %s", errs.Map(err).Code)
	}
}

func TestService_ToggleLike_RepoError(t *testing.T) {
	repo := &fakeCommunityRepo{toggleErr: errors.New("db error")}
	service := NewService(repo)

	_, err := service.ToggleLike(context.Background(), "user-1", "post-1")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if errs.Map(err).Code != errs.CodeInternalError {
		t.Fatalf("expected INTERNAL_ERROR, got %s", errs.Map(err).Code)
	}
}

func TestService_ToggleLike_UnlikeFlow(t *testing.T) {
	repo := &fakeCommunityRepo{
		hasCustomToggle: true,
		toggleResponse: ToggleLikePayload{
			LikedCount: 0,
			IsLiked:    false,
		},
	}
	service := NewService(repo)

	resp, err := service.ToggleLike(context.Background(), "user-1", "post-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.IsLiked {
		t.Fatal("expected is_liked to be false")
	}
	if resp.LikedCount != 0 {
		t.Fatalf("expected liked count 0, got %d", resp.LikedCount)
	}
}

func TestService_ListPosts_RepoError(t *testing.T) {
	repo := &fakeCommunityRepo{listErr: errors.New("db error")}
	service := NewService(repo)

	_, err := service.ListPosts(context.Background(), ListPostsQuery{})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if errs.Map(err).Code != errs.CodeInternalError {
		t.Fatalf("expected INTERNAL_ERROR, got %s", errs.Map(err).Code)
	}
}
