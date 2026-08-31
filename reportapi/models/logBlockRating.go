package models

import "time"

type LogBlockRating struct {
	ID         int        `json:"id" gorm:"column:id;primaryKey;autoIncrement"`
	RatingID   int        `json:"rating_id" gorm:"column:rating_id;type:integer"`
	IP         *string    `json:"ip" gorm:"column:ip;type:varchar(20)"`
	IsBlocked  string     `json:"is_blocked" gorm:"column:is_blocked;type:varchar(255);not null;default:'false'"`
	CreatedAt  *time.Time `json:"created_at" gorm:"column:created_at;autoCreateTime"`
	UpdatedAt  *time.Time `json:"updated_at" gorm:"column:updated_at;autoUpdateTime"`
	DeviceID   string     `json:"device_id" gorm:"column:device_id;type:varchar(191);not null"`
	DeviceInfo *string    `json:"device_info" gorm:"column:device_info;type:text"`
}

func (LogBlockRating) TableName() string {
	return "log_block_ratings"
}
