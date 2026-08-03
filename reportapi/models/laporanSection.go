package models

import "time"

type LaporanSection struct {
	ID            int       `gorm:"primaryKey;autoIncrement" json:"id"`
	LaporanPageID int       `gorm:"not null" json:"laporan_page_id"`
	Title         string    `gorm:"type:varchar(255);not null" json:"title"`
	Sequence      int       `gorm:"not null" json:"sequence"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`

	// Relasi (Has Many)
	SubSections []LaporanSubSection `gorm:"foreignKey:LaporanSectionID" json:"sub_sections,omitempty"`
}

func (LaporanSection) TableName() string {
	return "laporan_sections"
}
