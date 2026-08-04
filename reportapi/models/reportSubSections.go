package models

import "time"

type ReportSubSection struct {
	ID              int64      `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	ReportSectionID int64      `gorm:"column:report_section_id;not null;index:idx_report_subsections_section_id;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"report_section_id"`
	Title           string     `gorm:"column:title;type:varchar(255);not null" json:"title"`
	Description     *string    `gorm:"column:description;type:text" json:"description,omitempty"`
	Sequence        int        `gorm:"column:sequence;not null;index:idx_report_subsections_seq" json:"sequence"`
	CreatedAt       *time.Time `gorm:"column:created_at" json:"created_at,omitempty"`
	UpdatedAt       *time.Time `gorm:"column:updated_at" json:"updated_at,omitempty"`

	// Relasi
	Components []ReportComponent `gorm:"foreignKey:ReportSubSectionID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"components,omitempty"`
}

func (ReportSubSection) TableName() string {
	return "report_subsections"
}
