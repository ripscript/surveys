package seeders

import (
	"backend/masterapi/models"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"gorm.io/gorm"
)

func WilayahKecamatanGeoJSONSeed(db *gorm.DB) error {
	filePath := "assets/BatasWilayahKotaBandung/3273-kota-bandung-level-kecamatan_002.json"
	jsonFile, err := os.ReadFile(filePath)
	if err != nil {
		log.Fatalf("Gagal membaca file JSON: %v", err)
	}

	var geoData models.GeoJSON
	if err := json.Unmarshal(jsonFile, &geoData); err != nil {
		log.Fatalf("Gagal unmarshal JSON: %v", err)
	}

	for _, feature := range geoData.Features {
		nama := feature.Properties.NamaKecamatan
		kecamatan, err := getKecamatan(db, nama)
		if err != nil {
			log.Printf("Error mengecek kecamatan %s: %v", nama, err)
			continue
		}

		if kecamatan != nil {
			err := db.Model(&kecamatan).Update("geo_name", nama).Error
			if err != nil {
				log.Printf("Gagal update GeoJSON untuk %s: %v", nama, err)
			} else {
				fmt.Printf("[BERHASIL UPDATE] Kecamatan: %s\n", nama)
			}
		} else {
			fmt.Printf("[BELUM ADA] Kecamatan: %s. Perlu ditambahkan!\n", nama)
		}
	}

	return nil
}

func getKecamatan(db *gorm.DB, kecamatanName string) (*models.KecamatanModel, error) {
	var kecamatan models.KecamatanModel

	normalizedJSONName := strings.ReplaceAll(strings.ToLower(kecamatanName), " ", "")

	err := db.Where("REPLACE(LOWER(sub_district_name), ' ', '') = ?", normalizedJSONName).First(&kecamatan).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}

	return &kecamatan, nil
}
