package models

import "time"

type SurveyWilayah struct {
	ID             int       `gorm:"primaryKey" json:"id"`
	KecamatanId    int64     `gorm:"column:kecamatan_id" json:"kecamatan_id"`
	SurveyId       int64     `gorm:"column:survey_id" json:"survey_id"`
	TingkatWilayah string    `gorm:"column:tingkat_wilayah" json:"tingkat_wilayah"`
	KelurahanId    *int64    `gorm:"column:kelurahan_id" json:"kelurahan_id"`
	RWId           *int64    `gorm:"column:rw_id" json:"rw_id"`
	CreatedAt      time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt      time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (u *SurveyWilayah) TableName() string {
	return "survey_wilayahs"
}
