package models

import (
	"encoding/json"
	"time"
)

type LaporanKonten struct {
	ID            uint            `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	LaporanID     int64           `gorm:"column:laporan_id;not null" json:"laporan_id"`
	ImgDepan      *string         `gorm:"column:img_depan;type:text" json:"img_depan,omitempty"`
	ImgBelakang   *string         `gorm:"column:img_belakang;type:text" json:"img_belakang,omitempty"`
	Page          json.RawMessage `gorm:"column:page;type:json" json:"page,omitempty"`
	CreatedAt     *time.Time      `gorm:"column:created_at" json:"created_at,omitempty"`
	UpdatedAt     *time.Time      `gorm:"column:updated_at" json:"updated_at,omitempty"`
	TextDepan     *string         `gorm:"column:text_depan;type:text" json:"text_depan,omitempty"`
	TextBelakang  *string         `gorm:"column:text_belakang;type:text" json:"text_belakang,omitempty"`
	KataPengantar *string         `gorm:"column:kata_pengantar;type:text" json:"kata_pengantar,omitempty"`

	// Relation
	Laporan *Laporan `gorm:"foreignKey:LaporanID" json:"laporan,omitempty"`
}

func (LaporanKonten) TableName() string {
	return "laporan_kontens"
}
