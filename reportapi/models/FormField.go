package models

import (
	"time"

	"gorm.io/gorm"
)

type FormFieldModel struct {
	ID            int            `gorm:"primaryKey" json:"id"`
	FormId        int            `gorm:"column:form_id" json:"form_id"`
	Template      string         `gorm:"column:template;type:varchar(191)" json:"template"`
	Attribute     string         `gorm:"column:attribute;type:varchar(191)" json:"attribute"`
	Deskripsi     *string        `gorm:"column:deskripsi;type:text" json:"deskripsi"`
	Question      string         `gorm:"column:question;type:varchar(191)" json:"question"`
	Required      bool           `gorm:"column:required" json:"required"`
	Options       *string        `gorm:"column:options;type:text" json:"options,omitempty"`
	Filled        bool           `gorm:"column:filled" json:"filled"`
	DeletedAt     gorm.DeletedAt `gorm:"column:deleted_at" json:"deleted_at"`
	CreatedAt     time.Time      `gorm:"column:created_at" json:"created_at"`
	UpdatedAt     time.Time      `gorm:"column:updated_at" json:"updated_at"`
	IsVerified    bool           `gorm:"column:is_verified;default:false" json:"is_verified"`
	Sequence      int            `gorm:"column:sequence" json:"sequence"`
	ImageQuantity *string        `gorm:"column:image_quantity;type:varchar(10)" json:"image_quantity,omitempty"`
}

func (u *FormField) TableName() string {
	return "form_fields"
}

type FormFieldWithOption struct {
	FormField
	T_Options []FormAnswerField `json:"options"`
}

type UpdateFormFieldWithOption struct {
	FormField
	Options       []FormAnswerField `json:"options"`
	OptionsToKeep []int             `json:"options_to_keep"`
}

type FullFormField struct {
	ID            int               `gorm:"primaryKey" json:"id"`
	FormId        int               `gorm:"column:form_id" json:"form_id"`
	Template      string            `gorm:"column:template;type:varchar(191)" json:"template"`
	Attribute     string            `gorm:"column:attribute;type:varchar(191)" json:"attribute"`
	Deskripsi     *string           `gorm:"column:deskripsi;type:text" json:"deskripsi"`
	Question      string            `gorm:"column:question;type:varchar(191)" json:"question"`
	Required      bool              `gorm:"column:required" json:"required"`
	Filled        bool              `gorm:"column:filled" json:"filled"`
	DeletedAt     gorm.DeletedAt    `gorm:"column:deleted_at" json:"deleted_at"`
	CreatedAt     time.Time         `gorm:"column:created_at" json:"created_at"`
	UpdatedAt     time.Time         `gorm:"column:updated_at" json:"updated_at"`
	IsVerified    bool              `gorm:"column:is_verified;default:false" json:"is_verified"`
	Sequence      int               `gorm:"column:sequence" json:"sequence"`
	ImageQuantity *string           `gorm:"column:image_quantity;type:varchar(10)" json:"image_quantity,omitempty"`
	Options       []FormAnswerField `gorm:"foreignKey:FormFieldId;references:ID" json:"options"`
}

func (u *FullFormField) TableName() string {
	return "form_fields"
}
