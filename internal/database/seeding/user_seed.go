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

type dummyUserData struct {
	Username  string
	Email     string
	FirstName string
	LastName  string
	Phone     string
	Role      string
}

var sellerSeedData = []dummyUserData{
	{"budi_material", "budi.material@example.com", "Budi", "Santoso", "081234567801", "seller"},
	{"siti_daurulang", "siti.daurulang@example.com", "Siti", "Rahayu", "081234567802", "seller"},
	{"agus_recycle", "agus.recycle@example.com", "Agus", "Wijaya", "081234567803", "seller"},
	{"dewi_kreasi", "dewi.kreasi@example.com", "Dewi", "Lestari", "081234567804", "seller"},
	{"hendra_lapak", "hendra.lapak@example.com", "Hendra", "Gunawan", "081234567805", "seller"},
}

var buyerSeedData = []dummyUserData{
	{"rina_beli", "rina.beli@example.com", "Rina", "Puspita", "081234567901", "user"},
	{"joko_cari", "joko.cari@example.com", "Joko", "Prasetyo", "081234567902", "user"},
	{"maya_hijau", "maya.hijau@example.com", "Maya", "Anggraini", "081234567903", "user"},
	{"fajar_daur", "fajar.daur@example.com", "Fajar", "Nugroho", "081234567904", "user"},
	{"lina_material", "lina.material@example.com", "Lina", "Kusuma", "081234567905", "user"},
	{"eko_borongan", "eko.borongan@example.com", "Eko", "Setiawan", "081234567906", "user"},
	{"nita_zerowaste", "nita.zerowaste@example.com", "Nita", "Handayani", "081234567907", "user"},
	{"rudi_bahan", "rudi.bahan@example.com", "Rudi", "Firmansyah", "081234567908", "user"},
	{"putri_hijau", "putri.hijau@example.com", "Putri", "Wulandari", "081234567909", "user"},
	{"dian_material", "dian.material@example.com", "Dian", "Permana", "081234567910", "user"},
}

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

func SeedUsers(db *gorm.DB) ([]user.User, error) {
	var existingCount int64
	if err := db.Model(&user.User{}).
		Where("role IN ?", []string{"seller", "user"}).
		Count(&existingCount).Error; err != nil {
		return nil, fmt.Errorf("gagal mengecek user existing: %w", err)
	}

	if existingCount > 0 {
		log.Println("User seller/buyer sudah ada, skip seeding user...")
		var existingUsers []user.User
		if err := db.Where("role IN ?", []string{"seller", "user"}).
			Find(&existingUsers).Error; err != nil {
			return nil, fmt.Errorf("gagal mengambil user existing: %w", err)
		}
		return existingUsers, nil
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("gagal melakukan hash password: %w", err)
	}
	passwordStr := string(hashedPassword)

	allData := make([]dummyUserData, 0, len(sellerSeedData)+len(buyerSeedData))
	allData = append(allData, sellerSeedData...)
	allData = append(allData, buyerSeedData...)

	users := make([]user.User, 0, len(allData))
	for _, d := range allData {
		users = append(users, user.User{
			ID:        uuid.New(),
			Username:  d.Username,
			Email:     d.Email,
			Password:  &passwordStr,
			Role:      d.Role,
			FirstName: d.FirstName,
			LastName:  d.LastName,
			PhoneNumber: &d.Phone,
		})
	}

	if err := db.Create(&users).Error; err != nil {
		return nil, fmt.Errorf("gagal insert users: %w", err)
	}

	log.Printf("Berhasil seeding %d user (%d seller, %d buyer)\n",
		len(users), len(sellerSeedData), len(buyerSeedData))

	return users, nil
}
