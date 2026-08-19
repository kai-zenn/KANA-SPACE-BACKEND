package database

import (
	"KANA-SPACE-BACKEND/internal/modules/lapak"
	"KANA-SPACE-BACKEND/internal/modules/space"
	"KANA-SPACE-BACKEND/internal/modules/user"

	"gorm.io/gorm"
)

func Migrate(db *gorm.DB) error {
  return db.AutoMigrate(
    &user.User{},
    &space.Post{},
    &space.PostImage{},
    &space.Comment{},
    &space.PostLike{},
    &lapak.Category{},
    &lapak.Product{},
    &lapak.ProductImage{},
    &lapak.Transaction{},
  )
}
