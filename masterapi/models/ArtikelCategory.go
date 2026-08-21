package models

import (
	"time"

	"gorm.io/gorm"
)

type ArtikelCategory struct {
	ID        int            `json:"id" gorm:"column:id;primaryKey;autoIncrement"`
	Name      string         `json:"name" gorm:"column:name;type:varchar(191);not null"`
	CreatedAt *time.Time     `json:"created_at" gorm:"column:created_at;autoCreateTime"`
	UpdatedAt *time.Time     `json:"updated_at" gorm:"column:updated_at;autoUpdateTime"`
	DeletedAt gorm.DeletedAt `json:"deleted_at" gorm:"column:deleted_at;index"`
}

func (ArtikelCategory) TableName() string {
	return "artikel_categories"
}

type ArtikelCategoryDatatable struct {
	No           int64          `json:"no"`
	ID           int            `json:"id"`
	TotalArtikel int64          `json:"total_artikel"`
	Name         string         `json:"name"`
	CreatedAt    *time.Time     `json:"created_at"`
	UpdatedAt    *time.Time     `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `json:"deleted_at"`
}

type ArtikelCategoryPublic struct {
	ID   int    `json:"id" gorm:"column:id;primaryKey;autoIncrement"`
	Name string `json:"name" gorm:"column:name;type:varchar(191);not null"`
}
