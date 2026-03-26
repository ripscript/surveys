package models

import (
	"time"

	"gorm.io/gorm"
)

type ArtikelKategori struct {
	ID        int            `gorm:"primaryKey" json:"id"`
	Name      string         `gorm:"column:name" json:"name"`
	CreatedAt time.Time      `gorm:"column:created_at" json:"created_at"`
	UpdatedAt time.Time      `gorm:"column:updated_at" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at" json:"deleted_at"`
}

func (u *ArtikelKategori) TableName() string {
	return "artikel_categories"
}
