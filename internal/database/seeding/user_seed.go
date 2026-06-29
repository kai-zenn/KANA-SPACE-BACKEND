package seeding

import (
	"KANA-SPACE-BACKEND/internal/modules/user"
	"fmt"
	"log"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/google/uuid"
)

func SeedAdminUser(db *gorm.DB) error {
  adminTemplates := []struct {
		Username  string
		Email     string
		FirstName string
		LastName  string
	}{
		{Username: "admin_ayb", Email: "admin1@kana.com", FirstName: "Super", LastName: "Admin satu"},
		{Username: "admin_fhr", Email: "admin2@kana.com", FirstName: "Super", LastName: "Admin dua"},
		{Username: "admin_rnl", Email: "admin3@kana.com", FirstName: "Super", LastName: "Admin Tiga"},
		{Username: "admin_gb", Email: "admin4@kana.com", FirstName: "Super", LastName: "Admin Empat"},
	}
  
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	passwordStr := string(hashedPassword)
  
	for _, t := range adminTemplates {
		var count int64
		db.Model(&user.User{}).Where("email = ?", t.Email).Count(&count)
		
		if count > 0 {
			log.Printf("Admin %s (%s) sudah terdaftar, skip...", t.Username, t.Email)
			continue
		}
  
		newAdmin := user.User{
			ID:               uuid.New(),
			FirstName:        t.FirstName,
			LastName:         t.LastName,
			Username:         t.Username,
			Email:            t.Email,
			Password:         &passwordStr,
			Role:             "admin",
			ProfilePhotoLink: "",
			CreatedAt:        time.Now(),
			UpdatedAt:        time.Now(),
		}
  
		if err := db.Create(&newAdmin).Error; err != nil {
			return fmt.Errorf("gagal membuat admin %s: %w", t.Username, err)
		}
		log.Printf("Berhasil menambahkan admin: %s", t.Username)
	}
  
	return nil
}
