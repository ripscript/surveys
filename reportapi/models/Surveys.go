package models

import "time"

type Log_Surveys struct {
	RespondentID int       `json:"respondent_id"`
	SurveyID     int       `json:"survey_id"`
	Keterangan   string    `json:"keterangan"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	IsRead       string    `json:"is_read"`
	LogLevel     string    `json:"log_level"`
}

type Respondents struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

func (u *Log_Surveys) TableName() string {
	return "log__surveys"
}
