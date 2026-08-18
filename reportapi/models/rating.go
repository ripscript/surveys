package models

import "time"

type Rating struct {
	ID        int        `json:"id" gorm:"column:id;primaryKey;autoIncrement"`
	Rating    int        `json:"rating" gorm:"column:rating;type:integer;not null"`
	Ulasan    string     `json:"ulasan" gorm:"column:ulasan;type:text;not null"`
	CreatedAt *time.Time `json:"created_at" gorm:"column:created_at;autoCreateTime"`
	UpdatedAt *time.Time `json:"updated_at" gorm:"column:updated_at;autoUpdateTime"`
}

// Menyesuaikan nama tabel secara eksplisit
func (Rating) TableName() string {
	return "ratings"
}

type RatingDatatable struct {
	No        int64      `json:"no" gorm:"-"`
	ID        int        `json:"id" gorm:"column:id;primaryKey;autoIncrement"`
	Rating    int        `json:"rating" gorm:"column:rating;type:integer;not null"`
	Ulasan    string     `json:"ulasan" gorm:"column:ulasan;type:text;not null"`
	CreatedAt *time.Time `json:"created_at" gorm:"column:created_at;autoCreateTime"`
	UpdatedAt *time.Time `json:"updated_at" gorm:"column:updated_at;autoUpdateTime"`
}
