package repository

import (
	"backend/reportapi/models"

	"gorm.io/gorm"
)

type SurveyExportRepo interface {
	GetSurveyWithResponses(surveyID uint, respondentIDs []uint) (*models.SurveyExport, error)
	GetFormAnswerOptions(fieldIDs []uint) (map[uint][]models.FormAnswerFieldExport, error)
	GetRespondentsByKelurahanIDs(kelurahanIDs []uint) ([]uint, error)
	GetRespondentsByRwIDs(rwIDs []uint) ([]uint, error)
	GetRespondentsByRtIDs(rtIDs []uint) ([]uint, error)
	GetKelurahanIDsByKecamatan(kecamatanID uint) ([]uint, error)
	GetRwIDsByKelurahan(kelurahanID uint) ([]uint, error)
	GetRtIDsByRw(rwID uint) ([]uint, error)
	GetKecamatanByID(id uint) (*models.KecamatanExport, error)
	GetKelurahanByID(id uint) (*models.KelurahanExport, error)
	GetRwByID(id uint) (*models.DataRwExport, error)
}

func NewSurveyExportRepo(dbSlave, dbMaster *gorm.DB) *surveyExportRepo {
	return &surveyExportRepo{dbSlave, dbMaster}
}

type surveyExportRepo struct {
	dbSlave  *gorm.DB
	dbMaster *gorm.DB
}

// GetSurveyWithResponses loads survey with nested respondent data, filtered by respondent IDs
func (r *surveyExportRepo) GetSurveyWithResponses(surveyID uint, respondentIDs []uint) (*models.SurveyExport, error) {
	var survey models.SurveyExport

	query := r.dbSlave.
		Preload("FlowDetail.Form.Fields").
		Preload("SurveyDataRespondents", func(db *gorm.DB) *gorm.DB {
			if len(respondentIDs) > 0 {
				db = db.Where("respondent_id IN ?", respondentIDs)
			}
			return db.Order("respondent_id ASC")
		}).
		Preload("SurveyDataRespondents.Respondent", func(db *gorm.DB) *gorm.DB {
			return db.
				Order("kecamatan_id ASC").
				Order("kelurahan_id ASC").
				Order("rw_id ASC").
				Order("rt_id ASC")
		}).
		Preload("SurveyDataRespondents.Respondent.Kecamatan").
		Preload("SurveyDataRespondents.Respondent.Kelurahan").
		Preload("SurveyDataRespondents.Respondent.RW").
		Preload("SurveyDataRespondents.Respondent.RT").
		Preload("SurveyDataRespondents.FieldDataResponses").
		Where("id = ?", surveyID).
		First(&survey)

	if query.Error != nil {
		return nil, query.Error
	}
	return &survey, nil
}

// GetFormAnswerOptions returns map[fieldID][]options for choice-type fields
func (r *surveyExportRepo) GetFormAnswerOptions(fieldIDs []uint) (map[uint][]models.FormAnswerFieldExport, error) {
	var options []models.FormAnswerFieldExport
	if err := r.dbSlave.Where("form_field_id IN ?", fieldIDs).Find(&options).Error; err != nil {
		return nil, err
	}
	result := make(map[uint][]models.FormAnswerFieldExport)
	for _, opt := range options {
		result[opt.FormFieldID] = append(result[opt.FormFieldID], opt)
	}
	return result, nil
}

func (r *surveyExportRepo) GetRespondentsByKelurahanIDs(kelurahanIDs []uint) ([]uint, error) {
	var ids []uint
	err := r.dbSlave.Model(&models.RespondentExport{}).
		Where("kelurahan_id IN ? AND role_id = 2", kelurahanIDs).
		Pluck("id", &ids).Error
	return ids, err
}

func (r *surveyExportRepo) GetRespondentsByRwIDs(rwIDs []uint) ([]uint, error) {
	var ids []uint
	err := r.dbSlave.Model(&models.RespondentExport{}).
		Where("rw_id IN ? AND role_id = 2", rwIDs).
		Pluck("id", &ids).Error
	return ids, err
}

func (r *surveyExportRepo) GetRespondentsByRtIDs(rtIDs []uint) ([]uint, error) {
	var ids []uint
	err := r.dbSlave.Model(&models.RespondentExport{}).
		Where("rt_id IN ? AND role_id = 2", rtIDs).
		Pluck("id", &ids).Error
	return ids, err
}

func (r *surveyExportRepo) GetKelurahanIDsByKecamatan(kecamatanID uint) ([]uint, error) {
	var ids []uint
	err := r.dbSlave.Model(&models.KelurahanExport{}).
		Where("sub_district_id = ?", kecamatanID).
		Pluck("id", &ids).Error
	return ids, err
}

func (r *surveyExportRepo) GetRwIDsByKelurahan(kelurahanID uint) ([]uint, error) {
	var ids []uint
	err := r.dbSlave.Model(&models.DataRwExport{}).
		Where("kelurahan_id = ?", kelurahanID).
		Pluck("id", &ids).Error
	return ids, err
}

func (r *surveyExportRepo) GetRtIDsByRw(rwID uint) ([]uint, error) {
	var ids []uint
	err := r.dbSlave.Model(&models.DataRtExport{}).
		Where("rw_id = ?", rwID).
		Pluck("id", &ids).Error
	return ids, err
}

func (r *surveyExportRepo) GetKecamatanByID(id uint) (*models.KecamatanExport, error) {
	var kecamatan models.KecamatanExport
	err := r.dbSlave.Where("id = ?", id).First(&kecamatan).Error
	return &kecamatan, err
}

func (r *surveyExportRepo) GetKelurahanByID(id uint) (*models.KelurahanExport, error) {
	var kelurahan models.KelurahanExport
	err := r.dbSlave.Preload("Kecamatan").Where("id = ?", id).First(&kelurahan).Error
	return &kelurahan, err
}

func (r *surveyExportRepo) GetRwByID(id uint) (*models.DataRwExport, error) {
	var rw models.DataRwExport
	err := r.dbSlave.Preload("Kelurahan.Kecamatan").Where("id = ?", id).First(&rw).Error
	return &rw, err
}
