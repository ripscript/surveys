package models

import (
	"encoding/json"
	"time"
)

type ReportCover struct {
	ID            uint            `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	ReportID      int64           `gorm:"column:report_id;not null;uniqueIndex:idx_report_covers_report_id;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"report_id"`
	ImgDepan      *string         `gorm:"column:img_depan;type:text" json:"img_depan,omitempty"`
	ImgBelakang   *string         `gorm:"column:img_belakang;type:text" json:"img_belakang,omitempty"`
	Page          json.RawMessage `gorm:"column:page;type:jsonb" json:"page,omitempty"`
	TextDepan     *string         `gorm:"column:text_depan;type:text" json:"text_depan,omitempty"`
	TextBelakang  *string         `gorm:"column:text_belakang;type:text" json:"text_belakang,omitempty"`
	KataPengantar *string         `gorm:"column:kata_pengantar;type:text" json:"kata_pengantar,omitempty"`
	CreatedAt     *time.Time      `gorm:"column:created_at" json:"created_at,omitempty"`
	UpdatedAt     *time.Time      `gorm:"column:updated_at" json:"updated_at,omitempty"`

	// Relasi balik
	Report *Report `gorm:"foreignKey:ReportID;references:ID" json:"report,omitempty"`
}

func (ReportCover) TableName() string {
	return "report_covers" // Atau "laporan_covers"
}
