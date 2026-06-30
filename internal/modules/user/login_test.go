package user

import (
	"context"
	"errors"
	"mime/multipart"
	"testing"

	"github.com/google/uuid"
)

type mockLoginUserRepository struct {
	GetProfileByUsernameFunc func(ctx context.Context, username string) (*User, error)
}

func (m *mockLoginUserRepository) CreateUser(ctx context.Context, user *User) error { return nil }
func (m *mockLoginUserRepository) GetProfileByUsername(ctx context.Context, username string) (*User, error) {
	return m.GetProfileByUsernameFunc(ctx, username)
}
func (m *mockLoginUserRepository) GetProfile(ctx context.Context, param UserParam) (*User, error) {
	return nil, nil
}
func (m *mockLoginUserRepository) GetByID(ctx context.Context, userID uuid.UUID) (*User, error) {
	return nil, nil
}
func (m *mockLoginUserRepository) UpdateUser(ctx context.Context, userID uuid.UUID, updates map[string]interface{}) error {
	return nil
}
func (m *mockLoginUserRepository) UpdatePhoto(ctx context.Context, userID uuid.UUID, photoLink string) error {
	return nil
}
func (m *mockLoginUserRepository) FollowUsers(ctx context.Context, followerID, followingID uuid.UUID) error {
	return nil
}
func (m *mockLoginUserRepository) UnfollowUser(ctx context.Context, followerID, followingID uuid.UUID) error {
	return nil
}

type mockLoginBcrypt struct {
	CompareHashPasswordFunc func(hashedPassword, password string) error
}

func (m *mockLoginBcrypt) GenerateHashPassword(password string) (string, error) { return "", nil }
func (m *mockLoginBcrypt) CompareHashPassword(hashedPassword, password string) error {
	return m.CompareHashPasswordFunc(hashedPassword, password)
}

type mockLoginJWT struct {
	GenerateTokenFunc func(userID uuid.UUID, role string) (string, error)
}

func (m *mockLoginJWT) GenerateToken(userID uuid.UUID, role string) (string, error) {
	return m.GenerateTokenFunc(userID, role)
}

// Stub kosong penunjang constructor UseCase
type mockLoginStorage struct{}
func (m *mockLoginStorage) UploadPhotoProfile(ctx context.Context, id uuid.UUID, file *multipart.FileHeader) (string, error) { return "", nil }
func (m *mockLoginStorage) DeletePhotoProfile(ctx context.Context, fileURL string) error { return nil }

type mockLoginGoogleVerifier struct{}
func (m *mockLoginGoogleVerifier) VerifyToken(ctx context.Context, idToken string) (*GoogleClaims, error) { return nil, nil }

