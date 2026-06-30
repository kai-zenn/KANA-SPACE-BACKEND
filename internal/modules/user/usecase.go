package user

import (
	"context"
	"errors"
	"fmt"
	"mime/multipart"
	"strings"
	"time"

	"github.com/google/uuid"
)

type BcryptInterface interface {
  GenerateHashPassword(password string) (string, error)
  CompareHashPassword(hashedPassword, password string) error
}

type JWTInterface interface {
  GenerateToken(userID uuid.UUID, role string) (string, error)
}

type StorageInterface interface {
  UploadPhotoProfile(ctx context.Context, id uuid.UUID, file *multipart.FileHeader) (string, error)
  DeletePhotoProfile(ctx context.Context, fileURL string) error
}

type GoogleClaims struct {
  GoogleID string
  Email    string
  FirstName     string
  LastName      string
  Picture string
}

type GoogleVerifierInterface interface {
  VerifyToken(ctx context.Context, idToken string) (*GoogleClaims, error)
}

type IUserUseCase interface {
  Register(ctx context.Context, req UserRegisterRequest) error
  Login(ctx context.Context, req UserLoginRequest) (*UserLoginResponse, error)
  LoginWithGoogle(ctx context.Context, req GoogleAuthRequest) (*UserLoginResponse, error)
  GetProfileByUsername(ctx context.Context, username string) (*User, error)
  UpgradeToSeller(ctx context.Context, req UpgradeSellerRequest) error
  Update(ctx context.Context, userID uuid.UUID, req UpdateProfileRequest) error
  UpdatePassword(ctx context.Context, userID uuid.UUID, req UpdatePasswordRequest) error
  UpdatePhotoProfile(ctx context.Context, param PhotoUpdate) (string, error)
  FollowUsers(ctx context.Context, param FollowParam) error
  UnfollowUser(ctx context.Context, param FollowParam) error
}

type UserUseCase struct {
  ur IUserRepository
  bcrypt BcryptInterface
  jwtAuth JWTInterface
  storage StorageInterface
  googleVerifier GoogleVerifierInterface
}

func NewUserUseCase(
	ur IUserRepository,
	bcrypt BcryptInterface,
	jwt JWTInterface,
	storage StorageInterface,
	googleVerifier GoogleVerifierInterface,
) IUserUseCase {
	return &UserUseCase{
		ur:            ur,
		bcrypt:       bcrypt,
		jwtAuth:          jwt,
		storage:        storage,
		googleVerifier: googleVerifier,
	}
}

var (
  ErrFollowSelf = errors.New("Tidak dapat mem-follow diri sendiri")
  ErrUserNotFound = errors.New("User yang diikuti tidak ditemukan")
)

func (uc *UserUseCase) Register(ctx context.Context, req UserRegisterRequest) error {
  existingUser, _ := uc.ur.GetProfileByUsername(ctx, req.Username)
  if existingUser != nil {
    return errors.New("Username sudah terdaftar")
  }

  existingEmail, _ := uc.ur.GetProfile(ctx, UserParam{Email: req.Email})
  if existingEmail != nil {
    return errors.New("Email sudah terdaftar")
  }

  hashedPassword, err := uc.bcrypt.GenerateHashPassword(req.Password)
  if err != nil {
    return err
  }

  user := &User{
    ID:        uuid.New(),
    FirstName: req.FirstName,
    LastName:  req.LastName,
    Username:  req.Username,
    Email:     req.Email,
    PhoneNumber: &req.PhoneNumber,
    Password:  &hashedPassword,
    ProfilePhotoLink: "",
    Role: RoleUser,
    CreatedAt: time.Now(),
    UpdatedAt: time.Now(),
  }

  return uc.ur.CreateUser(ctx, user)
}

func (uc *UserUseCase) Login(ctx context.Context, req UserLoginRequest) (*UserLoginResponse, error) {
	user, err := uc.ur.GetProfileByUsername(ctx, req.Username)
	if err != nil {
		return nil, errors.New("Username Salah")
	}

	if user.Password == nil {
		return nil, errors.New("akun ini terdaftar menggunakan Google Sign-In. Silakan login menggunakan Google")
	}

	err = uc.bcrypt.CompareHashPassword(*user.Password, req.Password)
	if err != nil {
		return nil, errors.New("Password Salah")
	}

	token, err := uc.jwtAuth.GenerateToken(user.ID, user.Role)
	if err != nil {
		return nil, err
	}

	return &UserLoginResponse{Token: token}, nil
}

func (uc *UserUseCase) UpgradeToSeller(ctx context.Context, req UpgradeSellerRequest) error {
  user, err := uc.ur.GetByID(ctx, req.UserID)
  if err != nil{
    return errors.New("User Tidak ditemukan")
  }

  if user.Role == RoleSeller || user.Role == RoleAdmin {
    return errors.New("Kamu sudah terdaftar sebagai seller")
  } 

  updates := map[string]interface{}{
    "role": RoleSeller,
    "address": req.Address,
    "shop_name": req.ShopName,
    "updated_at": time.Now(),
  }

  return uc.ur.UpdateUser(ctx, req.UserID, updates)
}

func (uc *UserUseCase) GetProfileByUsername(ctx context.Context, username string) (*User, error) {
	return uc.ur.GetProfileByUsername(ctx, username)
}

