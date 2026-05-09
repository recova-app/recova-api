package community

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/recova-app/backend-v2/internal/platform/database/models"
	"github.com/recova-app/backend-v2/internal/shared/errs"
)

type communityRepository interface {
	FindUserByID(ctx context.Context, userID string) (models.User, error)
	CreatePost(ctx context.Context, row models.CommunityPost) (models.CommunityPost, error)
	ListPosts(ctx context.Context, category *PostCategory) ([]communityPostListRow, error)
	CreateCommentAndIncrement(ctx context.Context, userID string, postID string, content string) (models.CommunityComment, error)
	CreateReplyAndIncrement(ctx context.Context, userID string, postID string, parentCommentID string, content string, depth int16) (models.CommunityComment, error)
	FindCommentByID(ctx context.Context, commentID string) (models.CommunityComment, error)
	ListCommentThreadByPostID(ctx context.Context, postID string, rootLimit int) ([]models.CommunityComment, error)
	ToggleLike(ctx context.Context, userID string, postID string) (ToggleLikePayload, error)
}

// Service owns community business rules.
type Service struct {
	repo communityRepository
	now  func() time.Time
}

// NewService constructs community service.
func NewService(repo communityRepository) *Service {
	return &Service{
		repo: repo,
		now: func() time.Time {
			return time.Now().UTC()
		},
	}
}

// ListPosts returns community feed payload with optional category filter.
func (s *Service) ListPosts(ctx context.Context, query ListPostsQuery) ([]PostPayload, error) {
	category, err := NormalizeListPostsQuery(query)
	if err != nil {
		return nil, err
	}

	rows, err := s.repo.ListPosts(ctx, category)
	if err != nil {
		return nil, errs.New(errs.CodeInternalError, "Gagal membaca postingan komunitas", nil, err)
	}

	payload := make([]PostPayload, 0, len(rows))
	nowUTC := s.now().UTC()
	for _, row := range rows {
		payload = append(payload, PostPayload{
			ID:           row.ID,
			Title:        row.Title,
			Content:      row.Content,
			Category:     row.Category,
			CommentCount: row.CommentCount,
			LikeCount:    row.LikeCount,
			CreatedAt:    formatRFC3339UTC(row.CreatedAt),
			Author: AuthorPayload{
				Nickname:      row.AuthorNickname,
				CurrentStreak: currentStreakDays(nowUTC, row.StreakStartDate),
			},
		})
	}

	return payload, nil
}

// CreatePost stores a community post created by authenticated user.
func (s *Service) CreatePost(ctx context.Context, userID string, req CreatePostRequest) (PostPayload, error) {
	input, err := NormalizeCreatePostRequest(req)
	if err != nil {
		return PostPayload{}, err
	}

	user, err := s.repo.FindUserByID(ctx, userID)
	if err != nil {
		if IsRecordNotFound(err) {
			return PostPayload{}, errs.New(errs.CodeNotFound, "Pengguna tidak ditemukan", nil, err)
		}
		return PostPayload{}, errs.New(errs.CodeInternalError, "Gagal membaca data pengguna", nil, err)
	}

	created, err := s.repo.CreatePost(ctx, models.CommunityPost{
		UserID:   strings.TrimSpace(userID),
		Title:    input.Title,
		Content:  input.Content,
		Category: strings.TrimSpace(string(input.Category)),
	})
	if err != nil {
		return PostPayload{}, errs.New(errs.CodeInternalError, "Gagal menyimpan postingan komunitas", nil, err)
	}

	return PostPayload{
		ID:           created.ID,
		Title:        created.Title,
		Content:      created.Content,
		Category:     created.Category,
		CommentCount: created.CommentCount,
		LikeCount:    created.LikeCount,
		CreatedAt:    formatRFC3339UTC(created.CreatedAt),
		Author: AuthorPayload{
			Nickname:      user.Nickname,
			CurrentStreak: 0,
		},
	}, nil
}

