package models

import (
	"backend/surveyapi/enums"
	"time"

	"gorm.io/gorm"
)

type Form struct {
	ID          int              `gorm:"primaryKey" json:"id"`
	UserId      int64            `gorm:"column:user_id" json:"user_id"`
	Title       string           `gorm:"column:title;type:varchar(191)" json:"title"`
	Description string           `gorm:"column:description;type:text" json:"description"`
	Code        string           `gorm:"column:code;type:varchar(191)" json:"code"`
	Status      enums.StatusForm `gorm:"column:status;default:'open'" json:"status"`
	CreatedAt   time.Time        `gorm:"column:created_at" json:"created_at"`
	UpdatedAt   time.Time        `gorm:"column:updated_at" json:"updated_at"`
	DeletedAt   gorm.DeletedAt   `gorm:"column:deleted_at" json:"deleted_at"`
	IsVerified  bool             `gorm:"column:is_verified;default:false" json:"is_verified"`
	FlagTematik string           `gorm:"column:flag_tematik;type:varchar(255)" json:"flag_tematik"`
}

func (u *Form) TableName() string {
	return "forms"
}

type FormDetail struct {
	Form
	FormFields []FormFieldWithOption `json:"questions,omitempty"`
}

type FullForm struct {
	ID          int              `gorm:"primaryKey" json:"id"`
	UserId      int64            `gorm:"column:user_id" json:"user_id"`
	Title       string           `gorm:"column:title;type:varchar(191)" json:"title"`
	Description string           `gorm:"column:description;type:text" json:"description"`
	Code        string           `gorm:"column:code;type:varchar(191)" json:"code"`
	Status      enums.StatusForm `gorm:"column:status;default:'open'" json:"status"`
	CreatedAt   time.Time        `gorm:"column:created_at" json:"created_at"`
	UpdatedAt   time.Time        `gorm:"column:updated_at" json:"updated_at"`
	DeletedAt   gorm.DeletedAt   `gorm:"column:deleted_at" json:"deleted_at"`
	IsVerified  bool             `gorm:"column:is_verified;default:false" json:"is_verified"`
	FlagTematik string           `gorm:"column:flag_tematik;type:varchar(255)" json:"flag_tematik"`
	FormFields  []FullFormField  `gorm:"foreignKey:FormId;references:ID" json:"questions,omitempty"`
}

func (u *FullForm) TableName() string {
	return "forms"
}

type FormDatatableResponse struct {
	No            int64          `json:"no"`
	Id            int            `json:"id"`
	Code          string         `json:"code"`
	Title         string         `json:"title"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `json:"deleted_at"`
	FlagTematik   string         `json:"flag_tematik"`
	CreatedByName string         `json:"created_by_name"`
}
