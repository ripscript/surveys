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
	CreatedAt       time.Time      `gorm:"column:created_at" json:"created_at"`
	UpdatedAt       time.Time      `gorm:"column:updated_at" json:"updated_at"`
	DeletedAt       gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}

func (u *Kecamatan) TableName() string {
	return "kecamatans"
}

type KecamatanDatatableResponse struct {
	No              int64          `json:"no"`
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

type ResultKecamatan struct {
	ID              int64
	SubDistrictName string
	TotalKelurahan  int64
	TotalRw         int64
	TotalRt         int64
}
