package models

import (
	"time"
)

type PejabatWilayah struct {
	ID           int64      `json:"id" gorm:"primaryKey"`
	TipeWilayah  *int64     `json:"tipe_wilayah" gorm:"column:tipe_wilayah"`
	IdWilayah    *int64     `json:"id_wilayah" gorm:"column:id_wilayah"`
	PeriodeAwal  *time.Time `json:"periode_awal" gorm:"column:periode_awal"`
	PeriodeAkhir *time.Time `json:"periode_akhir" gorm:"column:periode_akhir"`
	StatusJabat  *int       `json:"status_jabat" gorm:"column:status_jabat"` // 1: Aktif, 0: Tidak Aktif
	IdResponden  *int64     `json:"id_responden" gorm:"column:id_responden"`
	CreatedAt    time.Time  `gorm:"column:created_at" json:"created_at"`
	UpdatedAt    time.Time  `gorm:"column:updated_at" json:"updated_at"`
	NoSK         *string    `json:"no_sk" gorm:"column:no_sk"`
}

func (u *PejabatWilayah) TableName() string {
	return "pejabat__wilayahs"
}

type DetailPejabatWilayah struct {
	ID           int64      `json:"id" gorm:"primaryKey"`
	TipeWilayah  *int64     `json:"tipe_wilayah" gorm:"column:tipe_wilayah"`
	IdWilayah    *int64     `json:"id_wilayah" gorm:"column:id_wilayah"`
	PeriodeAwal  *time.Time `json:"periode_awal" gorm:"column:periode_awal"`
	PeriodeAkhir *time.Time `json:"periode_akhir" gorm:"column:periode_akhir"`
	StatusJabat  *int       `json:"status_jabat" gorm:"column:status_jabat"` // 1: Aktif, 0: Tidak Aktif
	IdResponden  *int64     `json:"id_responden" gorm:"column:id_responden"`
	CreatedAt    time.Time  `gorm:"column:created_at" json:"created_at"`
	UpdatedAt    time.Time  `gorm:"column:updated_at" json:"updated_at"`
	NoSK         *string    `json:"no_sk" gorm:"column:no_sk"`
	Respondent   Respondent `json:"respondent" gorm:"foreignKey:IdResponden;references:ID"`
}

func (u *DetailPejabatWilayah) TableName() string {
	return "pejabat__wilayahs"
}

type PejabatWilayahList struct {
	No            int64     `json:"no"`
	ID            int       `json:"id"`
	DaftarWilayah string    `json:"daftar_wilayah"`
	TipeWilayah   string    `json:"tipe_wilayah"`
	NamaPejabat   string    `json:"nama_pejabat"`
	PeriodeAwal   time.Time `json:"periode_awal"`
	PeriodeAkhir  time.Time `json:"periode_akhir"`
	StatusJabat   bool      `json:"status_jabat"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}
