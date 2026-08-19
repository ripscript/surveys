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
	Contents          *datatypes.JSON `gorm:"column:contents;type:jsonb" json:"contents"`
	Thumbnail         *string         `gorm:"column:thumbnail" json:"thumbnail"`
	CreatedBy         *int64          `gorm:"column:created_by" json:"created_by"`
	CreatedAt         *time.Time      `gorm:"column:created_at" json:"created_at"`
	UpdatedAt         *time.Time      `gorm:"column:updated_at" json:"updated_at"`
}

func (Artikel) TableName() string {
	return "artikels"
}

type ArtikelContentItem struct {
	Description *string `json:"description,omitempty"`
	Image       *string `json:"image,omitempty"`
	File        *string `json:"file,omitempty"`
	Placeholder *string `json:"placeholder,omitempty"`
}

type ArtikelContentResponse struct {
	Description *string `json:"description,omitempty"`
	Image       *string `json:"image,omitempty"`
	File        *string `json:"file,omitempty"`
	Placeholder *string `json:"placeholder,omitempty"`
}

type ArtikelDetailResponse struct {
	ID           int64                    `json:"id"`
	Title        string                   `json:"title"`
	CategoryID   int64                    `json:"category_id"`
	CategoryName string                   `json:"category_name"`
	Thumbnail    string                   `json:"thumbnail"`
	Contents     []ArtikelContentResponse `json:"contents"`
}

type ArtikelDatatable struct {
	No            int64      `json:"no"`
	ID            int64      `json:"id"`
	Judul         *string    `json:"judul"`
	NamaKategori  string     `json:"nama_kategori"`
	CreatedByName *string    `json:"created_by_name"`
	CreatedAt     *time.Time `json:"created_at"`
	UpdatedAt     *time.Time `json:"updated_at"`
}

type ArtikelListPublic struct {
	ID        int64      `json:"id"`
	Judul     *string    `json:"judul"`
	Thumbnail *string    `json:"thumbnail"`
	CreatedAt *time.Time `json:"created_at"`
}
