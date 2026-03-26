package models

import (
	"time"
)

type PejabatWilayah struct {
	ID            int64       `json:"id" gorm:"primaryKey;autoIncrement;not null"`
	TipeWilayah   *int64      `json:"tipe_wilayah" gorm:"type:int8;"` // 5: Kecamatan, 4: Kelurahan, 3: RW, 2: RT
	IdWilayah     *int64      `json:"id_wilayah" gorm:"type:int8;"`
	PeriodeAwal   *time.Time  `json:"periode_awal" gorm:"type:date;"`
	PeriodeAkhir  *time.Time  `json:"periode_akhir" gorm:"type:date;"`
	StatusJabatan *bool       `json:"status_jabatan" gorm:"type:boolean;default:true;"`
	IdResponden   *int64      `json:"id_responden" gorm:"type:int8;"`
	NoSK          *string     `json:"no_sk" gorm:"type:varchar(255);"`
	CreatedAt     time.Time   `json:"created_at" gorm:"type:timestamp;default:now()"`
	UpdatedAt     time.Time   `json:"updated_at" gorm:"type:timestamp;default:now()"`
	Respondent    *Respondent `json:"respondent" gorm:"foreignKey:IdResponden;references:ID"`
	// Relasi ke tabel Respondent
	// pejabat_wilayahs.id_responden <-> respondent.id
}

func (u *PejabatWilayah) TableName() string {
	return "pejabat__wilayahs"
}
