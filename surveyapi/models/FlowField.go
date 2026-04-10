package models

import (
	"time"
)

type FlowField struct {
	ID                int       `gorm:"primaryKey" json:"id"`
	FormFieldId       int       `gorm:"column:form_field_id" json:"form_field_id"`
	FormAnswerFieldId *int      `gorm:"column:form_answer_field_id" json:"form_answer_field_id"`
	ChildId           int       `gorm:"column:child_id" json:"child_id"`
	SectionId         *int      `gorm:"column:section_id" json:"section_id"`
	CreatedAt         time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt         time.Time `gorm:"column:updated_at" json:"updated_at"`
	FlowDetailId      int       `gorm:"column:flow_detail_id" json:"flow_detail_id"`
	Sequence          int       `gorm:"column:sequence" json:"sequence"`
	Breakdown         bool      `gorm:"column:breakdown" json:"breakdown"`
	IsAdvancedOption  bool      `gorm:"column:is_advanced_option" json:"is_advanced_option"`
	GroupChildId      int       `gorm:"column:group_child_id" json:"group_child_id"`
	GroupId           int       `gorm:"column:group_id" json:"group_id"`
}

func (u *FlowField) TableName() string {
	return "flow_fields"
}
