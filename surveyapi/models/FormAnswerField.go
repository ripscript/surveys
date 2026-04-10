package models

import (
	"time"
)

type FormAnswerField struct {
	ID          int       `gorm:"primaryKey" json:"id"`
	FormId      int       `gorm:"column:form_id" json:"form_id"`
	FormFieldId int       `gorm:"column:form_field_id" json:"form_field_id"`
	Option      string    `gorm:"column:option;type:text" json:"option,omitempty"`
	CreatedAt   time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at" json:"updated_at"`
	Sequence    string    `gorm:"column:sequence;type:varchar(191)" json:"sequence"`
}

func (u *FormAnswerField) TableName() string {
	return "form_answer_fields"
}
