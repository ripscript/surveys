package models

import (
	"time"

	"gorm.io/datatypes"
)

type Report struct {
	ID             int64          `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Name           string         `gorm:"column:name;type:varchar(191);not null" json:"name"`
	RespondentID   int64          `gorm:"column:respondent_id;not null;index:idx_reports_respondent_id" json:"respondent_id"`
	TingkatWilayah string         `gorm:"column:tingkat_wilayah;type:varchar(255);not null" json:"tingkat_wilayah"`
	KecamatanID    datatypes.JSON `gorm:"column:kecamatan_id;type:jsonb" json:"kecamatan_id"`
	KelurahanID    datatypes.JSON `gorm:"column:kelurahan_id;type:jsonb" json:"kelurahan_id"`
	SurveyID       datatypes.JSON `gorm:"column:survey_id;type:jsonb;not null" json:"survey_id"`
	CreatedAt      *time.Time     `gorm:"column:created_at" json:"created_at,omitempty"`
	UpdatedAt      *time.Time     `gorm:"column:updated_at" json:"updated_at,omitempty"`

	Cover    *ReportCover    `gorm:"foreignKey:ReportID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"cover,omitempty"`
	Sections []ReportSection `gorm:"foreignKey:ReportID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"sections,omitempty"`
}

func (Report) TableName() string {
	return "reports"
}
