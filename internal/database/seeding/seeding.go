package seeding

import (
	"fmt"
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

	
	// Seed Users (seller + buyer)
	users, err := SeedUsers(db)
	if err != nil {
		return fmt.Errorf("gagal seeding users: %w", err)
	}

	// Seed Products (bergantung pada users & categories)
	if err := SeedProducts(db, users); err != nil {
		return fmt.Errorf("gagal seeding produk: %w", err)
	}

	// Seed Posts (bergantung pada users)
	if err := SeedPosts(db, users); err != nil {
		return fmt.Errorf("gagal seeding postingan: %w", err)
	}
  
	log.Println("Seeding selesai!")
	return nil
}
