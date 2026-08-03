package migrations

import (
	"backend/reportapi/models"

	"gorm.io/gorm"
)

func Migrate(db *gorm.DB) error {
	err := db.AutoMigrate(&models.LogActivity{}, &models.Laporan{}, &models.LaporanCover{}, &models.LaporanPage{}, &models.LaporanSection{}, &models.LaporanComponent{})
	if err != nil {
		return err
	}
	return nil
}
