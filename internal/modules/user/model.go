package user

import (
  "mime/multipart"
  "github.com/google/uuid"
)

const (
  RoleUser    = "user"
  RoleSeller  = "seller"
  RoleAdmin   = "admin"
)

type UserRegisterRequest struct {
  ID uuid.UUID `json:"-"`
  FirstName string `json:"first_name" binding:"required"` 
  LastName string `json:"last_name" binding:"required"` 
  Username string `json:"username" binding:"required"`
  PhoneNumber string `json:"phone_number" binding:"required,numeric,min=11"`
  Email string `json:"email" binding:"required,email"`
  Password string `json:"password" binding:"required,min=8"`
  Role string `json:"role" binding:"required,oneof=user seller"`
}

type UserLoginRequest struct {
  Username string `json:"username" binding:"required"`
  Password string `json:"password" binding:"required,min=8"`
}

type GoogleAuthRequest struct {
  IDToken string `json:"id_token" binding:"required"`
}

type UserLoginResponse struct {
  Token string `json:"token"`
}

type UserParam struct {
	ID       uuid.UUID `json:"-"`
	Username string    `json:"-"`
	Email    string    `json:"-"`
}

type UpdateProfileRequest struct {
	FirstName        *string `json:"first_name"`
	LastName         *string `json:"last_name"`
	Username         *string `json:"username"`
	PhoneNumber      *string `json:"phone_number"`
	Address          *string `json:"address"`
}

type PhotoUpdate struct {
	UserID    uuid.UUID             `json:"-"`
	PhotoLink string                `json:"-"`
	Image     *multipart.FileHeader `form:"image" binding:"required"`
}

type UpgradeSellerRequest struct {
    UserID   uuid.UUID `json:"-"`
    ShopName string    `json:"shop_name" binding:"required"`
    Address  string    `json:"address" binding:"required"`
}

type FollowParam struct {
  FollowerID uuid.UUID `json:"follower_id" binding:"required"`
  FollowingID uuid.UUID `json:"following_id" binding:"required"`
}
