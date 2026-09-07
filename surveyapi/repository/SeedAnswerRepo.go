package repository

import (
	"errors"

	"gorm.io/gorm"
)

type FormFieldRow struct {
	ID            int64  `gorm:"column:id"`
	Template      string `gorm:"column:template"`
	ImageQuantity string `gorm:"column:image_quantity"`
}

type SeedAnswerRepo interface {
	GetFormIDBySurveyID(surveyID int64) (int64, error)
	GetFormFieldsByFormID(formID int64) ([]FormFieldRow, error)
	GetAnswerOptionIDs(formFieldID int64) ([]int64, error)
	GetOrCreateSurveyRespondent(respondentID, surveyID int64) (int64, error)
	InsertFieldResponse(formFieldID, formResponseID int64, answer string) error
}

type seedAnswerRepo struct {
	dbMaster *gorm.DB
	dbSlave  *gorm.DB
}

func NewSeedAnswerRepo(dbMaster, dbSlave *gorm.DB) SeedAnswerRepo {
	return &seedAnswerRepo{dbMaster: dbMaster, dbSlave: dbSlave}
}

func (r *seedAnswerRepo) GetFormIDBySurveyID(surveyID int64) (int64, error) {
	var formID int64
	err := r.dbSlave.Raw(`
		SELECT fd.form_id
		FROM surveys s
		JOIN flow_details fd ON fd.id = s.flow_detail_id
		WHERE s.id = ?
	`, surveyID).Scan(&formID).Error
	return formID, err
}

func (r *seedAnswerRepo) GetFormFieldsByFormID(formID int64) ([]FormFieldRow, error) {
	var rows []FormFieldRow
	err := r.dbSlave.Raw(`
		SELECT id, template, image_quantity
		FROM form_fields
		WHERE form_id = ? AND deleted_at IS NULL
	`, formID).Scan(&rows).Error
	return rows, err
}

func (r *seedAnswerRepo) GetAnswerOptionIDs(formFieldID int64) ([]int64, error) {
	var ids []int64
	err := r.dbSlave.Raw(`
		SELECT id FROM form_answer_fields WHERE form_field_id = ?
	`, formFieldID).Scan(&ids).Error
	return ids, err
}

func (r *seedAnswerRepo) GetOrCreateSurveyRespondent(respondentID, surveyID int64) (int64, error) {
	var id int64
	r.dbSlave.Raw(`
		SELECT id FROM survey_respondents WHERE respondent_id = ? AND survey_id = ?
	`, respondentID, surveyID).Scan(&id)
	if id > 0 {
		return id, nil
	}

	if err := r.dbMaster.Exec(`
		INSERT INTO survey_respondents (respondent_id, survey_id, status, status_approval, is_edit_pertanyaan, created_at, updated_at)
		VALUES (?, ?, 1, 'validated_lurah', 'false', now(), now())
	`, respondentID, surveyID).Error; err != nil {
		return 0, err
	}

	r.dbSlave.Raw(`
		SELECT id FROM survey_respondents WHERE respondent_id = ? AND survey_id = ? ORDER BY id DESC LIMIT 1
	`, respondentID, surveyID).Scan(&id)

	if id == 0 {
		return 0, errors.New("gagal membuat survey_respondents")
	}
	return id, nil
}

func (r *seedAnswerRepo) InsertFieldResponse(formFieldID, formResponseID int64, answer string) error {
	return r.dbMaster.Exec(`
		INSERT INTO field_responses (form_field_id, form_response_id, answer, group_id, created_at, updated_at)
		VALUES (?, ?, ?, 0, now(), now())
	`, formFieldID, formResponseID, answer).Error
}
