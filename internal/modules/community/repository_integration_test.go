package community

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/recova-app/backend-v2/internal/platform/config"
	"github.com/recova-app/backend-v2/internal/platform/database"
	"github.com/recova-app/backend-v2/internal/platform/database/models"
	databaseharness "github.com/recova-app/backend-v2/test/harness/database"
)

func TestIntegration_Repository_RelationshipQueryAndToggleLike(t *testing.T) {
	databaseURL := databaseharness.RequireDatabaseURLFromEnv(t, "RECOVA_DB_INTEGRATION_URL")
	databaseharness.ResetMigrations(t, databaseURL)

	client, err := database.Connect(integrationConfigCommunity(databaseURL))
	if err != nil {
		t.Fatalf("connect db: %v", err)
	}
	t.Cleanup(func() {
		_ = client.Close()
	})

	repo := NewRepository(client.Gorm())
	ctx := context.Background()

	author := models.User{GoogleID: "google-community-author", Email: "community-author@example.test", Nickname: "author"}
	commenter := models.User{GoogleID: "google-community-commenter", Email: "community-commenter@example.test", Nickname: "commenter"}
	if err := client.Gorm().WithContext(ctx).Create(&author).Error; err != nil {
		t.Fatalf("create author: %v", err)
	}
	if err := client.Gorm().WithContext(ctx).Create(&commenter).Error; err != nil {
		t.Fatalf("create commenter: %v", err)
	}

	streakStart := time.Date(2026, 5, 6, 0, 0, 0, 0, time.UTC)
	if err := client.Gorm().WithContext(ctx).Create(&models.Streak{UserID: author.ID, StartDate: streakStart, IsActive: true}).Error; err != nil {
		t.Fatalf("create streak: %v", err)
	}

	createdPost, err := repo.CreatePost(ctx, models.CommunityPost{
		UserID:   author.ID,
		Title:    ptrString("Test title"),
		Content:  "valid community content for integration test",
		Category: "motivation",
	})
	if err != nil {
		t.Fatalf("create post: %v", err)
	}

	if _, err := repo.CreateCommentAndIncrement(ctx, commenter.ID, createdPost.ID, "valid comment"); err != nil {
		t.Fatalf("create comment: %v", err)
	}

	firstLike, err := repo.ToggleLike(ctx, commenter.ID, createdPost.ID)
	if err != nil {
		t.Fatalf("toggle like first: %v", err)
	}
	if !firstLike.IsLiked || firstLike.LikedCount != 1 {
		t.Fatalf("unexpected first like payload: %+v", firstLike)
	}

	secondLike, err := repo.ToggleLike(ctx, commenter.ID, createdPost.ID)
	if err != nil {
		t.Fatalf("toggle like second: %v", err)
	}
	if secondLike.IsLiked || secondLike.LikedCount != 0 {
		t.Fatalf("unexpected second like payload: %+v", secondLike)
	}

	category := PostCategoryMotivation
	rows, err := repo.ListPosts(ctx, &category)
	if err != nil {
		t.Fatalf("list posts: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(rows))
	}

	row := rows[0]
	if row.CommentCount != 1 {
		t.Fatalf("expected comment_count=1, got %d", row.CommentCount)
	}
	if row.LikeCount != 0 {
		t.Fatalf("expected like_count=0 after second toggle, got %d", row.LikeCount)
	}
	if row.AuthorNickname != "author" {
		t.Fatalf("expected author nickname, got %q", row.AuthorNickname)
	}
	if row.StreakStartDate == nil || !row.StreakStartDate.Equal(streakStart) {
		t.Fatalf("expected streak_start_date=%s, got %#v", streakStart, row.StreakStartDate)
	}
}

func TestIntegration_Repository_CreateReplyAndListThread(t *testing.T) {
	databaseURL := databaseharness.RequireDatabaseURLFromEnv(t, "RECOVA_DB_INTEGRATION_URL")
	databaseharness.ResetMigrations(t, databaseURL)

	client, err := database.Connect(integrationConfigCommunity(databaseURL))
	if err != nil {
		t.Fatalf("connect db: %v", err)
	}
	t.Cleanup(func() {
		_ = client.Close()
	})

	repo := NewRepository(client.Gorm())
	ctx := context.Background()

	author := models.User{GoogleID: "google-community-thread-author", Email: "community-thread-author@example.test", Nickname: "author"}
	replyUser := models.User{GoogleID: "google-community-thread-reply", Email: "community-thread-reply@example.test", Nickname: "reply"}
	if err := client.Gorm().WithContext(ctx).Create(&author).Error; err != nil {
		t.Fatalf("create author: %v", err)
	}
	if err := client.Gorm().WithContext(ctx).Create(&replyUser).Error; err != nil {
		t.Fatalf("create reply user: %v", err)
	}

	post, err := repo.CreatePost(ctx, models.CommunityPost{
		UserID:   author.ID,
		Content:  "post for thread comment integration",
		Category: "story",
	})
	if err != nil {
		t.Fatalf("create post: %v", err)
	}

	root, err := repo.CreateCommentAndIncrement(ctx, author.ID, post.ID, "root comment")
	if err != nil {
		t.Fatalf("create root comment: %v", err)
	}

	reply, err := repo.CreateReplyAndIncrement(ctx, replyUser.ID, post.ID, root.ID, "first reply", 1)
	if err != nil {
		t.Fatalf("create reply: %v", err)
	}
	if reply.ParentCommentID == nil || *reply.ParentCommentID != root.ID {
		t.Fatalf("expected parent comment id %s, got %#v", root.ID, reply.ParentCommentID)
	}

	updatedRoot, err := repo.FindCommentByID(ctx, root.ID)
	if err != nil {
		t.Fatalf("find root comment: %v", err)
	}
	if updatedRoot.ReplyCount != 1 {
		t.Fatalf("expected root reply_count=1, got %d", updatedRoot.ReplyCount)
	}

	thread, err := repo.ListCommentThreadByPostID(ctx, post.ID, 50)
	if err != nil {
		t.Fatalf("list thread: %v", err)
	}
	if len(thread) != 2 {
		t.Fatalf("expected 2 comments in thread, got %d", len(thread))
	}
	if thread[0].Depth != 0 {
		t.Fatalf("expected root depth=0, got %d", thread[0].Depth)
	}
	if thread[1].Depth != 1 {
		t.Fatalf("expected reply depth=1, got %d", thread[1].Depth)
	}

	otherPost, err := repo.CreatePost(ctx, models.CommunityPost{
		UserID:   author.ID,
		Content:  "other post",
		Category: "advice",
	})
	if err != nil {
		t.Fatalf("create other post: %v", err)
	}

	if _, err := repo.CreateReplyAndIncrement(ctx, replyUser.ID, otherPost.ID, root.ID, "must fail", 1); !errors.Is(err, errParentCommentPostMismatch) {
		t.Fatalf("expected errParentCommentPostMismatch, got %v", err)
	}
}

func integrationConfigCommunity(databaseURL string) config.Config {
	return config.Config{
		Database: config.DatabaseConfig{
			URL:                databaseURL,
			MaxOpenConns:       10,
			MaxIdleConns:       5,
			ConnMaxLifetimeSec: 300,
		},
		Observability: config.ObservabilityConfig{
			HealthCheckTimeoutMs: 3000,
		},
	}
}

func ptrString(value string) *string {
	return &value
}
