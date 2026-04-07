package models

import (
	"backend/surveyapi/enums"
	"time"

	"gorm.io/gorm"
)

type GeneralTemplate struct {
	ID         int                      `gorm:"primaryKey" json:"id"`
	Type       enums.TypeTemplateUcapan `gorm:"column:type;type:varchar(255)" json:"type"`
	Name       string                   `gorm:"column:name;type:varchar(191)" json:"name"`
	Content    string                   `gorm:"column:content;type:text" json:"content"`
	IsVerified bool                     `gorm:"column:is_verified;default:false" json:"is_verified"`
	CreatedAt  time.Time                `gorm:"column:created_at" json:"created_at"`
	UpdatedAt  time.Time                `gorm:"column:updated_at" json:"updated_at"`
	DeletedAt  gorm.DeletedAt           `gorm:"column:deleted_at" json:"deleted_at"`
	CreatedBy  int                      `gorm:"column:created_by" json:"created_by"`
}

func (u *GeneralTemplate) TableName() string {
	return "general_templates"
}

type GeneralTemplateDatatableResponse struct {
	ID            int                      `json:"id"`
	Type          enums.TypeTemplateUcapan `json:"type"`
	Name          string                   `json:"name"`
	CreatedAt     time.Time                `json:"created_at"`
	UpdatedAt     time.Time                `json:"updated_at"`
	DeletedAt     gorm.DeletedAt           `json:"deleted_at"`
	CreatedByName string                   `json:"created_by_name"`
}
