package migrations

import (
	"backend/userapi/models"

	"gorm.io/gorm"
)

func Migrate(db *gorm.DB) error {
	err := db.AutoMigrate(&models.UserModel{}, &models.Menu{}, &models.MenuPermission{}, &models.Survey{})
	if err != nil {
		return err
	}
	return nil
}
