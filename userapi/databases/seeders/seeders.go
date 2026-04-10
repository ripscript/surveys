package seeders

import (
	"gorm.io/gorm"
)

func Seed(db *gorm.DB) error {

	// seeder users
	err := MenuSeed(db)
	if err != nil {
		return err
	}

	return nil
}
