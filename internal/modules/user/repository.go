package user

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type IUserRepository interface {
  CreateUser(ctx context.Context, user *User) error
  GetProfileByUsername(ctx context.Context, username string) (*User, error)
  GetProfile(ctx context.Context, param UserParam) (*User, error)
  GetByID(ctx context.Context, userID uuid.UUID) (*User, error)
  UpdateUser(ctx context.Context, userID uuid.UUID, updates map[string]interface{}) error
  UpdatePhoto(ctx context.Context, userID uuid.UUID, photoLink string) error
  FollowUsers(ctx context.Context, followerID, followingID uuid.UUID) error
  UnfollowUser(ctx context.Context, followerID, followingID uuid.UUID) error
}

type UserRepository struct {
  db *gorm.DB
}

func NewUserRepository(db *gorm.DB) IUserRepository {
	return &UserRepository{
		db: db,
	}
}

func (ur *UserRepository) CreateUser(ctx context.Context, user *User) error {
  return ur.db.WithContext(ctx).Create(user).Error
}

func (ur *UserRepository) GetProfileByUsername(ctx context.Context, username string) (*User, error) {
	var user User

	err := ur.db.WithContext(ctx).Where("username = ?", username).First(&user).Error
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (ur *UserRepository) GetProfile(ctx context.Context, param UserParam) (*User, error) {
	var user User

	err := ur.db.WithContext(ctx).Where(&param).First(&user).Error
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (ur *UserRepository) GetByID(ctx context.Context, userID uuid.UUID) (*User, error) {
	var user User

	err := ur.db.WithContext(ctx).Where("id = ?", userID).First(&user).Error
	if err != nil {
		return nil, err
	}

	return &user, nil
}


func (ur *UserRepository) UpdateUser(ctx context.Context, userID uuid.UUID, updates map[string]interface{}) error {
	err := ur.db.WithContext(ctx).Model(&User{}).Where("id = ?", userID).Updates(updates).Error
	if err != nil {
		return err
	}
	return nil
}

func (ur *UserRepository) UpdatePhoto(ctx context.Context, userID uuid.UUID, photoLink string) error {
  err := ur.db.WithContext(ctx).Model(&User{}).Where("id = ?", userID).Update("profile_photo_link", photoLink).Error
  if err != nil {
  	return err
  }
  return nil
}

func (ur *UserRepository) FollowUsers(ctx context.Context, followerID, followingID uuid.UUID) error {
  var following User
  var follower User
  
  err := ur.db.WithContext(ctx).First(&following, "id = ?", followingID).Error
  if err != nil {
  	return err
  }
  err = ur.db.WithContext(ctx).First(&follower, "id = ?", followerID).Error
  if err != nil {
  	return err
  }
  
  return ur.db.WithContext(ctx).Model(&follower).Association("Following").Append(&following)
}

func (ur *UserRepository) UnfollowUser(ctx context.Context, followerID, followingID uuid.UUID) error {
  var following User
  var follower User
  
  if err := ur.db.WithContext(ctx).First(&following, "id = ?", followingID).Error; err != nil {
    return err
  }
  if err := ur.db.WithContext(ctx).First(&follower, "id = ?", followerID).Error; err != nil {
    return err
  }
  
  return ur.db.WithContext(ctx).Model(&follower).Association("Following").Delete(&following)
}
