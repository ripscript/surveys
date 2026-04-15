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
	AlasanReject   string    `gorm:"column:alasan_reject;type:text" json:"alasan_reject"`
}

func (u *Survey) TableName() string {
	return "surveys"
}

type SurveyDatatableResponse struct {
	ID              int       `json:"id" gorm:"column:id"`
	SurveyCode      string    `json:"survey_code" gorm:"column:survey_code"`
	SurveyName      string    `json:"survey_name" gorm:"column:survey_name"`
	SurveyDimulai   time.Time `json:"start_date" gorm:"column:start_date"`
	SurveyBerakhir  time.Time `json:"end_date" gorm:"column:end_date"`
	AlurName        string    `json:"flow_name" gorm:"column:flow_name"`
	CreatedAt       time.Time `json:"created_at" gorm:"column:created_at"`
	UpdatedAt       time.Time `json:"updated_at" gorm:"column:updated_at"`
	CreatedBy       int       `json:"-" gorm:"column:created_by"`
	CreatedByName   string    `json:"created_by_name" gorm:"column:created_by_name"`
	IsMyOwn         bool      `json:"is_my_own"`
	TotalResponden  int       `json:"total_responden" gorm:"column:total_responden"`
	Status          string    `json:"status" gorm:"column:status"`
	Approval        string    `json:"approval_survey" gorm:"column:approval_survey"`
	PosibleUpdate   bool      `json:"posible_update" gorm:"column:posible_update"`
	PosibleDelete   bool      `json:"posible_delete" gorm:"column:posible_delete"`
	PosibleApproval bool      `json:"posible_approval" gorm:"column:posible_approval"`
}
