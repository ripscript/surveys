package models

import (
	"time"

	"gorm.io/gorm"
)

type KelurahanModel struct {
	ID                uint           `gorm:"primaryKey" json:"id"`
	SubDistrictID     uint           `gorm:"type:bigint;not null" json:"sub_district_id"`
	VillageName       string         `gorm:"type:varchar(50);not null" json:"village_name"`
	VillageNameSlug   string         `gorm:"type:varchar(70);not null" json:"village_name_slug"`
	VillagePostalCode *string        `gorm:"type:varchar(191)" json:"village_postal_code"`
	KodeWilayah       *string        `gorm:"type:varchar(191)" json:"kode_wilayah"`
	Lat               *string        `gorm:"type:varchar(191)" json:"lat"`
	Long              *string        `gorm:"type:varchar(191)" json:"long"`
	CreatedAt         time.Time      `json:"created_at"`
	UpdatedAt         time.Time      `json:"updated_at"`
	DeletedAt         gorm.DeletedAt `gorm:"index" json:"deleted_at"`
	GeoName           *string        `gorm:"type:varchar(255)" json:"geo_name"`

	Kecamatan *Kecamatan `gorm:"foreignKey:SubDistrictID" json:"kecamatan,omitempty"`
}

func (u *KelurahanModel) TableName() string {
	return "kelurahans"
}
