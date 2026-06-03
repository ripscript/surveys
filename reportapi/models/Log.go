package models

import "time"

type LogActivity struct {
	ID          uint64    `gorm:"primaryKey" json:"id"`
	No          int       `json:"no"`
	ServiceName string    `gorm:"size:100;not null" json:"service_name"`
	Module      string    `gorm:"size:100;not null" json:"module"`
	Action      string    `gorm:"size:50;not null" json:"action"`
	UserID      uint64    `json:"user_id"`
	UserName    string    `gorm:"size:255" json:"user_name"`
	EntityID    string    `gorm:"size:255" json:"entity_id"`
	Description string    `gorm:"type:text" json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}

type LogActivitySave struct {
	ServiceName string    `gorm:"size:100;not null" json:"serviceName"`
	Module      string    `gorm:"size:100;not null" json:"module"`
	Action      string    `gorm:"size:50;not null" json:"action"`
	UserID      int       `json:"userId"`
	UserName    string    `gorm:"size:255" json:"userName"`
	EntityID    string    `gorm:"size:255" json:"entityID"`
	Description string    `gorm:"type:text" json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}

func (LogActivitySave) TableName() string {
	return "log_activities"
}
