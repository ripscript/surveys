package migrations

import (
	"backend/reportapi/models"

	"gorm.io/gorm"
)

func Migrate(db *gorm.DB) error {
	db.Exec(`ALTER TABLE log_block_ratings ADD CONSTRAINT uni_log_block_ratings_device_id UNIQUE (device_id)`)

	err := db.AutoMigrate(&models.LogActivity{}, &models.Report{}, &models.ReportCover{}, &models.ReportSection{}, &models.ReportSubSection{}, &models.ReportComponent{}, &models.ReportQueue{}, &models.LogBlockRating{})
	if err != nil {
		return err
	}

	if db.Migrator().HasColumn(&models.ReportSection{}, "description") {
		if err := db.Migrator().DropColumn(&models.ReportSection{}, "description"); err != nil {
			return err
		}
	}
	if db.Migrator().HasColumn(&models.ReportSubSection{}, "description") {
		if err := db.Migrator().DropColumn(&models.ReportSubSection{}, "description"); err != nil {
			return err
		}
	}
	return nil
}
