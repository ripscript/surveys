package models

import (
	"time"
)

type BLKS struct {
	ID        int64     `json:"id" gorm:"primaryKey;autoIncrement;not null"`
	Name      string    `json:"name" gorm:"type:text;not null"`
	CreatedAt time.Time `json:"created_at" gorm:"type:timestamp;default:now()"`
	UpdatedAt time.Time `json:"updated_at" gorm:"type:timestamp;default:now()"`
}

func (u *BLKS) TableName() string {
	return "blks"
}
