package models

import (
	"time"

	"gorm.io/gorm"
)

type DataRt struct {
	ID          int64          `json:"id" gorm:"primaryKey;autoIncrement;not null"`
	RwId        int64          `json:"rw_id" gorm:"type:bigint;not null"`
	NamaRt      string         `json:"nama_rt" gorm:"type:varchar(4);not null"`
	Lat         *string        `json:"lat" gorm:"type:varchar(191);"`
	Long        *string        `json:"long" gorm:"type:varchar(191);"`
	CreatedAt   time.Time      `json:"created_at" gorm:"type:timestamp;autoCreateTime"`
	UpdatedAt   time.Time      `json:"updated_at" gorm:"type:timestamp;autoUpdateTime"`
	DeletedAt   gorm.DeletedAt `json:"deleted_at" gorm:"index"`
	KodeWilayah *string        `json:"kode_wilayah" gorm:"type:varchar(191);"`
}

func (u *DataRt) TableName() string {
	return "data__rts"
}

type DataRtDetail struct {
	ID              int64          `json:"id" gorm:"primaryKey;autoIncrement;not null"`
	SubDistrictId   int64          `json:"sub_district_id" gorm:"->;type:bigint;not null"`
	SubDistrictName string         `json:"sub_district_name" gorm:"->;type:varchar(191);not null"`
	VillageId       int64          `json:"village_id" gorm:"->;type:bigint;not null"`
	VillageName     string         `json:"village_name" gorm:"->;type:varchar(191);not null"`
	NamaRw          string         `json:"nama_rw" gorm:"->;type:varchar(191);not null"`
	RwId            int64          `json:"rw_id" gorm:"type:bigint;not null"`
	NamaRt          string         `json:"nama_rt" gorm:"type:varchar(4);not null"`
	Lat             *string        `json:"lat" gorm:"type:varchar(191);"`
	Long            *string        `json:"long" gorm:"type:varchar(191);"`
	CreatedAt       time.Time      `json:"created_at" gorm:"type:timestamp;default:now()"`
	UpdatedAt       time.Time      `json:"updated_at" gorm:"type:timestamp;default:now()"`
	DeletedAt       gorm.DeletedAt `json:"deleted_at" gorm:"index"`
	KodeWilayah     *string        `json:"kode_wilayah" gorm:"type:varchar(191);"`
}

func (u *DataRtDetail) TableName() string {
	return "data__rts"
}

type RtDatatableResponse struct {
	ID              int64      `json:"id"`
	KodeWilayah     *string    `json:"kode_wilayah"`
	SubDistrictName string     `json:"sub_district_name"`
	SubDistrictId   int64      `json:"sub_district_id"`
	VillageName     string     `json:"village_name"`
	VillageId       int64      `json:"village_id"`
	NamaRw          string     `json:"nama_rw"`
	RwId            int64      `json:"rw_id"`
	NamaRt          string     `json:"nama_rt"`
	NamaPejabat     *string    `json:"nama_pejabat"`
	PeriodeAwal     *time.Time `json:"periode_awal"`
	PeriodeAkhir    *time.Time `json:"periode_akhir"`
	Lat             *string    `json:"lat"`
	Long            *string    `json:"long"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}
