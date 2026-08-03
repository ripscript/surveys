package models

import "time"

type LaporanPage struct {
	ID          int       `gorm:"primaryKey;autoIncrement" json:"id"`
	LaporanID   int       `gorm:"not null" json:"laporan_id"` // Relasi ke tabel legacy `laporans`
	Title       string    `gorm:"type:varchar(255);not null" json:"title"`
	Description *string   `gorm:"type:text" json:"description"` // Nullable jika tidak dipakai
	HasSection  bool      `gorm:"default:true;not null" json:"has_section"`
	Sequence    int       `gorm:"not null" json:"sequence"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	// Relasi (Has Many)
	Sections   []LaporanSection   `gorm:"foreignKey:LaporanPageID" json:"sections,omitempty"`
	Components []LaporanComponent `gorm:"foreignKey:LaporanPageID" json:"components,omitempty"` // Jika HasSection = false
}

func (LaporanPage) TableName() string {
	return "laporan_pages"
}
