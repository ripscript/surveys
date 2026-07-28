package migrations

import (
	"backend/masterapi/models"

	"gorm.io/gorm"
)

func Migrate(db *gorm.DB) error {
	err := db.AutoMigrate(
		&models.CMSSection{},
		&models.CMSContent{},
		&models.CMSItem{},
		&models.CMSMedia{},
	)
	if err != nil {
		return err
	}
	return nil
}
