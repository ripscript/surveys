package migrations

import (
	"backend/reportapi/models"

	"gorm.io/gorm"
)

func Migrate(db *gorm.DB) error {
	err := db.AutoMigrate(&models.LogActivity{}, &models.Report{}, &models.ReportCover{}, &models.ReportSection{}, &models.ReportSubSection{}, &models.ReportComponent{})
	if err != nil {
		return err
	}

	db.Migrator().DropColumn(&models.ReportSection{}, "description")
	db.Migrator().DropColumn(&models.ReportSubSection{}, "description")
	return nil
}
