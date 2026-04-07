package models

import (
	"time"
)

type SurveyWilayah struct {
	ID             int       `gorm:"primaryKey" json:"id"`
	KecamatanId    int       `gorm:"column:kecamatan_id" json:"kecamatan_id"`
	SurveyId       int       `gorm:"column:survey_id" json:"survey_id"`
	CreatedAt      time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt      time.Time `gorm:"column:updated_at" json:"updated_at"`
	TingkatWilayah *string   `gorm:"column:tingkat_wilayah;type:char(1)" json:"tingkat_wilayah"`
	KelurahanId    *int      `gorm:"column:kelurahan_id" json:"kelurahan_id"`
	RWId           *int      `gorm:"column:rw_id" json:"rw_id"`
}

func (u *SurveyWilayah) TableName() string {
	return "survey_wilayahs"
}
