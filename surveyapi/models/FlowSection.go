package models

import (
	"time"
)

type FlowSection struct {
	ID        int       `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"column:name;type:varchar(50)" json:"name"`
	CreatedAt time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (u *FlowSection) TableName() string {
	return "flow__sections"
}
