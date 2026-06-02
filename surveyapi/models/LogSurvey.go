package models

import "time"

type LogSurvey struct {
	ID           int        `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	RespondentID int64      `gorm:"column:respondent_id" json:"respondent_id"`
	SurveyID     int64      `gorm:"column:survey_id" json:"survey_id"`
	Keterangan   *string    `gorm:"column:keterangan" json:"keterangan"`
	CreatedAt    *time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt    *time.Time `gorm:"column:updated_at" json:"updated_at"`
	IsRead       *string    `gorm:"column:is_read;size:255" json:"is_read"`
	LogLevel     *string    `gorm:"column:log_level;size:191" json:"log_level"`
	NotifFor     *string    `gorm:"column:notif_for;size:191" json:"notif_for"`
}

func (LogSurvey) TableName() string {
	return "log__surveys"
}

type LogSurveyDetail struct {
	Keterangan *string    `gorm:"column:keterangan" json:"keterangan"`
	CreatedAt  *time.Time `gorm:"column:created_at" json:"created_at"`
}

func (LogSurveyDetail) TableName() string {
	return "log__surveys"
}
