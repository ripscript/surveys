package seeders

import (
	"fmt"

	"gorm.io/gorm"
)

func Seed(db *gorm.DB) error {

	fmt.Println("BEGIN SEEDING CMS SECTION ============================")
	err := CMSSectionSeed(db)
	if err != nil {
		return err
	}
	fmt.Println("END SEEDING CMS SECTION ============================")

	fmt.Println("BEGIN SEEDING METRICS MAPPING ============================")
	if err := DashboardMetricSeed(db); err != nil {
		fmt.Println("Error seeding DashboardMetric:", err)
	}
	fmt.Println("END SEEDING METRICS MAPPING ============================")

	fmt.Println("BEGIN SEEDING KECAMATAN GEOJSON ============================")
	if err := WilayahKecamatanGeoJSONSeedV2(db); err != nil {
		return err
	}
	fmt.Println("END SEEDING KECAMATAN GEOJSON ============================")

	fmt.Println("BEGIN SEEDING KELURAHAN ============================")
	if err := WilayahKelurahanSeed(db); err != nil {
		return err
	}
	fmt.Println("END SEEDING KELURAHAN ============================")

	fmt.Println("BEGIN SEEDING KELURAHAN GEOJSON ============================")
	if err := WilayahKelurahanGeoJSONSeedV2(db); err != nil {
		return err
	}
	fmt.Println("END SEEDING KELURAHAN GEOJSON ============================")

	return nil
}
