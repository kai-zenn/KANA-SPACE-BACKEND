package seeding

import "gorm.io/gorm"

func SeedDatabase(db *gorm.DB) error {
	err := SeedAdminUser(db)
	if err != nil {
		return err
	}
	return nil
}
