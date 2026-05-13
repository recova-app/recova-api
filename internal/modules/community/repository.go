package community

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/recova-app/backend-v2/internal/platform/database"
	"github.com/recova-app/backend-v2/internal/platform/database/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const uniqueViolationCode = "23505"

var errParentCommentPostMismatch = errors.New("parent comment does not belong to selected post")

// Repository provides persistence operations for community module.
type Repository struct {
	db *gorm.DB
}

// NewRepository constructs community repository.
func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// DB returns underlying GORM handle.
func (r *Repository) DB() *gorm.DB {
	return r.db
}

// WithTx clones repository bound to transaction handle.
func (r *Repository) WithTx(tx *gorm.DB) *Repository {
	return &Repository{db: tx}
}

// CloneTx clones repository bound to transaction handle as interface type.
func (r *Repository) CloneTx(tx *gorm.DB) communityRepository {
	return r.WithTx(tx)
}

// FindUserByID checks user existence by identifier.
func (r *Repository) FindUserByID(ctx context.Context, userID string) (models.User, error) {
	var user models.User
	if err := r.db.WithContext(ctx).Where("id = ?", strings.TrimSpace(userID)).First(&user).Error; err != nil {
		return models.User{}, err
	}
	return user, nil
}

// CreatePost inserts one community post.
func (r *Repository) CreatePost(ctx context.Context, row models.CommunityPost) (models.CommunityPost, error) {
	if err := r.db.WithContext(ctx).Create(&row).Error; err != nil {
		return models.CommunityPost{}, err
	}
	return row, nil
}

type communityPostListRow struct {
	ID              string
	Title           *string
	Content         string
	Category        string
	CommentCount    int
	LikeCount       int
	CreatedAt       time.Time
	AuthorNickname  string
	StreakStartDate *time.Time
}

