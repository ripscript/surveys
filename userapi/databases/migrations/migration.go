package migrations

import (
	"backend/userapi/models"

	"gorm.io/gorm"
)

func Migrate(db *gorm.DB) error {
	err := db.AutoMigrate(&models.Menu{}, &models.MenuPermission{})
	if err != nil {
		return err
	}
	return nil
}