// CreateComment creates comment for one post.
func (s *Service) CreateComment(ctx context.Context, userID string, postID string, req CreateCommentRequest) (CommentPayload, error) {
	input, err := NormalizeCreateCommentRequest(req)
	if err != nil {
		return CommentPayload{}, err
	}

	if _, err := s.repo.FindUserByID(ctx, userID); err != nil {
		if IsRecordNotFound(err) {
			return CommentPayload{}, errs.New(errs.CodeNotFound, "Pengguna tidak ditemukan", nil, err)
		}
		return CommentPayload{}, errs.New(errs.CodeInternalError, "Gagal membaca data pengguna", nil, err)
	}

	created, err := s.repo.CreateCommentAndIncrement(ctx, userID, postID, input.Content)
	if err != nil {
		if IsRecordNotFound(err) {
			return CommentPayload{}, errs.New(errs.CodeNotFound, "Postingan tidak ditemukan", nil, err)
		}
		return CommentPayload{}, errs.New(errs.CodeInternalError, "Gagal menyimpan komentar", nil, err)
	}

	return CommentPayload{
		ID:              created.ID,
		PostID:          created.PostID,
		UserID:          created.UserID,
		ParentCommentID: created.ParentCommentID,
		Content:         created.Content,
		Depth:           int(created.Depth),
		ReplyCount:      created.ReplyCount,
		CreatedAt:       formatRFC3339UTC(created.CreatedAt),
	}, nil
}

// ListCommentThread returns one post comment thread.
func (s *Service) ListCommentThread(ctx context.Context, userID string, postID string, query ListCommentThreadQuery) (CommentThreadPayload, error) {
	if _, err := s.repo.FindUserByID(ctx, userID); err != nil {
		if IsRecordNotFound(err) {
			return CommentThreadPayload{}, errs.New(errs.CodeNotFound, "Pengguna tidak ditemukan", nil, err)
		}
		return CommentThreadPayload{}, errs.New(errs.CodeInternalError, "Gagal membaca data pengguna", nil, err)
	}

	limit, err := NormalizeListCommentThreadQuery(query)
	if err != nil {
		return CommentThreadPayload{}, err
	}

	rows, err := s.repo.ListCommentThreadByPostID(ctx, postID, limit)
	if err != nil {
		return CommentThreadPayload{}, errs.New(errs.CodeInternalError, "Gagal membaca thread komentar", nil, err)
	}

	return CommentThreadPayload{
		PostID:   strings.TrimSpace(postID),
		Comments: buildCommentTree(rows),
	}, nil
}

// CreateReply creates reply for one parent comment.
func (s *Service) CreateReply(
	ctx context.Context,
	userID string,
	postID string,
	parentCommentID string,
	req CreateReplyRequest,
) (ReplyPayload, error) {
	input, err := NormalizeCreateReplyRequest(req)
	if err != nil {
		return ReplyPayload{}, err
	}

	if _, err := s.repo.FindUserByID(ctx, userID); err != nil {
		if IsRecordNotFound(err) {
			return ReplyPayload{}, errs.New(errs.CodeNotFound, "Pengguna tidak ditemukan", nil, err)
		}
		return ReplyPayload{}, errs.New(errs.CodeInternalError, "Gagal membaca data pengguna", nil, err)
	}

	parent, err := s.repo.FindCommentByID(ctx, parentCommentID)
	if err != nil {
		if IsRecordNotFound(err) {
			return ReplyPayload{}, errs.New(errs.CodeNotFound, "Komentar parent tidak ditemukan", nil, err)
		}
		return ReplyPayload{}, errs.New(errs.CodeInternalError, "Gagal membaca komentar parent", nil, err)
	}

	if !strings.EqualFold(strings.TrimSpace(parent.PostID), strings.TrimSpace(postID)) {
		return ReplyPayload{}, errs.New(errs.CodeValidationError, "Komentar parent tidak sesuai dengan postingan", []map[string]string{
			{"field": "commentId", "message": "Komentar parent harus berada pada postingan yang sama"},
		}, nil)
	}

	replyDepth, err := validateReplyDepth(parent.Depth)
	if err != nil {
		return ReplyPayload{}, err
	}

	created, err := s.repo.CreateReplyAndIncrement(ctx, userID, postID, parentCommentID, input.Content, replyDepth)
	if err != nil {
		switch {
		case IsRecordNotFound(err):
			return ReplyPayload{}, errs.New(errs.CodeNotFound, "Postingan atau komentar parent tidak ditemukan", nil, err)
		case errors.Is(err, errParentCommentPostMismatch):
			return ReplyPayload{}, errs.New(errs.CodeValidationError, "Komentar parent tidak sesuai dengan postingan", []map[string]string{
				{"field": "commentId", "message": "Komentar parent harus berada pada postingan yang sama"},
			}, nil)
		default:
			return ReplyPayload{}, errs.New(errs.CodeInternalError, "Gagal menyimpan balasan komentar", nil, err)
		}
	}

	parentID := ""
	if created.ParentCommentID != nil {
		parentID = strings.TrimSpace(*created.ParentCommentID)
	}

	return ReplyPayload{
		ID:              created.ID,
		PostID:          created.PostID,
		UserID:          created.UserID,
		ParentCommentID: parentID,
		Content:         created.Content,
		Depth:           int(created.Depth),
		CreatedAt:       formatRFC3339UTC(created.CreatedAt),
	}, nil
}

