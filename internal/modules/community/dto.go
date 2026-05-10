package community

import "time"

// PostCategory is allowed classification for community post content.
type PostCategory string

const (
	PostCategorySaran      PostCategory = "saran"
	PostCategoryMotivasi   PostCategory = "motivasi"
	PostCategoryCerita     PostCategory = "cerita"
	PostCategoryPertanyaan PostCategory = "pertanyaan"
	PostCategoryBantuan    PostCategory = "bantuan"
)

// ListPostsQuery captures list-posts query parameters.
type ListPostsQuery struct {
	Category *string `query:"category"`
}

// CreatePostRequest is payload for post creation.
type CreatePostRequest struct {
	Title    *string `json:"title"`
	Content  string  `json:"content"`
	Category string  `json:"category"`
}

// CreateCommentRequest is payload for one post comment.
type CreateCommentRequest struct {
	Content string `json:"content"`
}

// CreateReplyRequest is payload for reply creation on one parent comment.
type CreateReplyRequest struct {
	Content string `json:"content"`
}

// CreatePostInput is normalized create-post payload.
type CreatePostInput struct {
	Title    *string
	Content  string
	Category PostCategory
}

// CreateCommentInput is normalized create-comment payload.
type CreateCommentInput struct {
	Content string
}

// CreateReplyInput is normalized create-reply payload.
type CreateReplyInput struct {
	Content string
}

// ListCommentThreadQuery captures thread query parameters.
type ListCommentThreadQuery struct {
	Limit *int `query:"limit"`
}

// AuthorPayload describes visible public author info on community feed.
type AuthorPayload struct {
	Nickname      string `json:"nickname"`
	CurrentStreak int    `json:"currentStreak"`
}

// PostPayload is one community post payload.
type PostPayload struct {
	ID           string        `json:"id"`
	Title        *string       `json:"title"`
	Content      string        `json:"content"`
	Category     string        `json:"category"`
	CommentCount int           `json:"commentCount"`
	LikeCount    int           `json:"likeCount"`
	CreatedAt    string        `json:"createdAt"`
	Author       AuthorPayload `json:"author"`
}

// CommentPayload is payload for created comment response.
type CommentPayload struct {
	ID              string  `json:"id"`
	PostID          string  `json:"postId"`
	UserID          string  `json:"userId"`
	ParentCommentID *string `json:"parentCommentId"`
	Content         string  `json:"content"`
	Depth           int     `json:"depth"`
	ReplyCount      int     `json:"replyCount"`
	CreatedAt       string  `json:"createdAt"`
}

// CommentThreadPayload is payload for one post comment thread.
type CommentThreadPayload struct {
	PostID   string               `json:"postId"`
	Comments []CommentNodePayload `json:"comments"`
}

// CommentNodePayload is one threaded comment node.
type CommentNodePayload struct {
	ID              string               `json:"id"`
	PostID          string               `json:"postId"`
	UserID          string               `json:"userId"`
	ParentCommentID *string              `json:"parentCommentId"`
	Content         string               `json:"content"`
	Depth           int                  `json:"depth"`
	ReplyCount      int                  `json:"replyCount"`
	CreatedAt       string               `json:"createdAt"`
	Replies         []CommentNodePayload `json:"replies"`
}

// ReplyPayload is payload for one created reply.
type ReplyPayload struct {
	ID              string `json:"id"`
	PostID          string `json:"postId"`
	UserID          string `json:"userId"`
	ParentCommentID string `json:"parentCommentId"`
	Content         string `json:"content"`
	Depth           int    `json:"depth"`
	CreatedAt       string `json:"createdAt"`
}

// ToggleLikePayload is payload for toggle-like response.
type ToggleLikePayload struct {
	LikedCount int  `json:"likedCount"`
	IsLiked    bool `json:"isLiked"`
}

func formatRFC3339UTC(ts time.Time) string {
	return ts.UTC().Format(time.RFC3339)
}
