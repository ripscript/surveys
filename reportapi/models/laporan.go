package models

import (
	"time"

	"gorm.io/datatypes"
)

type Laporan struct {
	ID             uint           `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Name           string         `gorm:"column:name;type:varchar(191);not null;uniqueIndex:idx_laporan_name" json:"name"`
	RespondentID   int64          `gorm:"column:respondent_id;not null" json:"respondent_id"`
	TingkatWilayah string         `gorm:"column:tingkat_wilayah;type:varchar(255);not null" json:"tingkat_wilayah"`
	KecamatanID    datatypes.JSON `gorm:"column:kecamatan_id;type:json" json:"kecamatan_id,omitempty"`
	KelurahanID    datatypes.JSON `gorm:"column:kelurahan_id;type:json" json:"kelurahan_id,omitempty"`
	SurveyID       datatypes.JSON `gorm:"column:survey_id;type:json;not null" json:"survey_id"`
	CreatedAt      *time.Time     `gorm:"column:created_at" json:"created_at,omitempty"`
	UpdatedAt      *time.Time     `gorm:"column:updated_at" json:"updated_at,omitempty"`

	LaporanCover *LaporanCover `gorm:"foreignKey:LaporanID" json:"laporan_cover,omitempty"`
}

func (Laporan) TableName() string {
	return "laporans"
}

type LaporanDatatable struct {
	No        int64      `gorm:"-" json:"no"`
	ID        uint       `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Name      string     `gorm:"column:name;type:varchar(191);not null" json:"name"`
	CreatedAt *time.Time `gorm:"column:created_at" json:"created_at,omitempty"`
	UpdatedAt *time.Time `gorm:"column:updated_at" json:"updated_at,omitempty"`
}
