package seeders

import (
	"gorm.io/gorm"
)

func Seed(db *gorm.DB) error {

	// seeder users
	err := CMSSectionSeed(db)
	if err != nil {
		return err
	}

	return nil
}
