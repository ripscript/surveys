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
	return nil
}
