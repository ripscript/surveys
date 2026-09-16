package seeders

import (
	"backend/masterapi/models"
	"backend/masterapi/utils"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"gorm.io/gorm"
)

func WilayahKelurahanGeoJSONSeed(db *gorm.DB) error {
	filePath := "assets/BatasWilayahKotaBandung/3273-kota-bandung-level-kelurahan_002.json"
	jsonFile, err := os.ReadFile(filePath)
	if err != nil {
		log.Fatalf("Gagal membaca file JSON: %v", err)
	}

	var geoData models.GeoJSON
	if err := json.Unmarshal(jsonFile, &geoData); err != nil {
		log.Fatalf("Gagal unmarshal JSON: %v", err)
	}

	for _, feature := range geoData.Features {
		nama := feature.Properties.NamaKelurahan
		kelurahan, err := getKelurahan(db, nama)
		if err != nil {
			log.Printf("Error mengecek kelurahan %s: %v", nama, err)
			continue
		}

		if kelurahan != nil {
			err := db.Model(&kelurahan).Update("geo_name", nama).Error
			if err != nil {
				log.Printf("Gagal update GeoJSON untuk %s: %v", nama, err)
			} else {
				fmt.Printf("[BERHASIL UPDATE] Kelurahan: %s\n", nama)
			}
		} else {
			fmt.Printf("[BELUM ADA] Kelurahan: %s. Perlu ditambahkan!\n", nama)
		}
	}

	return nil
}

func WilayahKelurahanGeoJSONSeedV2(db *gorm.DB) error {
	filePath := "assets/BatasWilayahKotaBandung/kelurahan.geojson"
	jsonFile, err := os.ReadFile(filePath)
	if err != nil {
		log.Fatalf("Gagal membaca file JSON: %v", err)
	}

	var geoData models.GeoJSONV2
	if err := json.Unmarshal(jsonFile, &geoData); err != nil {
		log.Fatalf("Gagal unmarshal JSON: %v", err)
	}

	kelurahanPrefixes := []string{"Kelurahan", "Kel."}

	// itter := 0
	for _, feature := range geoData.Features {
		nama := utils.TrimPrefixes(feature.Properties.Nama, kelurahanPrefixes...)

		// itter++
		kelurahan, err := getKelurahan(db, nama)
		if err != nil {
			log.Printf("Error mengecek kelurahan %s: %v", nama, err)
			continue
		}

		if kelurahan != nil {
			err := db.Model(&kelurahan).Update("geo_name", nama).Error
			if err != nil {
				log.Printf("Gagal update GeoJSON untuk %s: %v", nama, err)
			} else {
				fmt.Printf("[BERHASIL UPDATE] Kelurahan: %s\n", nama)
			}
		} else {
			fmt.Printf("[BELUM ADA] Kelurahan: %s. Perlu ditambahkan!\n", nama)
		}
	}

	// fmt.Println(itter)

	return nil
}

func getKelurahan(db *gorm.DB, kelurahanName string) (*models.KelurahanModel, error) {
	var kecamatan models.KelurahanModel

	normalizedJSONName := strings.ReplaceAll(strings.ToLower(kelurahanName), " ", "")

	err := db.Where("REPLACE(LOWER(village_name), ' ', '') = ?", normalizedJSONName).First(&kecamatan).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}

	return &kecamatan, nil
}
