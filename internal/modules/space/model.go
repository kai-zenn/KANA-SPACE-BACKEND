package space

import (
	"mime/multipart"
	"time"

	"github.com/google/uuid"
)

type CreatePostRequest struct {
  ID        uuid.UUID               `json:"-"`
	UserID    uuid.UUID               `json:"-"`
	Content   string                  `form:"content" binding:"required"`
	Tag       string                  `form:"tag" binding:"required,oneof=CariMaterial PajangKarya TipsTrick DapurHijau Diskusi KabarKomunitas Lifestyle"`
	Latitude  *float64                `form:"latitude"`
	Longitude *float64                `form:"longitude"`
	Images    []*multipart.FileHeader `form:"images" binding:"max=4"`
}

type PostAuthor struct {
	ID               uuid.UUID `json:"id"`
	Username         string    `json:"username"`
	ProfilePhotoLink string    `json:"profile_photo_link"`
}

type PostResponse struct {
	ID            uuid.UUID  `json:"id"`
	User          PostAuthor `json:"user"`
	Content       string     `json:"content"`
	Tag           string     `json:"tag"`
	PhotoURLs     []string   `json:"photo_urls"`
	LikeCount     int        `json:"like_count"`
	CommentCount  int        `json:"comment_count"`
	IsLiked       bool       `json:"is_liked"`
	RequestStatus *string    `json:"request_status,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
}

type FeedQueryParam struct {
	Tag    string `form:"tag"`
	Cursor string `form:"cursor"` // RFC3339 timestamp dari created_at post terakhir di page sebelumnya
	Limit  int    `form:"limit"`
}

type FeedResponse struct {
	Posts      []PostResponse `json:"posts"`
	NextCursor *time.Time     `json:"next_cursor,omitempty"`
}

type LikeParam struct {
	UserID uuid.UUID `json:"-"`
	PostID uuid.UUID `json:"-"`
}

type LikeResponse struct {
	LikeCount int  `json:"like_count"`
	IsLiked   bool `json:"is_liked"`
}

type CreateCommentRequest struct {
	UserID  uuid.UUID `json:"-"`
	PostID  uuid.UUID `json:"-"`
	Content string    `json:"content" binding:"required"`
}

type CommentResponse struct {
	ID        uuid.UUID  `json:"id"`
	User      PostAuthor `json:"user"`
	Content   string     `json:"content"`
	CreatedAt time.Time  `json:"created_at"`
}

type CommentsResponse struct {
	Comments   []CommentResponse `json:"comments"`
	NextCursor *time.Time        `json:"next_cursor,omitempty"`
}