// ToggleLike toggles like state for one user on one post.
func (s *Service) ToggleLike(ctx context.Context, userID string, postID string) (ToggleLikePayload, error) {
	if _, err := s.repo.FindUserByID(ctx, userID); err != nil {
		if IsRecordNotFound(err) {
			return ToggleLikePayload{}, errs.New(errs.CodeNotFound, "Pengguna tidak ditemukan", nil, err)
		}
		return ToggleLikePayload{}, errs.New(errs.CodeInternalError, "Gagal membaca data pengguna", nil, err)
	}

	result, err := s.repo.ToggleLike(ctx, userID, postID)
	if err != nil {
		if IsRecordNotFound(err) {
			return ToggleLikePayload{}, errs.New(errs.CodeNotFound, "Postingan tidak ditemukan", nil, err)
		}
		return ToggleLikePayload{}, errs.New(errs.CodeInternalError, "Gagal memperbarui status suka", nil, err)
	}

	return result, nil
}

func currentStreakDays(nowUTC time.Time, startDate *time.Time) int {
	if startDate == nil {
		return 0
	}

	start := startDate.UTC()
	if start.After(nowUTC) {
		return 0
	}

	days := int(nowUTC.Sub(start).Hours()/24) + 1
	if days < 0 {
		return 0
	}
	return days
}

func buildCommentTree(rows []models.CommunityComment) []CommentNodePayload {
	byID := make(map[string]*CommentNodePayload, len(rows))
	order := make([]string, 0, len(rows))
	for _, row := range rows {
		id := strings.TrimSpace(row.ID)
		node := &CommentNodePayload{
			ID:              id,
			PostID:          strings.TrimSpace(row.PostID),
			UserID:          strings.TrimSpace(row.UserID),
			ParentCommentID: row.ParentCommentID,
			Content:         row.Content,
			Depth:           int(row.Depth),
			ReplyCount:      row.ReplyCount,
			CreatedAt:       formatRFC3339UTC(row.CreatedAt),
			Replies:         nil,
		}
		byID[id] = node
		order = append(order, id)
	}

	childrenByParent := make(map[string][]*CommentNodePayload, len(rows))
	roots := make([]*CommentNodePayload, 0)
	for _, id := range order {
		node := byID[id]
		if node == nil {
			continue
		}
		if node.ParentCommentID == nil {
			roots = append(roots, node)
			continue
		}

		parentID := strings.TrimSpace(*node.ParentCommentID)
		_, ok := byID[parentID]
		if !ok {
			roots = append(roots, node)
			continue
		}

		childrenByParent[parentID] = append(childrenByParent[parentID], node)
	}

	out := make([]CommentNodePayload, 0, len(roots))
	for i := range roots {
		out = append(out, rebuildNodeWithReplies(roots[i], childrenByParent))
	}

	return out
}

func rebuildNodeWithReplies(node *CommentNodePayload, childrenByParent map[string][]*CommentNodePayload) CommentNodePayload {
	rebuilt := *node
	children := childrenByParent[strings.TrimSpace(node.ID)]
	rebuilt.Replies = make([]CommentNodePayload, 0, len(children))
	for _, child := range children {
		rebuilt.Replies = append(rebuilt.Replies, rebuildNodeWithReplies(child, childrenByParent))
	}
	return rebuilt
}