// ListPosts returns posts sorted latest first with visible author info.
func (r *Repository) ListPosts(ctx context.Context, category *PostCategory) ([]communityPostListRow, error) {
	query := r.db.WithContext(ctx).
		Table("community_posts AS p").
		Select(`
			p.id,
			p.title,
			p.content,
			p.category,
			p.comment_count,
			p.like_count,
			p.created_at,
			u.nickname AS author_nickname,
			s.start_date AS streak_start_date
		`).
		Joins("JOIN users AS u ON u.id = p.user_id").
		Joins(`LEFT JOIN LATERAL (
			SELECT start_date
			FROM streaks
			WHERE user_id = p.user_id AND is_active = true
			ORDER BY start_date DESC
			LIMIT 1
		) AS s ON true`).
		Order("p.created_at DESC")

	if category != nil {
		query = query.Where("p.category = ?", strings.TrimSpace(string(*category)))
	}

	var rows []communityPostListRow
	if err := query.Scan(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}

// FindPostByID loads one post by identifier.
func (r *Repository) FindPostByID(ctx context.Context, postID string) (models.CommunityPost, error) {
	var post models.CommunityPost
	if err := r.db.WithContext(ctx).Where("id = ?", strings.TrimSpace(postID)).First(&post).Error; err != nil {
		return models.CommunityPost{}, err
	}
	return post, nil
}

// FindPostByIDForUpdate loads one post and acquires row lock for atomic count mutation.
func (r *Repository) FindPostByIDForUpdate(ctx context.Context, postID string) (models.CommunityPost, error) {
	var post models.CommunityPost
	if err := r.db.WithContext(ctx).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("id = ?", strings.TrimSpace(postID)).
		First(&post).Error; err != nil {
		return models.CommunityPost{}, err
	}
	return post, nil
}

// ToggleLike toggles like/unlike state for one post-user tuple atomically.
func (r *Repository) ToggleLike(ctx context.Context, userID string, postID string) (ToggleLikePayload, error) {
	result := ToggleLikePayload{}

	err := database.WithTransaction(ctx, r.db, func(tx *gorm.DB) error {
		txRepo := r.WithTx(tx)

		if _, err := txRepo.FindPostByIDForUpdate(ctx, postID); err != nil {
			return err
		}

		_, err := txRepo.FindLikeByUserAndPost(ctx, userID, postID)
		switch {
		case err == nil:
			if err := txRepo.DeleteLike(ctx, userID, postID); err != nil {
				return err
			}
			liked_count, err := txRepo.UpdateLikeCountByDelta(ctx, postID, -1)
			if err != nil {
				return err
			}
			result = ToggleLikePayload{LikedCount: liked_count, IsLiked: false}
			return nil
		case IsRecordNotFound(err):
			if err := txRepo.CreateLike(ctx, userID, postID); err != nil {
				if IsUniqueViolation(err) {
					liked_count, cntErr := txRepo.UpdateLikeCountByDelta(ctx, postID, 0)
					if cntErr != nil {
						return cntErr
					}
					result = ToggleLikePayload{LikedCount: liked_count, IsLiked: true}
					return nil
				}
				return err
			}
			liked_count, err := txRepo.UpdateLikeCountByDelta(ctx, postID, 1)
			if err != nil {
				return err
			}
			result = ToggleLikePayload{LikedCount: liked_count, IsLiked: true}
			return nil
		default:
			return err
		}
	})
	if err != nil {
		return ToggleLikePayload{}, err
	}
	return result, nil
}

// CreateCommentAndIncrement inserts comment and increments post comment count in one transaction.
func (r *Repository) CreateCommentAndIncrement(ctx context.Context, userID string, postID string, content string) (models.CommunityComment, error) {
	var created models.CommunityComment

	err := database.WithTransaction(ctx, r.db, func(tx *gorm.DB) error {
		txRepo := r.WithTx(tx)

		if _, err := txRepo.FindPostByIDForUpdate(ctx, postID); err != nil {
			return err
		}

		row := models.CommunityComment{
			UserID:  strings.TrimSpace(userID),
			PostID:  strings.TrimSpace(postID),
			Content: strings.TrimSpace(content),
		}
		if err := tx.WithContext(ctx).Create(&row).Error; err != nil {
			return err
		}

		if err := tx.WithContext(ctx).
			Model(&models.CommunityPost{}).
			Where("id = ?", strings.TrimSpace(postID)).
			Updates(map[string]any{
				"comment_count": gorm.Expr("comment_count + 1"),
				"updated_at":    gorm.Expr("now()"),
			}).Error; err != nil {
			return err
		}

		created = row
		return nil
	})
	if err != nil {
		return models.CommunityComment{}, err
	}

	return created, nil
}

// FindCommentByID loads one comment by identifier.
func (r *Repository) FindCommentByID(ctx context.Context, commentID string) (models.CommunityComment, error) {
	var comment models.CommunityComment
	if err := r.db.WithContext(ctx).Where("id = ?", strings.TrimSpace(commentID)).First(&comment).Error; err != nil {
		return models.CommunityComment{}, err
	}
	return comment, nil
}

// FindCommentByIDForUpdate loads one comment and acquires row lock for atomic reply_count mutation.
func (r *Repository) FindCommentByIDForUpdate(ctx context.Context, commentID string) (models.CommunityComment, error) {
	var comment models.CommunityComment
	if err := r.db.WithContext(ctx).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("id = ?", strings.TrimSpace(commentID)).
		First(&comment).Error; err != nil {
		return models.CommunityComment{}, err
	}
	return comment, nil
}

// CreateReplyAndIncrement inserts reply, increments parent reply_count and post comment_count atomically.
func (r *Repository) CreateReplyAndIncrement(
	ctx context.Context,
	userID string,
	postID string,
	parentCommentID string,
	content string,
	depth int16,
) (models.CommunityComment, error) {
	var created models.CommunityComment

	err := database.WithTransaction(ctx, r.db, func(tx *gorm.DB) error {
		txRepo := r.WithTx(tx)

		if _, err := txRepo.FindPostByIDForUpdate(ctx, postID); err != nil {
			return err
		}

		parent, err := txRepo.FindCommentByIDForUpdate(ctx, parentCommentID)
		if err != nil {
			return err
		}
		if !strings.EqualFold(strings.TrimSpace(parent.PostID), strings.TrimSpace(postID)) {
			return errParentCommentPostMismatch
		}

		parentCommentIDNormalized := strings.TrimSpace(parentCommentID)
		row := models.CommunityComment{
			UserID:          strings.TrimSpace(userID),
			PostID:          strings.TrimSpace(postID),
			ParentCommentID: &parentCommentIDNormalized,
			Content:         strings.TrimSpace(content),
			Depth:           depth,
		}
		if err := tx.WithContext(ctx).Create(&row).Error; err != nil {
			return err
		}

		if err := tx.WithContext(ctx).
			Model(&models.CommunityComment{}).
			Where("id = ?", parent.ID).
			Updates(map[string]any{
				"reply_count": gorm.Expr("reply_count + 1"),
			}).Error; err != nil {
			return err
		}

		if err := tx.WithContext(ctx).
			Model(&models.CommunityPost{}).
			Where("id = ?", strings.TrimSpace(postID)).
			Updates(map[string]any{
				"comment_count": gorm.Expr("comment_count + 1"),
				"updated_at":    gorm.Expr("now()"),
			}).Error; err != nil {
			return err
		}

		created = row
		return nil
	})
	if err != nil {
		return models.CommunityComment{}, err
	}

	return created, nil
}

// ListCommentThreadByPostID returns thread comments for one post with deterministic ordering.
func (r *Repository) ListCommentThreadByPostID(ctx context.Context, postID string, rootLimit int) ([]models.CommunityComment, error) {
	type threadRow struct {
		ID              string
		UserID          string
		PostID          string
		ParentCommentID *string
		Content         string
		Depth           int16
		ReplyCount      int
		CreatedAt       time.Time
	}

	var rows []threadRow
	err := r.db.WithContext(ctx).
		Raw(`
			WITH RECURSIVE root_comments AS (
				SELECT id
				FROM community_comments
				WHERE post_id = ? AND parent_comment_id IS NULL
				ORDER BY created_at ASC, id ASC
				LIMIT ?
			),
			thread AS (
				SELECT
					c.id,
					c.user_id,
					c.post_id,
					c.parent_comment_id,
					c.content,
					c.depth,
					c.reply_count,
					c.created_at
				FROM community_comments c
				JOIN root_comments rc ON rc.id = c.id
				UNION ALL
				SELECT
					child.id,
					child.user_id,
					child.post_id,
					child.parent_comment_id,
					child.content,
					child.depth,
					child.reply_count,
					child.created_at
				FROM community_comments child
				JOIN thread parent ON child.parent_comment_id = parent.id
			)
			SELECT
				id,
				user_id,
				post_id,
				parent_comment_id,
				content,
				depth,
				reply_count,
				created_at
			FROM thread
			ORDER BY created_at ASC, id ASC
		`, strings.TrimSpace(postID), rootLimit).
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	out := make([]models.CommunityComment, 0, len(rows))
	for _, row := range rows {
		out = append(out, models.CommunityComment{
			ID:              row.ID,
			UserID:          row.UserID,
			PostID:          row.PostID,
			ParentCommentID: row.ParentCommentID,
			Content:         row.Content,
			Depth:           row.Depth,
			ReplyCount:      row.ReplyCount,
			CreatedAt:       row.CreatedAt,
		})
	}
	return out, nil
}

// FindLikeByUserAndPost loads like state for one post-user tuple.
func (r *Repository) FindLikeByUserAndPost(ctx context.Context, userID string, postID string) (models.CommunityPostLike, error) {
	var like models.CommunityPostLike
	err := r.db.WithContext(ctx).
		Where("user_id = ?", strings.TrimSpace(userID)).
		Where("post_id = ?", strings.TrimSpace(postID)).
		First(&like).Error
	if err != nil {
		return models.CommunityPostLike{}, err
	}
	return like, nil
}

// CreateLike inserts one like row.
func (r *Repository) CreateLike(ctx context.Context, userID string, postID string) error {
	row := models.CommunityPostLike{
		UserID: strings.TrimSpace(userID),
		PostID: strings.TrimSpace(postID),
	}
	return r.db.WithContext(ctx).Create(&row).Error
}

// DeleteLike deletes one like row.
func (r *Repository) DeleteLike(ctx context.Context, userID string, postID string) error {
	return r.db.WithContext(ctx).
		Where("user_id = ?", strings.TrimSpace(userID)).
		Where("post_id = ?", strings.TrimSpace(postID)).
		Delete(&models.CommunityPostLike{}).Error
}

// UpdateLikeCountByDelta mutates like_count atomically and returns latest value.
func (r *Repository) UpdateLikeCountByDelta(ctx context.Context, postID string, delta int) (int, error) {
	var result struct {
		LikeCount int `gorm:"column:like_count"`
	}

	err := r.db.WithContext(ctx).
		Raw(`
			UPDATE community_posts
			SET like_count = GREATEST(like_count + ?, 0),
				updated_at = now()
			WHERE id = ?
			RETURNING like_count
		`, delta, strings.TrimSpace(postID)).
		Scan(&result).Error
	if err != nil {
		return 0, err
	}
	return result.LikeCount, nil
}

// IsRecordNotFound reports gorm record-not-found errors.
func IsRecordNotFound(err error) bool {
	return errors.Is(err, gorm.ErrRecordNotFound)
}

// IsUniqueViolation reports postgres unique violation errors.
func IsUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return false
	}
	return pgErr.Code == uniqueViolationCode
}
