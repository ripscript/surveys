package models

import (
	"time"

	"gorm.io/gorm"
)

type KecamatanModel struct {
	ID              uint           `gorm:"primaryKey" json:"id"`
	SubDistrictName string         `gorm:"type:varchar(191);not null" json:"sub_district_name"`
	SubDistrictSlug string         `gorm:"type:varchar(191);not null" json:"sub_district_slug"`
	KodeWilayah     *string        `gorm:"type:varchar(191)" json:"kode_wilayah"`
	Lat             *string        `gorm:"type:varchar(191)" json:"lat"`
	Long            *string        `gorm:"type:varchar(191)" json:"long"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"deleted_at"`
	GeoName         *string        `gorm:"type:varchar(255)" json:"geo_name"`

	Kelurahans []KelurahanModel `gorm:"foreignKey:SubDistrictID" json:"kelurahans,omitempty"`
}

func (u *KecamatanModel) TableName() string {
	return "kecamatans"
}
