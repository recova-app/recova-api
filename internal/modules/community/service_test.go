package community

import (
	"context"
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
				Content:         "konten komunitas contoh yang panjang",
				Category:        "motivation",
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

	category := "motivation"
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

func TestService_CreatePost_UserNotFound(t *testing.T) {
	repo := &fakeCommunityRepo{userErr: gorm.ErrRecordNotFound}
	service := NewService(repo)

	_, err := service.CreatePost(context.Background(), "user-1", CreatePostRequest{
		Content:  "konten komunitas valid minimal sepuluh",
		Category: "advice",
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

type fakeCommunityRepo struct {
	userErr   error
	createErr error
	listErr   error
	toggleErr error

	posts []communityPostListRow
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

func (r *fakeCommunityRepo) ListPosts(_ context.Context, _ *PostCategory) ([]communityPostListRow, error) {
	if r.listErr != nil {
		return nil, r.listErr
	}
	return r.posts, nil
}

func (r *fakeCommunityRepo) CreateCommentAndIncrement(_ context.Context, userID string, postID string, content string) (models.CommunityComment, error) {
	return models.CommunityComment{
		ID:        "comment-1",
		UserID:    userID,
		PostID:    postID,
		Content:   content,
		CreatedAt: time.Date(2026, 5, 8, 12, 0, 0, 0, time.UTC),
	}, nil
}

func (r *fakeCommunityRepo) ToggleLike(_ context.Context, _ string, _ string) (ToggleLikePayload, error) {
	if r.toggleErr != nil {
		return ToggleLikePayload{}, r.toggleErr
	}
	return ToggleLikePayload{LikedCount: 1, IsLiked: true}, nil
}