func TestUserUseCase_Login(t *testing.T) {
	userID := uuid.New()
	hashedPassword := "hashed_super_secret"
	
	validRequest := UserLoginRequest{
		Username: "kanadev",
		Password: "password123",
	}

	tests := []struct {
		name          string
		req           UserLoginRequest
		setupMocks    func(repo *mockLoginUserRepository, bc *mockLoginBcrypt, jwt *mockLoginJWT)
		expectedToken string
		expectedError string
	}{
		{
			name: "Sukses Login",
			req:  validRequest,
			setupMocks: func(repo *mockLoginUserRepository, bc *mockLoginBcrypt, jwt *mockLoginJWT) {
				repo.GetProfileByUsernameFunc = func(ctx context.Context, username string) (*User, error) {
					return &User{
						ID:       userID,
						Username: "kanadev",
						Password: &hashedPassword,
						Role:     RoleUser,
					}, nil
				}
				bc.CompareHashPasswordFunc = func(hashedPassword, password string) error {
					return nil // Password cocok
				}
				jwt.GenerateTokenFunc = func(id uuid.UUID, role string) (string, error) {
					return "valid_jwt_token_string", nil
				}
			},
			expectedToken: "valid_jwt_token_string",
			expectedError: "",
		},
		{
			name: "Gagal - Username Salah / Tidak Ditemukan",
			req:  validRequest,
			setupMocks: func(repo *mockLoginUserRepository, bc *mockLoginBcrypt, jwt *mockLoginJWT) {
				repo.GetProfileByUsernameFunc = func(ctx context.Context, username string) (*User, error) {
					return nil, errors.New("gorm: record not found") // DB melempar error kosong
				}
			},
			expectedToken: "",
			expectedError: "Username Salah",
		},
		{
			name: "Gagal - Akun Terdaftar Lewat Google OAuth",
			req:  validRequest,
			setupMocks: func(repo *mockLoginUserRepository, bc *mockLoginBcrypt, jwt *mockLoginJWT) {
				repo.GetProfileByUsernameFunc = func(ctx context.Context, username string) (*User, error) {
					return &User{
						ID:       userID,
						Username: "kanadev",
						Password: nil, // Password nil berarti register via Google
					}, nil
				}
			},
			expectedToken: "",
			expectedError: "akun ini terdaftar menggunakan Google Sign-In. Silakan login menggunakan Google",
		},
		{
			name: "Gagal - Password Salah",
			req:  validRequest,
			setupMocks: func(repo *mockLoginUserRepository, bc *mockLoginBcrypt, jwt *mockLoginJWT) {
				repo.GetProfileByUsernameFunc = func(ctx context.Context, username string) (*User, error) {
					return &User{
						ID:       userID,
						Username: "kanadev",
						Password: &hashedPassword,
					}, nil
				}
				bc.CompareHashPasswordFunc = func(hashedPassword, password string) error {
					return errors.New("crypto/bcrypt: hashedPassword is not the password") // Bcrypt mismatch
				}
			},
			expectedToken: "",
			expectedError: "Password Salah",
		},
		{
			name: "Gagal - JWT Generation Error",
			req:  validRequest,
			setupMocks: func(repo *mockLoginUserRepository, bc *mockLoginBcrypt, jwt *mockLoginJWT) {
				repo.GetProfileByUsernameFunc = func(ctx context.Context, username string) (*User, error) {
					return &User{
						ID:       userID,
						Username: "kanadev",
						Password: &hashedPassword,
						Role:     RoleUser,
					}, nil
				}
				bc.CompareHashPasswordFunc = func(hashedPassword, password string) error {
					return nil
				}
				jwt.GenerateTokenFunc = func(id uuid.UUID, role string) (string, error) {
					return "", errors.New("jwt signing failed") // Key bermasalah
				}
			},
			expectedToken: "",
			expectedError: "jwt signing failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Inisialisasi Mocking khusus login
			mockRepo := &mockLoginUserRepository{}
			mockBc := &mockLoginBcrypt{}
			mockJw := &mockLoginJWT{}
			mockStore := &mockLoginStorage{}
			mockGVerify := &mockLoginGoogleVerifier{}

			tt.setupMocks(mockRepo, mockBc, mockJw)

			// Masukkan mocks ke usecase
			uc := NewUserUseCase(mockRepo, mockBc, mockJw, mockStore, mockGVerify)

			// Eksekusi fungsi Login
			res, err := uc.Login(context.Background(), tt.req)

			// Validasi Ekspektasi Error
			if tt.expectedError != "" {
				if err == nil {
					t.Fatalf("Ekspektasi error '%s', tetapi fungsi mengembalikan nil", tt.expectedError)
				}
				if err.Error() != tt.expectedError {
					t.Errorf("Pesan error salah. Ekspektasi: '%s', Didapat: '%s'", tt.expectedError, err.Error())
				}
			} else {
				if err != nil {
					t.Fatalf("Ekspektasi sukses tanpa error, tetapi mendapat error: %v", err)
				}
				if res == nil {
					t.Fatal("Response login mengembalikan nil padahal status sukses")
				}
				if res.Token != tt.expectedToken {
					t.Errorf("Token tidak sesuai. Ekspektasi: '%s', Didapat: '%s'", tt.expectedToken, res.Token)
				}
			}
		})
	}
}
