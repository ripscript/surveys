package repository

import (
	"backend/surveyapi/enums"
	"backend/surveyapi/models"
	"backend/surveyapi/payloads"
	"backend/surveyapi/utils"
	"fmt"
	"strings"

	"gorm.io/gorm"
)

type SurveyRepo interface {
	RunInTransaction(fn func(txRepo SurveyRepo) error) error
	IsSurveyExistsByNameLower(name string) (bool, error)
	CreateSurvey(survey models.Survey) (*models.Survey, error)
	AssignSurveyorToSurvey(surveyorId int64, surveyId int64) (*models.SurveySurveyor, error)
	AssignWilayahToSurvey(surveyWilayah models.SurveyWilayah) (*models.SurveyWilayah, error)
	GetListSurvey(userLogin *models.Respondent, req payloads.SurveyDatatablePayload) ([]models.SurveyDatatableResponse, int64, error)
}

type surveyRepo struct {
	dbSlave  *gorm.DB
	dbMaster *gorm.DB
}

func NewSurveyRepo(dbSlave, dbMaster *gorm.DB) SurveyRepo {
	defer utils.GeneralRecover()
	return &surveyRepo{
		dbSlave,
		dbMaster,
	}
}

func (repository *surveyRepo) RunInTransaction(fn func(txRepo SurveyRepo) error) error {
	tx := repository.dbMaster.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	txRepo := &surveyRepo{
		dbSlave:  tx,
		dbMaster: tx,
	}

	err := fn(txRepo)
	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

func (repository *surveyRepo) CreateSurvey(survey models.Survey) (*models.Survey, error) {
	defer utils.GeneralRecover()

	err := repository.dbMaster.Create(&survey).Error
	if err != nil {
		return nil, err
	}
	return &survey, nil
}

func (repository *surveyRepo) AssignSurveyorToSurvey(surveyorId int64, surveyId int64) (*models.SurveySurveyor, error) {
	defer utils.GeneralRecover()

	surveySurveyor := models.SurveySurveyor{
		RespondentId: surveyorId,
		SurveyId:     int(surveyId),
	}

	err := repository.dbMaster.Create(&surveySurveyor).Error
	if err != nil {
		return nil, err
	}
	return &surveySurveyor, nil
}

func (repository *surveyRepo) AssignWilayahToSurvey(surveyWilayah models.SurveyWilayah) (*models.SurveyWilayah, error) {
	defer utils.GeneralRecover()

	err := repository.dbMaster.Create(&surveyWilayah).Error
	if err != nil {
		return nil, err
	}
	return &surveyWilayah, nil
}

func (repository *surveyRepo) IsSurveyExistsByNameLower(name string) (bool, error) {
	defer utils.GeneralRecover()

	var count int64
	err := repository.dbSlave.Model(&models.Survey{}).Where("LOWER(name) = ?", strings.ToLower(name)).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (repository *surveyRepo) GetListSurvey(userLogin *models.Respondent, req payloads.SurveyDatatablePayload) ([]models.SurveyDatatableResponse, int64, error) {
	defer utils.GeneralRecover()
	var data []models.SurveyDatatableResponse
	var totalData int64

	db := repository.dbSlave.Table("surveys").
		Joins("LEFT JOIN users ON users.id = surveys.created_by").
		Joins("LEFT JOIN flow_details ON flow_details.id = surveys.flow_detail_id")

	if req.SurveyDiikuti {
		if userLogin.RoleId != nil {
			if *userLogin.RoleId != int64(enums.ROLE_ADMIN) {
				if *userLogin.RoleId == int64(enums.ROLE_KECAMATAN) {
					tingkatWilayahStr := fmt.Sprintf("%d", enums.KECAMATAN)
					db = db.Where(`EXISTS (
						SELECT 1 FROM survey_wilayahs 
						WHERE survey_wilayahs.survey_id = surveys.id 
						AND survey_wilayahs.tingkat_wilayah = ? 
						AND survey_wilayahs.kecamatan_id = ?
					)`, tingkatWilayahStr, userLogin.KecamatanId)
				} else {
					db = db.Where("1 = 0")
				}
			}
		}
	} else {
		db = db.Where("surveys.created_by = ?", userLogin.ID)
	}

	if req.Search != "" {
		searchTerm := "%" + req.Search + "%"
		searchStr := strings.TrimSpace(req.Search)
		parsedDate, isDate := utils.TryParseIndonesianDate(req.Search)

		if isDate {
			db = db.Where(`
				surveys.name ILIKE ? OR
				flow_details.name ILIKE ? OR
				users.first_name ILIKE ? OR
				DATE(surveys.start_date) = ? OR
				DATE(surveys.end_date) = ? OR
				DATE(flow_details.created_at) = ? OR
				DATE(flow_details.updated_at) = ?
			`, searchTerm, searchTerm, searchTerm, parsedDate, parsedDate, parsedDate, parsedDate)
		} else if len(searchStr) == 4 {
			db = db.Where(`
				surveys.name ILIKE ? OR
				flow_details.name ILIKE ? OR
				users.first_name ILIKE ? OR
				EXTRACT(YEAR FROM surveys.start_date)::TEXT = ? OR
				EXTRACT(YEAR FROM surveys.end_date)::TEXT = ? OR
				EXTRACT(YEAR FROM flow_details.created_at)::TEXT = ? OR
				EXTRACT(YEAR FROM flow_details.updated_at)::TEXT = ?
			`, searchTerm, searchTerm, searchTerm, searchStr, searchStr, searchStr, searchStr)
		} else {
			db = db.Where(`
				surveys.name ILIKE ? OR
				flow_details.name ILIKE ? OR
				users.first_name ILIKE ?
			`, searchTerm, searchTerm, searchTerm)
		}
	}

	err := db.Count(&totalData).Error
	if err != nil {
		return nil, 0, err
	}

	db = db.Select(`
		surveys.id,
		surveys.name AS survey_name,
		surveys.start_date,
		surveys.end_date,
		flow_details.name AS flow_name,
		surveys.created_at,
		surveys.updated_at,
		surveys.created_by,
		users.first_name AS created_by_name,
		(SELECT COUNT(id) FROM survey_respondents WHERE survey_respondents.survey_id = surveys.id) AS total_responden,
		surveys.status,
		surveys.approval_survey
	`)

	if req.OrderBy != "" {
		finalOrderBy := "surveys.created_at"
		finalOrderDir := "desc"

		allowedOrderCols := map[string]string{
			"id":              "surveys.id",
			"survey_name":     "surveys.name",
			"start_date":      "surveys.start_date",
			"end_date":        "surveys.end_date",
			"flow_name":       "flow_details.name",
			"created_at":      "surveys.created_at",
			"updated_at":      "surveys.updated_at",
			"created_by_name": "users.first_name",
			"total_responden": "total_responden",
			"status":          "surveys.status",
			"approval_survey": "surveys.approval_survey",
		}

		if mappedCol, isAllowed := allowedOrderCols[req.OrderBy]; isAllowed {
			finalOrderBy = mappedCol
		}

		if strings.ToLower(req.OrderDir) == "asc" {
			finalOrderDir = "asc"
		}

		db = db.Order(finalOrderBy + " " + finalOrderDir)
	} else {
		db = db.Order("surveys.created_at desc")
	}

	offset := (req.Page - 1) * req.Limit
	err = db.Limit(req.Limit).Offset(offset).Find(&data).Error
	if err != nil {
		return nil, 0, err
	}

	return data, totalData, nil
}
