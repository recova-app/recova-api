package community

import "time"

// PostCategory is allowed classification for community post content.
type PostCategory string

const (
	PostCategoryAdvice     PostCategory = "advice"
	PostCategoryMotivation PostCategory = "motivation"
	PostCategoryStory      PostCategory = "story"
	PostCategoryQuestion   PostCategory = "question"
	PostCategoryAssistance PostCategory = "assistance"
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
	ID        string `json:"id"`
	PostID    string `json:"postId"`
	UserID    string `json:"userId"`
	Content   string `json:"content"`
	CreatedAt string `json:"createdAt"`
}

// ToggleLikePayload is payload for toggle-like response.
type ToggleLikePayload struct {
	LikedCount int  `json:"likedCount"`
	IsLiked    bool `json:"isLiked"`
}

func formatRFC3339UTC(ts time.Time) string {
	return ts.UTC().Format(time.RFC3339)
}