func (uc *UserUseCase) Update(ctx context.Context, userID uuid.UUID, req UpdateProfileRequest) error {
	updates := make(map[string]interface{})
	if req.FirstName != nil {
		updates["first_name"] = *req.FirstName
	}
	if req.LastName != nil {
		updates["last_name"] = *req.LastName
	}
	if req.Username != nil {
  	usernameVal := *req.Username
  	if usernameVal == "" {
  		return errors.New("username tidak boleh kosong")
  		}
  
  	existing, _ := uc.ur.GetProfileByUsername(ctx, usernameVal)
  	if existing != nil && existing.ID != userID {
  		return errors.New("username sudah digunakan oleh orang lain")
  		}
    updates["username"] = usernameVal
	}
	if req.PhoneNumber != nil {
		updates["phone_number"] = *req.PhoneNumber
	}
	if req.Address != nil {
		updates["address"] = *req.Address
	}
	if len(updates) == 0 {
		return nil
	}

	updates["updated_at"] = time.Now()
	
	return uc.ur.UpdateUser(ctx, userID, updates)
}

func (uc *UserUseCase) UpdatePassword(ctx context.Context, userID uuid.UUID, req UpdatePasswordRequest) error {
	user, err := uc.ur.GetByID(ctx, userID)
	if err != nil {
		return errors.New("gagal mengambil data user")
	}

	// Antisipasi jika kolom password di DB bernilai nil
	if user.Password == nil {
		return errors.New("akun ini belum memiliki password")
	}

	err = uc.bcrypt.CompareHashPassword(*user.Password, *req.OldPassword)
	if err != nil {
		return errors.New("password lama salah")
	}

	newPasswordHash, err := uc.bcrypt.GenerateHashPassword(*req.NewPassword)
	if err != nil {
		return fmt.Errorf("gagal mengubah password: %w", err)
	}

	updates := map[string]interface{}{
		"password": newPasswordHash,
	}

	err = uc.ur.UpdateUser(ctx, userID, updates)
	if err != nil {
		return fmt.Errorf("gagal memperbarui password: %w", err)
	}

	return nil
}

func (uc *UserUseCase) UpdatePhotoProfile(ctx context.Context, param PhotoUpdate) (string, error) {
  user, err := uc.ur.GetByID(ctx, param.UserID)
	if err != nil {
		return "", errors.New("User Tidak ditemukan")
	}
  
	if user.ProfilePhotoLink != "" && !strings.Contains(user.ProfilePhotoLink, "googleusercontent.com") {
		_ = uc.storage.DeletePhotoProfile(ctx, user.ProfilePhotoLink)
	}
  
	newPhotoLink, err := uc.storage.UploadPhotoProfile(ctx, param.UserID, param.Image)
	if err != nil {
		return "", fmt.Errorf("gagal mengunggah foto profil: %w", err)
	}
  
	err = uc.ur.UpdatePhoto(ctx, param.UserID, newPhotoLink)
	if err != nil {
		return "", fmt.Errorf("gagal memperbarui foto profil: %w", err)
	}
	return newPhotoLink, nil
}

func (uc *UserUseCase) LoginWithGoogle(ctx context.Context, req GoogleAuthRequest) (*UserLoginResponse, error) {
  googleClaims, err := uc.googleVerifier.VerifyToken(ctx, req.IDToken)
  if err != nil{
    return nil, fmt.Errorf("Google authentication failed: %w", err)
  }

  var user *User
  user, _ = uc.ur.GetProfile(ctx, UserParam{Email: googleClaims.Email})
  if user == nil {
    baseUsername := strings.Split(strings.Split(googleClaims.Email, "@")[0], ".")[0]
    username := fmt.Sprintf("%s_%s", baseUsername, uuid.New().String()[:5])
    newUser := User{
      ID:        uuid.New(),
      FirstName: googleClaims.FirstName,
      LastName:  googleClaims.LastName,
      Username:  username,
      Email:     googleClaims.Email,
      Password: nil,
      GoogleID:  &googleClaims.GoogleID,
      ProfilePhotoLink: googleClaims.Picture,
      Role:      RoleUser,
      CreatedAt: time.Now(),
      UpdatedAt: time.Now(),
    }
    err := uc.ur.CreateUser(ctx, &newUser)
    if err != nil {
      return nil, fmt.Errorf("Gagal mendaftarkan user via Google: %w", err)
    }
    user = &newUser
    
  } else if user.GoogleID == nil {
    updates := map[string]interface{}{
      "google_id": googleClaims.GoogleID,
      "updated_at": time.Now(),
    }
    err := uc.ur.UpdateUser(ctx, user.ID, updates)
    if err != nil {
      return nil, fmt.Errorf("Gagal memperbarui user: %w", err)
    }
  }

  token, err := uc.jwtAuth.GenerateToken(user.ID, user.Role)
  if err != nil {
    return nil, fmt.Errorf("Gagal menghasilkan token: %w", err)
  }
  return &UserLoginResponse{Token: token}, nil
}

func (uc *UserUseCase) FollowUsers(ctx context.Context, param FollowParam) error {
  if param.FollowerID == param.FollowingID {
    return ErrFollowSelf
  }
  
  _, err := uc.ur.GetProfile(ctx, UserParam{ID: param.FollowingID})
  if err != nil {
    return ErrUserNotFound
  }
  
  err = uc.ur.FollowUsers(ctx, param.FollowerID, param.FollowingID)
  if err != nil {
    return fmt.Errorf("Gagal mem-follow user: %w", err)
  }
  
  return nil
}

func (uc *UserUseCase) UnfollowUser(ctx context.Context, param FollowParam) error {
  if param.FollowerID == param.FollowingID {
    return ErrFollowSelf
  }

  _, err := uc.ur.GetProfile(ctx, UserParam{ID: param.FollowingID})
  if err != nil {
    return ErrUserNotFound
  }

  err = uc.ur.UnfollowUser(ctx, param.FollowerID, param.FollowingID)
  if err != nil {
    return fmt.Errorf("Gagal mem-unfollow user: %w", err)
  }
  
  return nil
}
