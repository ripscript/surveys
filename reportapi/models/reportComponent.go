package models

import (
	"time"

	"gorm.io/datatypes"
)

type ReportComponent struct {
	ID                 int64   `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	ReportSectionID    *int64  `gorm:"column:report_section_id;index:idx_report_components_section_id;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"report_section_id,omitempty"`             // Terisi JIKA Section TIDAK punya Sub-Section
	ReportSubSectionID *int64  `gorm:"column:report_sub_section_id;index:idx_report_components_sub_section_id;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"report_sub_section_id,omitempty"` // Terisi JIKA Section PUNYA Sub-Section
	Type               string  `gorm:"column:type;type:varchar(50);not null" json:"type"`                                                                                                            // isinya: 'chart', 'table', 'narrative'
	Sequence           int     `gorm:"column:sequence;not null;index:idx_report_components_seq" json:"sequence"`
	Title              *string `gorm:"column:title;type:text" json:"title,omitempty"`

	// --- UMUM (Dipakai oleh Chart & Tabel) ---
	FormFieldIDs   datatypes.JSON `gorm:"column:form_field_ids;type:jsonb" json:"form_field_ids,omitempty"` // Array of int (ID Soal), cth: [12, 13]
	IsMultipleData bool           `gorm:"column:is_multiple_data;default:false;not null" json:"is_multiple_data"`

	// --- SPESIFIK: GRAFIK ---
	ChartType      *string `gorm:"column:chart_type;type:varchar(50)" json:"chart_type,omitempty"`           // 'bar', 'pie', 'line'
	ChartDirection *string `gorm:"column:chart_direction;type:varchar(20)" json:"chart_direction,omitempty"` // 'horizontal', 'vertical'

	// --- SPESIFIK: TABEL ---
	TableStyle  *string        `gorm:"column:table_style;type:varchar(50)" json:"table_style,omitempty"` // 'simple', 'grouped_header'
	TableConfig datatypes.JSON `gorm:"column:table_config;type:jsonb" json:"table_config,omitempty"`     // Simpan show_territory_col, custom label, dan mapping group

	// --- SPESIFIK: NARASI ---
	NarrativeTemplate *string        `gorm:"column:narrative_template;type:text" json:"narrative_template,omitempty"`
	NarrativeLogic    datatypes.JSON `gorm:"column:narrative_logic;type:jsonb" json:"narrative_logic,omitempty"`

	CreatedAt *time.Time `gorm:"column:created_at" json:"created_at,omitempty"`
	UpdatedAt *time.Time `gorm:"column:updated_at" json:"updated_at,omitempty"`
}

func (ReportComponent) TableName() string {
	return "report_components"
}
