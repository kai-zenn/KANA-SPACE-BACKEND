package user

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
  ID        uuid.UUID `gorm:"primary_key"`
  FirstName string    `gorm:"not null"`
  LastName  string    `gorm:"not null"`
  Username  string    `gorm:"not null;unique"`
  Email     string    `gorm:"type:varchar(100);not null;unique"`
  PhoneNumber *string `gorm:"type:varchar(20);default:null"`
  Password  *string    `gorm:"type:varchar(255);default:null"`
  Address *string `gorm:"type:text;default:null"`
  GoogleID *string    `gorm:"unique;default:null"`
  ProfilePhotoLink string    `gorm:"default:''"`
  Role             string    `gorm:"type:varchar(20);not null;default:'user'"`
  Following []*User `gorm:"many2many:user_follows;joinForeignKey:follower_id;joinReferences:following_id"`
  Followers []*User `gorm:"many2many:user_follows;joinForeignKey:following_id;joinReferences:follower_id"`
  CreatedAt time.Time
	UpdatedAt time.Time
}
