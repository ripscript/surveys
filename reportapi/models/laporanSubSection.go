package models

import "time"

type LaporanSubSection struct {
	ID               int       `gorm:"primaryKey;autoIncrement" json:"id"`
	LaporanSectionID int       `gorm:"not null" json:"laporan_section_id"`
	Title            string    `gorm:"type:varchar(255);not null" json:"title"`
	Description      *string   `gorm:"type:text" json:"description"`
	Sequence         int       `gorm:"not null" json:"sequence"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`

	// Relasi (Has Many)
	Components []LaporanComponent `gorm:"foreignKey:LaporanSubSectionID" json:"components,omitempty"`
}

func (LaporanSubSection) TableName() string {
	return "laporan_sub_sections"
}
