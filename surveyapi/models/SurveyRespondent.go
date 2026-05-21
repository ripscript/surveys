package models

import (
	"time"
)

type SurveyRespondent struct {
	ID               int64      `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	RespondentID     int64      `gorm:"column:respondent_id;not null" json:"respondent_id"`
	SurveyID         int64      `gorm:"column:survey_id;not null" json:"survey_id"`
	CreatedAt        *time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt        *time.Time `gorm:"column:updated_at" json:"updated_at"`
	Status           *int       `gorm:"column:status;type:smallint;default:0" json:"status"`
	StatusApproval   *string    `gorm:"column:status_approval;type:varchar(20)" json:"status_approval"`
	IsEditPertanyaan string     `gorm:"column:is_edit_pertanyaan;type:varchar(255);not null;default:'false'" json:"is_edit_pertanyaan"`
}

func (SurveyRespondent) TableName() string {
	return "survey_respondents"
}
