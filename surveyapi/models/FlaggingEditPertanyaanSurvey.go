package models

import "time"

type FlaggingEditPertanyaanSurvey struct {
	ID                 int64      `gorm:"primaryKey;autoIncrement" json:"id"`
	SurveyRespondentID int64      `gorm:"column:survey_respondent_id" json:"survey_respondent_id"`
	FormFieldID        int64      `gorm:"column:form_field_id" json:"form_field_id"`
	IsRevisied         string     `gorm:"column:is_revisied;default:'false'" json:"is_revisied"` // Tipe Varchar
	CreatedAt          *time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt          *time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (FlaggingEditPertanyaanSurvey) TableName() string {
	return "flagging_edit_pertanyaan_surveys"
}
