package models

import "time"

type SurveySurveyor struct {
	ID           int       `gorm:"primaryKey" json:"id"`
	RespondentId int64     `gorm:"column:respondent_id" json:"respondent_id"`
	SurveyId     int       `gorm:"column:survey_id" json:"survey_id"`
	CreatedAt    time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt    time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (u *SurveySurveyor) TableName() string {
	return "survey__surveyors"
}
