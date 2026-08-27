package seeders

import (
	"gorm.io/gorm"
)

func Seed(db *gorm.DB) error {

	err := CMSSectionSeed(db)
	if err != nil {
		return err
	}

	if err := DashboardMetricSeed(db); err != nil {
		return err
	}

	if err := WilayahKecamatanGeoJSONSeed(db); err != nil {
		return err
	}

	if err := WilayahKelurahanGeoJSONSeed(db); err != nil {
		return err
	}

	return nil
}
