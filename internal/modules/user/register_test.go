package user

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
)

type mockUserRepository struct {
	CreateUserFunc           func(ctx context.Context, user *User) error
	GetProfileByUsernameFunc func(ctx context.Context, username string) (*User, error)
	GetProfileFunc           func(ctx context.Context, param UserParam) (*User, error)
	GetByIDFunc              func(ctx context.Context, userID uuid.UUID) (*User, error)
	UpdateUserFunc           func(ctx context.Context, userID uuid.UUID, updates map[string]interface{}) error
	UpdatePhotoFunc          func(ctx context.Context, userID uuid.UUID, photoLink string) error
}

func (m *mockUserRepository) CreateUser(ctx context.Context, user *User) error {
	return m.CreateUserFunc(ctx, user)
}
func (m *mockUserRepository) GetProfileByUsername(ctx context.Context, username string) (*User, error) {
	return m.GetProfileByUsernameFunc(ctx, username)
}
func (m *mockUserRepository) GetProfile(ctx context.Context, param UserParam) (*User, error) {
	return m.GetProfileFunc(ctx, param)
}
func (m *mockUserRepository) GetByID(ctx context.Context, userID uuid.UUID) (*User, error) {
	return m.GetByIDFunc(ctx, userID)
}
func (m *mockUserRepository) UpdateUser(ctx context.Context, userID uuid.UUID, updates map[string]interface{}) error {
	return m.UpdateUserFunc(ctx, userID, updates)
}
func (m *mockUserRepository) UpdatePhoto(ctx context.Context, userID uuid.UUID, photoLink string) error {
	return m.UpdatePhotoFunc(ctx, userID, photoLink)
}

type mockBcrypt struct {
	GenerateHashPasswordFunc func(password string) (string, error)
	CompareHashPasswordFunc  func(hashedPassword, password string) error
}

func (m *mockBcrypt) GenerateHashPassword(password string) (string, error) {
	return m.GenerateHashPasswordFunc(password)
}
func (m *mockBcrypt) CompareHashPassword(hashedPassword, password string) error {
	return m.CompareHashPasswordFunc(hashedPassword, password)
}

type mockJWT struct{}
func (m *mockJWT) GenerateToken(userID uuid.UUID, role string) (string, error) { return "", nil }

type mockStorage struct{}
func (m *mockStorage) UploadPhotoProfile(ctx context.Context, param PhotoUpdate) (string, error) { return "", nil }
func (m *mockStorage) DeletePhotoProfile(ctx context.Context, fileURL string) error { return nil }

type mockGoogleVerifier struct{}
func (m *mockGoogleVerifier) VerifyToken(ctx context.Context, idToken string) (*GoogleClaims, error) { return nil, nil }


func TestUserUseCase_Register(t *testing.T) {
	// Setup input standar untuk request registrasi
	validRequest := UserRegisterRequest{
		FirstName:   "Kana",
		LastName:    "Dev",
		Username:    "kanadev",
		Email:       "kana@gmail.com",
		PhoneNumber: "081234567890",
		Password:    "securepassword123",
		Role:        "user",
	}

	tests := []struct {
		name          string
		req           UserRegisterRequest
		setupMocks    func(repo *mockUserRepository, bc *mockBcrypt)
		expectedError string
	}{
		{
			name: "Sukses Registrasi",
			req:  validRequest,
			setupMocks: func(repo *mockUserRepository, bc *mockBcrypt) {
				repo.GetProfileByUsernameFunc = func(ctx context.Context, username string) (*User, error) {
					return nil, nil // Username aman (belum ada yang pakai)
				}
				repo.GetProfileFunc = func(ctx context.Context, param UserParam) (*User, error) {
					return nil, nil // Email aman (belum ada yang pakai)
				}
				bc.GenerateHashPasswordFunc = func(password string) (string, error) {
					return "hashed_password_example", nil
				}
				repo.CreateUserFunc = func(ctx context.Context, user *User) error {
					return nil // Berhasil disimpan ke database
				}
			},
			expectedError: "",
		},
		{
			name: "Gagal - Username Sudah Terdaftar",
			req:  validRequest,
			setupMocks: func(repo *mockUserRepository, bc *mockBcrypt) {
				repo.GetProfileByUsernameFunc = func(ctx context.Context, username string) (*User, error) {
					return &User{ID: uuid.New(), Username: "kanadev"}, nil // Username duplikat
				}
			},
			expectedError: "Username sudah terdaftar",
		},
		{
			name: "Gagal - Email Sudah Terdaftar",
			req:  validRequest,
			setupMocks: func(repo *mockUserRepository, bc *mockBcrypt) {
				repo.GetProfileByUsernameFunc = func(ctx context.Context, username string) (*User, error) {
					return nil, nil
				}
				repo.GetProfileFunc = func(ctx context.Context, param UserParam) (*User, error) {
					return &User{ID: uuid.New(), Email: "kana@gmail.com"}, nil // Email duplikat
				}
			},
			expectedError: "Email sudah terdaftar",
		},
		{
			name: "Gagal - Bcrypt Error",
			req:  validRequest,
			setupMocks: func(repo *mockUserRepository, bc *mockBcrypt) {
				repo.GetProfileByUsernameFunc = func(ctx context.Context, username string) (*User, error) {
					return nil, nil
				}
				repo.GetProfileFunc = func(ctx context.Context, param UserParam) (*User, error) {
					return nil, nil
				}
				bc.GenerateHashPasswordFunc = func(password string) (string, error) {
					return "", errors.New("bcrypt failure") // Kegagalan hashing internal
				}
			},
			expectedError: "bcrypt failure",
		},
		{
			name: "Gagal - Database Error Saat Create",
			req:  validRequest,
			setupMocks: func(repo *mockUserRepository, bc *mockBcrypt) {
				repo.GetProfileByUsernameFunc = func(ctx context.Context, username string) (*User, error) {
					return nil, nil
				}
				repo.GetProfileFunc = func(ctx context.Context, param UserParam) (*User, error) {
					return nil, nil
				}
				bc.GenerateHashPasswordFunc = func(password string) (string, error) {
					return "hashed_password_example", nil
				}
				repo.CreateUserFunc = func(ctx context.Context, user *User) error {
					return errors.New("database down") // Kegagalan koneksi DB
				}
			},
			expectedError: "database down",
		},
	}

	// Looping menjalankan semua skenario test
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Inisialisasi mock per skenario
			mockRepo := &mockUserRepository{}
			mockBc := &mockBcrypt{}
			mockJw := &mockJWT{}
			mockStore := &mockStorage{}
			mockGVerify := &mockGoogleVerifier{}

			// Konfigurasi kelakuan mock sesuai skenario test
			tt.setupMocks(mockRepo, mockBc)

			// Masukkan seluruh mock ke dalam constructor UseCase versi lu
			uc := NewUserUseCase(mockRepo, mockBc, mockJw, mockStore, mockGVerify)

			// Eksekusi fungsi utama
			err := uc.Register(context.Background(), tt.req)

			// Evaluasi Hasil Kontrak Error
			if tt.expectedError != "" {
				if err == nil {
					t.Fatalf("Ekspektasi error '%s', tetapi fungsi mengembalikan nil", tt.expectedError)
				}
				if err.Error() != tt.expectedError {
					t.Errorf("Pesan error salah. Ekspektasi: '%s', Didapat: '%s'", tt.expectedError, err.Error())
				}
			} else {
				if err != nil {
					t.Fatalf("Ekspektasi tidak ada error, tetapi mendapatkan error: %v", err)
				}
			}
		})
	}
}
