package models

import (
	"time"

	"gorm.io/gorm"
)

type Kelurahan struct {
	ID                int64          `json:"id" gorm:"primaryKey;autoIncrement;not null"`
	SubDistrictId     int64          `json:"sub_district_id" gorm:"type:bigint;not null"`
	VillageName       string         `json:"village_name" gorm:"type:varchar(50);not null"`
	VillageNameSlug   string         `json:"village_name_slug" gorm:"type:varchar(70);not null"`
	VillagePostalCode *string        `json:"village_postal_code" gorm:"type:varchar(191);"`
	KodeWilayah       *string        `json:"kode_wilayah" gorm:"type:varchar(191);"`
	Lat               *string        `json:"lat" gorm:"type:varchar(191);"`
	Long              *string        `json:"long" gorm:"type:varchar(191);"`
	CreatedAt         time.Time      `json:"created_at" gorm:"type:timestamp;default:now()"`
	UpdatedAt         time.Time      `json:"updated_at" gorm:"type:timestamp;default:now()"`
	DeletedAt         gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}

func (u *Kelurahan) TableName() string {
	return "kelurahans"
}

type KelurahanDetail struct {
	ID                int64          `json:"id" gorm:"primaryKey;autoIncrement;not null"`
	SubDistrictName   string         `json:"sub_district_name" gorm:"->;type:varchar(191);not null"`
	SubDistrictId     int64          `json:"sub_district_id" gorm:"type:bigint;not null"`
	VillageName       string         `json:"village_name" gorm:"type:varchar(50);not null"`
	VillageNameSlug   string         `json:"village_name_slug" gorm:"type:varchar(70);not null"`
	VillagePostalCode *string        `json:"village_postal_code" gorm:"type:varchar(191);"`
	KodeWilayah       *string        `json:"kode_wilayah" gorm:"type:varchar(191);"`
	Lat               *string        `json:"lat" gorm:"type:varchar(191);"`
	Long              *string        `json:"long" gorm:"type:varchar(191);"`
	CreatedAt         time.Time      `json:"created_at" gorm:"type:timestamp;default:now()"`
	UpdatedAt         time.Time      `json:"updated_at" gorm:"type:timestamp;default:now()"`
	DeletedAt         gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}

func (u *KelurahanDetail) TableName() string {
	return "kelurahans"
}

type KelurahanDatatableResponse struct {
	ID              int64          `json:"id"`
	SubDistrictName string         `json:"sub_district_name"`
	SubDistrictId   int64          `json:"sub_district_id"`
	KodeWilayah     *string        `json:"kode_wilayah"`
	VillageName     string         `json:"village_name"`
	NamaPejabat     *string        `json:"nama_pejabat"`
	PeriodeAwal     *time.Time     `json:"periode_awal"`
	PeriodeAkhir    *time.Time     `json:"periode_akhir"`
	Lat             *string        `json:"lat"`
	Long            *string        `json:"long"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}
