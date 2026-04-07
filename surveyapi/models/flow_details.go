package models

import (
	"time"
)

type FlowDetail struct {
	ID            int       `gorm:"primaryKey" json:"id"`
	FormId        int       `gorm:"column:form_id" json:"form_id"`
	Name          string    `gorm:"column:name;type:varchar(191)" json:"name"`
	Description   string    `gorm:"column:description;type:text" json:"description"`
	Code          string    `gorm:"column:code;type:varchar(191)" json:"code"`
	Version       int       `gorm:"column:version" json:"version"`
	StatusSection int       `gorm:"column:status_section" json:"status_section"`
	CreatedAt     time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt     time.Time `gorm:"column:updated_at" json:"updated_at"`
	OpeningId     int       `gorm:"column:opening_id" json:"opening_id"`
	ClosingId     int       `gorm:"column:closing_id" json:"closing_id"`
	CreatedBy     int       `gorm:"column:created_by" json:"created_by"`
}

func (u *FlowDetail) TableName() string {
	return "flow_details"
}
