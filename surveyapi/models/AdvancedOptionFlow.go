package models

import (
	"time"
)

type AdvancedOptionFlow struct {
	ID          int       `gorm:"primaryKey" json:"id"`
	FlowFieldId int       `gorm:"column:flow_field_id" json:"flow_field_id"`
	FormFieldId int       `gorm:"column:form_field_id" json:"form_field_id"`
	Option      int       `gorm:"column:option" json:"option"`
	ChildId     int       `gorm:"column:child_id" json:"child_id"`
	CreatedAt   time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (u *AdvancedOptionFlow) TableName() string {
	return "advanced_option_flows"
}
