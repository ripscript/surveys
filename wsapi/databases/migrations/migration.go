package migrations

import (
	"backend/wsapi/models"

	"gorm.io/gorm"
)

func Migrate(db *gorm.DB) error {
	err := db.AutoMigrate(
		&models.WsTicket{},
	)
	if err != nil {
		return err
	}
	return nil
}
