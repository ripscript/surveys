package models

import (
	"time"
)

type Survey struct {
	ID           int       `gorm:"primaryKey" json:"id"`
	Name         string    `gorm:"column:name;type:text" json:"name"`
	FlowDetailID int64     `gorm:"column:flow_detail_id" json:"flow_detail_id"`
	StartDate    time.Time `gorm:"column:start_date" json:"start_date"`
	EndDate      time.Time `gorm:"column:end_date" json:"end_date"`
	Type         string    `gorm:"column:type;type:varchar(255)" json:"type"`

	CreatedAt      time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt      time.Time `gorm:"column:updated_at" json:"updated_at"`
	Status         string    `gorm:"column:status;type:varchar(191)" json:"status"`
	CreatedBy      int64     `gorm:"column:created_by" json:"created_by"`
	Deskripsi      string    `gorm:"column:deskripsi;type:text" json:"deskripsi"`
	ApprovalSurvey string    `gorm:"column:approval_survey;type:varchar(255)" json:"approval_survey"`
	AlasanReject   *string   `gorm:"column:alasan_reject;type:text" json:"alasan_reject"`

	IsRepeated bool `gorm:"column:is_repeated;type:boolean;default:false" json:"is_repeated"`
}

func (u *Survey) TableName() string {
	return "surveys"
}
