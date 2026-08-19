package models

import "time"

type ArtikelPromote struct {
	ID        int        `json:"id" gorm:"column:id;primaryKey;autoIncrement"`
	ArtikelID int64      `json:"artikel_id" gorm:"column:artikel_id;not null"`
	Status    string     `json:"status" gorm:"column:status;type:varchar(255);not null;default:active"`
	CreatedAt *time.Time `json:"created_at" gorm:"column:created_at"`
	UpdatedAt *time.Time `json:"updated_at" gorm:"column:updated_at;autoUpdateTime"`
}

func (ArtikelPromote) TableName() string {
	return "artikel_promotes"
}

type ArtikelPromoteDatatable struct {
	No           int64      `json:"no"`
	ID           int        `json:"id"`
	NamaArtikel  string     `json:"nama_artikel"`
	NamaKategori string     `json:"nama_kategori"`
	Status       string     `json:"status"`
	CreatedAt    *time.Time `json:"created_at"`
	UpdatedAt    *time.Time `json:"updated_at"`
}
