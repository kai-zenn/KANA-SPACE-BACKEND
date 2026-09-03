package notification

import (
	"time"

	"github.com/google/uuid"
	"KANA-SPACE-BACKEND/internal/modules/user"
)

type Notification struct {
	ID            uuid.UUID  `gorm:"type:uuid;primary_key"`
	UserID        uuid.UUID  `gorm:"type:uuid;not null;index"`
	User          user.User  `gorm:"foreignKey:UserID;references:ID"`
	
	Type          string     `gorm:"type:varchar(50);not null"`
	Title         string     `gorm:"type:varchar(255);not null"`
	Body          string     `gorm:"type:text;not null"`
	
	ReferenceType *string    `gorm:"type:varchar(50)"`
	ReferenceID   *uuid.UUID `gorm:"type:uuid"`
	
	IsRead        bool       `gorm:"default:false"`
	ReadAt        *time.Time
	
	CreatedAt time.Time `gorm:"index:idx_category_created,sort:desc"`
	UpdatedAt time.Time
}

type UserDevice struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key"`
	UserID    uuid.UUID `gorm:"type:uuid;not null;index"`
	User      user.User `gorm:"foreignKey:UserID;references:ID"`
	
	FCMToken  string    `gorm:"type:text;not null;uniqueIndex"`
	Platform  string    `gorm:"type:varchar(20);not null"`

	CreatedAt time.Time
	UpdatedAt time.Time
}
