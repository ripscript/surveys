package migrations

import (
	"backend/reportapi/models"

	"gorm.io/gorm"
)

func Migrate(db *gorm.DB) error {
	err := db.AutoMigrate(&models.LogActivity{})
	if err != nil {
		return err
	}
	return nil
}
