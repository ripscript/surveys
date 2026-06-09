package repository

import (
	"backend/reportapi/models"
	"backend/reportapi/utils"
	"net/url"

	"gorm.io/gorm"
)

type SurveyRepo interface {
	LogSurveys(offset int, limit int, param url.Values) ([]models.LogSurveys, int64, error)
	GetSurveysByWilayahAndStatus(params models.GetSurveyParams) ([]models.Survey, error)
	GetSurveysBySurveyor(respondentID uint, status *string) ([]models.Survey, error)
	GetSurveysByDataRespondents(roleFilter models.DataRespondentFilter, wilayahFilter models.WilayahFilter, status *string) ([]models.Survey, error)
	GetFlowDetail(flowDetailID uint) (*models.FlowDetail, error)
	GetFlowFields(flowDetailID uint) ([]models.FlowField, error)
	GetFlowFieldsInSection(flowDetailID uint, sectionID uint) ([]models.FlowField, error)
	CountFormQuestions(flowFieldIDs []uint, excludeIDs []uint) (int64, error)
}

func NewSurveyRepo(dbSlave, dbMaster *gorm.DB) *surveyRepo {
	defer utils.GeneralRecover()
	return &surveyRepo{dbSlave, dbMaster}
}

type surveyRepo struct {
	dbSlave  *gorm.DB
	dbMaster *gorm.DB
}

func (r *surveyRepo) LogSurveys(offset int, limit int, param url.Values) ([]models.LogSurveys, int64, error) {
	defer utils.GeneralRecover()
	var data []models.LogSurveys
	var total int64
	db := r.dbSlave

	query := db.Model(data)

	if err := query.Model(&models.LogSurveys{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if limit != 1 {
		if err := query.Offset(offset).Limit(limit).Find(&data).Error; err != nil {
			return nil, 0, err
		}
	} else {
		if err := query.Offset(offset).Find(&data).Error; err != nil {
			return nil, 0, err
		}
	}

	for i := range data {
		data[i].No = offset + i + 1
	}
	return data, total, nil
}

func (r *surveyRepo) applyWilayahFilter(query *gorm.DB, f models.WilayahFilter) *gorm.DB {
	wilayahSubQuery := r.dbSlave.Table("survey_wilayahs").
		Select("survey_id").
		Where(
			r.dbSlave.Where("tingkat_wilayah = ? AND kecamatan_id = ?", 5, f.KecamatanID).
				Or("tingkat_wilayah = ? AND kecamatan_id = ? AND kelurahan_id = ?", 4, f.KecamatanID, f.KelurahanID).
				Or("tingkat_wilayah = ? AND kecamatan_id = ? AND kelurahan_id = ? AND rw_id = ?", 3, f.KecamatanID, f.KelurahanID, f.RwID),
		)

	return query.Where(
		r.dbSlave.Where("id IN (?)", wilayahSubQuery).
			Or("id NOT IN (SELECT survey_id FROM survey_wilayahs)"),
	)
}

func (r *surveyRepo) GetSurveysByWilayahAndStatus(params models.GetSurveyParams) ([]models.Survey, error) {
	var surveys []models.Survey
	query := r.dbSlave.
		Select("id, name, flow_detail_id, start_date, end_date, type, created_at, updated_at, status").
		Preload("ListFlowDetail").
		Preload("SurveyRespondents.FieldResponses").
		Preload("GetFlowDetail")

	if params.Status != nil {
		query = query.Where("status = ?", *params.Status)
	}

	query = r.applyWilayahFilter(query, params.WilayahFilter)
	query = query.Order("created_at DESC")

	if err := query.Find(&surveys).Error; err != nil {
		return nil, err
	}
	return surveys, nil
}

func (r *surveyRepo) GetSurveysBySurveyor(respondentID uint, status *string) ([]models.Survey, error) {
	var surveys []models.Survey
	query := r.dbSlave.Where(
		"id IN (SELECT survey_id FROM survey_surveyor WHERE respondent_id = ?)", respondentID,
	)
	if status != nil {
		query = query.Where("status = ?", *status)
	}
	if err := query.Order("created_at DESC").Find(&surveys).Error; err != nil {
		return nil, err
	}
	return surveys, nil
}

func (r *surveyRepo) GetSurveysByDataRespondents(roleFilter models.DataRespondentFilter, wilayahFilter models.WilayahFilter, status *string) ([]models.Survey, error) {
	var surveys []models.Survey

	subQuery := r.dbSlave.Table("survey_respondents sdr").
		Select("sdr.survey_id").
		Joins("JOIN respondents res ON res.id = sdr.respondent_id").
		Where("res.kecamatan_id = ?", roleFilter.KecamatanID)

	switch roleFilter.Role {
	case "rw":
		subQuery = subQuery.Where("res.kelurahan_id = ? AND res.rw_id = ?", roleFilter.KelurahanID, roleFilter.RwID)
	case "lurah":
		subQuery = subQuery.Where("res.kelurahan_id = ?", roleFilter.KelurahanID)
	}

	query := r.dbSlave.Preload("SurveyRespondents")
	// Where("id IN (?)", subQuery)

	if status != nil {
		query = query.Where("status = ?", *status)
	}

	wilayahSubQuery := r.dbSlave.Table("survey_wilayahs").
		Select("survey_id").
		Where(
			r.dbSlave.Where("tingkat_wilayah = ? AND kecamatan_id = ?", "5", wilayahFilter.KecamatanID).
				Or("tingkat_wilayah = ? AND kecamatan_id = ? AND kelurahan_id = ?", "4", wilayahFilter.KecamatanID, wilayahFilter.KelurahanID).
				Or("tingkat_wilayah = ? AND kecamatan_id = ? AND kelurahan_id = ? AND rw_id = ?", "3", wilayahFilter.KecamatanID, wilayahFilter.KelurahanID, wilayahFilter.RwID),
		)

	query = query.Where(
		r.dbSlave.Where("id IN (?)", wilayahSubQuery).
			Or("id NOT IN (SELECT survey_id FROM survey_wilayahs)"),
	)

	if err := query.Order("created_at DESC").Find(&surveys).Error; err != nil {
		return nil, err
	}
	return surveys, nil
}

func (r *surveyRepo) GetFlowDetail(flowDetailID uint) (*models.FlowDetail, error) {
	var fd models.FlowDetail
	if err := r.dbSlave.Where("id = ?", flowDetailID).First(&fd).Error; err != nil {
		return nil, err
	}
	return &fd, nil
}

func (r *surveyRepo) GetFlowFields(flowDetailID uint) ([]models.FlowField, error) {
	var fields []models.FlowField
	if err := r.dbSlave.Where("flow_detail_id = ?", flowDetailID).Find(&fields).Error; err != nil {
		return nil, err
	}
	return fields, nil
}

func (r *surveyRepo) GetFlowFieldsInSection(flowDetailID uint, sectionID uint) ([]models.FlowField, error) {
	var fields []models.FlowField
	if err := r.dbSlave.Where("flow_detail_id = ? AND section_id = ?", flowDetailID, sectionID).Find(&fields).Error; err != nil {
		return nil, err
	}
	return fields, nil
}

func (r *surveyRepo) CountFormQuestions(flowFieldIDs []uint, excludeIDs []uint) (int64, error) {
	var count int64
	query := r.dbSlave.Model(&models.FlowField{}).Where("form_field_id IN ?", flowFieldIDs)
	if len(excludeIDs) > 0 {
		query = query.Where("form_field_id NOT IN ?", excludeIDs)
	}
	if err := query.Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}
