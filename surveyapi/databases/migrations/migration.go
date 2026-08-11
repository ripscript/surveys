package migrations

import (
	"backend/surveyapi/models"

	"gorm.io/gorm"
)

func Migrate(db *gorm.DB) error {
	err := db.AutoMigrate(&models.FlowSection{})
	if err != nil {
		return err
	}
	return nil
}
