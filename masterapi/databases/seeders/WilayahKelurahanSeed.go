package seeders

import (
	"backend/masterapi/models"
	"backend/masterapi/utils"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"gorm.io/gorm"
)

type WilayahDTO struct {
	KodeKecamatan string `json:"kode_kecamatan"`
	Kecamatan     string `json:"kecamatan"`
	KodeKelurahan string `json:"kode_kelurahan"`
	Kelurahan     string `json:"kelurahan"`
}

type WilayahListDTO []WilayahDTO

func WilayahKelurahanSeed(db *gorm.DB) error {
	filePath := "assets/kelurahan.json"
	jsonFile, err := os.ReadFile(filePath)
	if err != nil {
		log.Fatalf("Gagal membaca file JSON: %v", err)
	}

	var datas WilayahListDTO
	if err := json.Unmarshal(jsonFile, &datas); err != nil {
		log.Fatalf("Gagal unmarshal JSON: %v", err)
	}

	for _, data := range datas {
		var searchName string
		switch data.Kelurahan {
		case "Cijaura":
			searchName = "Cijawura"
		case "Kebon Jayanti":
			searchName = "Kebun Jayanti"
		case "Kebon Kangkung":
			searchName = "Kebun Kangkung"
		case "Husein Sastranegara":
			searchName = "Husen Sastranegara"
		default:
			searchName = data.Kelurahan
		}
		kelurahan, err := getKelurahan(db, searchName)
		if err != nil {
			log.Printf("Error mengecek kelurahan %s: %v", data.Kelurahan, err)
			continue
		}

		if kelurahan != nil {
			fmt.Printf("[SUDAH ADA] Kelurahan: %s\n", data.Kelurahan)
			slug := utils.StringToSlug(data.Kelurahan, "-")
			err := db.Model(&kelurahan).Updates(map[string]interface{}{
				"kode_wilayah":      data.KodeKelurahan,
				"village_name":      data.Kelurahan,
				"village_name_slug": slug,
			}).Error

			if err != nil {
				log.Printf("Gagal kelurahan untuk %s: %v", data.Kelurahan, err)
			} else {
				fmt.Printf("[BERHASIL UPDATE] Kelurahan: %s\n", data.Kelurahan)
			}
		} else {
			if data.Kelurahan != "Cijaura" && data.Kelurahan != "Kebon Jayanti" && data.Kelurahan != "Kebon Kangkung" {
				fmt.Printf("[BELUM ADA] Kelurahan: %s. Perlu ditambahkan!\n", data.Kelurahan)
				kecamatan, err := getKecamatan(db, data.Kecamatan)
				if err != nil {
					log.Printf("Error mengecek kecamatan %s: %v", data.Kecamatan, err)
					continue
				}

				slug := utils.StringToSlug(data.Kelurahan, "-")
				err = db.Create(&models.Kelurahan{
					SubDistrictId:   int64(kecamatan.ID),
					VillageName:     data.Kelurahan,
					VillageNameSlug: slug,
					KodeWilayah:     &data.KodeKelurahan,
				}).Error
				if err != nil {
					fmt.Printf("[GAGAL TAMBAH] Kelurahan: %s. Error: %v\n", data.Kelurahan, err)
				} else {
					fmt.Printf("[BERHASIL TAMBAH] Kelurahan: %s\n", data.Kelurahan)
				}
			}
		}
	}

	return nil
}
