package models

import (
	"time"

	"gorm.io/gorm"
)

type Kecamatan struct {
	ID              int64          `json:"id" gorm:"primaryKey;autoIncrement;not null"`
	SubDistrictName string         `json:"sub_district_name" gorm:"type:varchar(191);not null"`
	SubDistrictSlug string         `json:"sub_district_slug" gorm:"type:varchar(191);not null"`
	KodeWilayah     *string        `json:"kode_wilayah" gorm:"type:varchar(191);"`
	Lat             *string        `json:"lat" gorm:"type:varchar(191);"`
	Long            *string        `json:"long" gorm:"type:varchar(191);"`
	CreatedAt       time.Time      `json:"created_at" gorm:"type:timestamp;default:now()"`
	UpdatedAt       time.Time      `json:"updated_at" gorm:"type:timestamp;default:now()"`
	DeletedAt       gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}

func (u *Kecamatan) TableName() string {
	return "kecamatans"
}

type KecamatanDatatableResponse struct {
	ID              int64          `json:"id"`
	SubDistrictName string         `json:"sub_district_name"`
	SubDistrictSlug string         `json:"sub_district_slug"`
	KodeWilayah     *string        `json:"kode_wilayah"`
	Lat             *string        `json:"lat"`
	Long            *string        `json:"long"`
	NamaPejabat     *string        `json:"nama_pejabat"`
	PeriodeAwal     *time.Time     `json:"periode_awal"`
	PeriodeAkhir    *time.Time     `json:"periode_akhir"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `json:"deleted_at"`
}

type KecamatanOptionItem struct {
	ID    *int64  `form:"id"`
	Label *string `form:"label"`
}
