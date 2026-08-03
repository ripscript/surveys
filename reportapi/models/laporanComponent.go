package models

import (
	"time"

	"gorm.io/datatypes"
)

type LaporanComponent struct {
	ID                  int     `gorm:"primaryKey;autoIncrement" json:"id"`
	LaporanPageID       *int    `json:"laporan_page_id"`                       // Nullable FK
	LaporanSubSectionID *int    `json:"laporan_sub_section_id"`                // Nullable FK
	Type                string  `gorm:"type:varchar(50);not null" json:"type"` // 'chart', 'table', 'chart_and_table', 'narrative'
	Sequence            int     `gorm:"not null" json:"sequence"`
	Title               *string `gorm:"type:varchar(255)" json:"title"`

	// --- KONFIGURASI GRAFIK & DATA ---
	FormFieldIDs   datatypes.JSON `gorm:"type:json" json:"form_field_ids"`         // Array ID pertanyaan, misal: [12, 13]
	ChartType      *string        `gorm:"type:varchar(50)" json:"chart_type"`      // 'bar', 'pie', 'line'
	ChartDirection *string        `gorm:"type:varchar(20)" json:"chart_direction"` // 'horizontal', 'vertical'
	IsMultipleData bool           `gorm:"default:false;not null" json:"is_multiple_data"`

	// --- KONFIGURASI NARASI ---
	NarrativeTemplate *string        `gorm:"type:text" json:"narrative_template"`
	NarrativeLogic    datatypes.JSON `gorm:"type:json" json:"narrative_logic"` // Config JSON untuk nilai Max/Min

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (LaporanComponent) TableName() string {
	return "laporan_components"
}
