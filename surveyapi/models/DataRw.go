package models

import (
	"time"

	"gorm.io/gorm"
)

type DataRw struct {
	ID          int64          `json:"id" gorm:"primaryKey;autoIncrement;not null"`
	KelurahanId int64          `json:"kelurahan_id" gorm:"type:bigint;not null"`
	NamaRw      string         `json:"nama_rw" gorm:"type:varchar(191);not null"`
	KodeWilayah *string        `json:"kode_wilayah" gorm:"type:varchar(191);"`
	Lat         *string        `json:"lat" gorm:"type:varchar(191);"`
	Long        *string        `json:"long" gorm:"type:varchar(191);"`
	CreatedAt   time.Time      `json:"created_at" gorm:"type:timestamp;autoCreateTime"`
	UpdatedAt   time.Time      `json:"updated_at" gorm:"type:timestamp;autoUpdateTime"`
	DeletedAt   gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}

func (u *DataRw) TableName() string {
	return "data__rws"
}

type DataRwDetail struct {
	ID              int64          `json:"id" gorm:"primaryKey;autoIncrement;not null"`
	SubDistrictId   int64          `json:"sub_district_id" gorm:"->;type:bigint;not null"`
	SubDistrictName string         `json:"sub_district_name" gorm:"->;type:varchar(191);not null"`
	VillageId       int64          `json:"village_id" gorm:"->;type:bigint;not null"`
	VillageName     string         `json:"village_name" gorm:"->;type:varchar(191);not null"`
	NamaRw          string         `json:"nama_rw" gorm:"type:varchar(191);not null"`
	KodeWilayah     *string        `json:"kode_wilayah" gorm:"type:varchar(191);"`
	Lat             *string        `json:"lat" gorm:"type:varchar(191);"`
	Long            *string        `json:"long" gorm:"type:varchar(191);"`
	CreatedAt       time.Time      `json:"created_at" gorm:"type:timestamp;default:now()"`
	UpdatedAt       time.Time      `json:"updated_at" gorm:"type:timestamp;default:now()"`
	DeletedAt       gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}

func (u *DataRwDetail) TableName() string {
	return "data__rws"
}

type RwDatatableResponse struct {
	No              int64          `json:"no"`
	ID              int64          `json:"id"`
	KodeWilayah     *string        `json:"kode_wilayah"`
	SubDistrictName string         `json:"sub_district_name"`
	SubDistrictId   int64          `json:"sub_district_id"`
	VillageName     string         `json:"village_name"`
	VillageId       int64          `json:"village_id"`
	NamaRw          string         `json:"nama_rw"`
	NamaPejabat     *string        `json:"nama_pejabat"`
	PeriodeAwal     *time.Time     `json:"periode_awal"`
	PeriodeAkhir    *time.Time     `json:"periode_akhir"`
	Lat             *string        `json:"lat"`
	Long            *string        `json:"long"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}
