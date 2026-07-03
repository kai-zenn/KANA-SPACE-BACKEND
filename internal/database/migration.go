package database

import (
	"KANA-SPACE-BACKEND/internal/modules/space"
	"KANA-SPACE-BACKEND/internal/modules/user"

	"gorm.io/gorm"
)

func Migrate(db *gorm.DB) error {
  return db.AutoMigrate(
    &user.User{},
    &space.Post{},
    &space.Comment{},
    &space.PostLike{},
  )
}
