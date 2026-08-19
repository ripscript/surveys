package models

import (
	"time"

	"gorm.io/datatypes"
)

type Artikel struct {
	ID                int64           `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	ArtikelCategoryID *int64          `gorm:"column:artikel_category_id" json:"artikel_category_id"`
	Judul             *string         `gorm:"column:judul;size:191" json:"judul"`
	Isi               *string         `gorm:"column:isi" json:"isi"`
	Content           *datatypes.JSON `gorm:"column:content;type:jsonb" json:"content"`
	Thumbnail         *string         `gorm:"column:thumbnail" json:"thumbnail"`
	CreatedBy         *int64          `gorm:"column:created_by" json:"created_by"`
	CreatedAt         *time.Time      `gorm:"column:created_at" json:"created_at"`
	UpdatedAt         *time.Time      `gorm:"column:updated_at" json:"updated_at"`
}

func (Artikel) TableName() string {
	return "artikels"
}
