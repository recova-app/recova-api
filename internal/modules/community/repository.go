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
			likedCount, err := txRepo.UpdateLikeCountByDelta(ctx, postID, -1)
			if err != nil {
				return err
			}
			result = ToggleLikePayload{LikedCount: likedCount, IsLiked: false}
			return nil
		case IsRecordNotFound(err):
			if err := txRepo.CreateLike(ctx, userID, postID); err != nil {
				if IsUniqueViolation(err) {
					likedCount, cntErr := txRepo.UpdateLikeCountByDelta(ctx, postID, 0)
					if cntErr != nil {
						return cntErr
					}
					result = ToggleLikePayload{LikedCount: likedCount, IsLiked: true}
					return nil
				}
				return err
			}
			likedCount, err := txRepo.UpdateLikeCountByDelta(ctx, postID, 1)
			if err != nil {
				return err
			}
			result = ToggleLikePayload{LikedCount: likedCount, IsLiked: true}
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
