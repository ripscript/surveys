package models

import (
	"time"
)

type FlowDetail struct {
	ID            int       `gorm:"primaryKey" json:"id"`
	FormId        int       `gorm:"column:form_id" json:"form_id"`
	Name          string    `gorm:"column:name;type:varchar(191)" json:"name"`
	Description   *string   `gorm:"column:description;type:text" json:"description"`
	Code          string    `gorm:"column:code;type:varchar(191)" json:"code"`
	Version       int       `gorm:"column:version" json:"version"`
	StatusSection string    `gorm:"column:status_section;type:varchar(1)" json:"status_section"`
	CreatedAt     time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt     time.Time `gorm:"column:updated_at" json:"updated_at"`
	OpeningId     int       `gorm:"column:opening_id" json:"opening_id"`
	ClosingId     int       `gorm:"column:closing_id" json:"closing_id"`
	CreatedBy     int       `gorm:"column:created_by" json:"created_by"`
}

func (u *FlowDetail) TableName() string {
	return "flow_details"
}

type FlowDetailDatatableResponse struct {
	Id            int       `json:"id"`
	FlowName      string    `json:"flow_name"`
	FormName      string    `json:"form_name"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	CreatedBy     int       `json:"-"`
	CreatedByName string    `json:"created_by_name"`
	IsMyOwn       bool      `json:"is_my_own"`
	PosibleUpdate bool      `json:"posible_update"`
	PosibleDelete bool      `json:"posible_delete"`
}

type FlowPreview struct {
	FlowName   string             `json:"flow_name"`
	HasSection bool               `json:"has_section"`
	Sections   FlowPreviewSection `json:"sections"`
}

type FlowPreviewSection struct {
	SectionId              int     `json:"section_id"`
	SectionName            *string `json:"section_name"`
	TotalRequiredQuestions int     `json:"total_required_questions"`
	TotalOptionalQuestions int     `json:"total_optional_questions"`
}
