package seeders

import (
	"backend/userapi/models"

	"gorm.io/gorm"
)

func MenuSeed(db *gorm.DB) error {
	Menus := []models.Menu{
		{},
	}
	return db.Create(&Menus).Error
}
