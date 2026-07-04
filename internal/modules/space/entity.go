package space

import (
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"KANA-SPACE-BACKEND/internal/modules/user"
)

const (
	TagCariMaterial  = "CariMaterial"
	TagPajangKarya   = "PajangKarya"
	TagTipsTrick     = "Tips&Trick"
	TagDapurHijau    = "DapurHijau"
	TagDiskusi       = "Diskusi"
	TagKabarKomunitas = "KabarKomunitas"
	TagLifestyle   = "Lifestyle"
)

const (
	RequestStatusActive = "ACTIVE"
	RequestStatusMatched = "MATCHED"
	RequestStatusClosed  = "CLOSED"
)

type Post struct {
	ID     uuid.UUID `gorm:"primary_key"`
	UserID uuid.UUID `gorm:"type:uuid;not null"`
	Content string `gorm:"type:text;not null"`
	Tag     string `gorm:"type:varchar(30);not null;index:idx_tag_created"`
	PhotoURLs pgtype.FlatArray[string] `gorm:"type:text[]"`
	Latitude  *float64
	Longitude *float64
	// Placeholder NLP
	Embedding      pgtype.FlatArray[float64] `gorm:"type:float8[]"`
	EmbeddingModel string          `gorm:"type:varchar(50)"`
	RequestStatus  *string         `gorm:"type:varchar(20)"` // cuma relevan kalau Tag = CariMaterial
	LikeCount    int `gorm:"not null;default:0"`
	CommentCount int `gorm:"not null;default:0"`
	User     user.User `gorm:"foreignKey:UserID;references:ID"`
	Comments []Comment `gorm:"foreignKey:PostID"`
	CreatedAt time.Time `gorm:"index:idx_tag_created,sort:desc"`
	UpdatedAt time.Time
}

type Comment struct {
	ID     uuid.UUID `gorm:"primary_key"`
	PostID uuid.UUID `gorm:"type:uuid;not null;index"`
	UserID uuid.UUID `gorm:"type:uuid;not null"`
	// ParentCommentID *uuid.UUID `gorm:"index"` // nullable, buat nested reply nanti
	Content string `gorm:"type:text;not null"`
	CreatedAt time.Time
	UpdatedAt time.Time
	User user.User `gorm:"foreignKey:UserID;references:ID"`
}

type PostLike struct {
	ID     uuid.UUID `gorm:"primary_key"`
	PostID uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_post_like_unique"`
	UserID uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_post_like_unique"`
	CreatedAt time.Time
}
