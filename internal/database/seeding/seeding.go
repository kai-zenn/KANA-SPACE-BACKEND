package seeding

import (
  "log"
  "gorm.io/gorm"
)

func SeedDatabase(db *gorm.DB) error {
  log.Println("Memulai proses seeding...")
  
	// Seed Admin
	if err := SeedAdminUser(db); err != nil {
		log.Fatalf("Gagal seeding admin: %v", err)
	}
  
	// Seed Kategori Lapak
	if err := SeedCategories(db); err != nil {
		log.Fatalf("Gagal seeding kategori: %v", err)
	}
  
	log.Println("Seeding selesai!")
	return nil
}
