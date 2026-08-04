package models

import "time"

type ReportSection struct {
	ID            int64      `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	ReportID      int64      `gorm:"column:report_id;not null;index:idx_report_sections_report_id;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"report_id"`
	Title         string     `gorm:"column:title;type:varchar(255);not null" json:"title"`
	Description   *string    `gorm:"column:description;type:text" json:"description,omitempty"`
	HasSubSection bool       `gorm:"column:has_sub_section;default:true;not null" json:"has_sub_section"` // true = pakai sub_section, false = langsung komponen
	Sequence      int        `gorm:"column:sequence;not null;index:idx_report_sections_seq" json:"sequence"`
	CreatedAt     *time.Time `gorm:"column:created_at" json:"created_at,omitempty"`
	UpdatedAt     *time.Time `gorm:"column:updated_at" json:"updated_at,omitempty"`

	// Relasi
	SubSections []ReportSubSection `gorm:"foreignKey:ReportSectionID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"sub_sections,omitempty"`
	Components  []ReportComponent  `gorm:"foreignKey:ReportSectionID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"components,omitempty"` // Dipakai JIKA HasSubSection = false
}

func (ReportSection) TableName() string {
	return "report_sections"
}
