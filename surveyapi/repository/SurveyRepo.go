package repository

import (
	"backend/surveyapi/dto"
	"backend/surveyapi/enums"
	"backend/surveyapi/models"
	"backend/surveyapi/payloads"
	"backend/surveyapi/utils"
	"context"
	"errors"
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
	GetListSurvey(userLogin models.JwtCustomClaims, respondentLogin *models.Respondent, req payloads.SurveyDatatablePayload) ([]models.SurveyDatatableResponse, int64, error)
	GetSurveyById(surveyId int64) (*models.Survey, error)
	UpdateSurvey(survey *models.Survey) error

	GetListSurveyWilayah(userLogin models.JwtCustomClaims, respondentLogin *models.Respondent, req payloads.SurveyWilayahDatatablePayload) ([]models.SurveyWilayahDatatableResponse, int64, error)

	GetRespondentExistsInSurvey(respondentId int64, surveyId int64) (*models.SurveyRespondent, error)
	IsRespondentExistsInSurvey(respondentId int64) (bool, error)

	BeginTransaction() *gorm.DB
	CreateSurveyRespondent(tx *gorm.DB, surveyRespondent models.SurveyRespondent) (*models.SurveyRespondent, error)
	BulkInsertFieldResponses(ctx context.Context, tx *gorm.DB, fieldResponses []models.FieldResponse) error
	CheckRespondentEligibility(ctx context.Context, tx *gorm.DB, surveyID int64, respondent *models.Respondent) (bool, error)
	GetAnswersByResponseID(formResponseID int64) ([]models.FieldResponse, error)

	GetOldResponsesBySection(tx *gorm.DB, formResponseID int64, fieldIDs []int64) ([]models.FieldResponse, error)
	DeleteOldResponsesBySection(tx *gorm.DB, formResponseID int64, fieldIDs []int64) error
	UpdateStatusSurvey(ctx context.Context, tx *gorm.DB, formResponseID int64, updatedStatus int) error
	GetUnansweredRequiredCount(tx *gorm.DB, respondentID int64, surveyID int64) (int, error)
	MakeLogSurvey(ctx context.Context, tx *gorm.DB, logSurvey models.LogSurvey) error
	MakeLogSurveyBulk(ctx context.Context, tx *gorm.DB, logSurveys []models.LogSurvey) error

	GetHistoryApproval(ctx context.Context, tx *gorm.DB, surveyID int64) ([]models.LogSurveyDetail, error)
	GetHistoryApprovalPerWilayah(ctx context.Context, tx *gorm.DB, surveyID int64, rtId int64) ([]models.LogSurveyDetail, error)
	GetSurveyKewilayahan(userLogin models.JwtCustomClaims, respondentLogin *models.Respondent, req payloads.SurveyWilayahDatatablePayload) ([]models.SurveyKewilayahanDatatableResponse, int64, error)

	GetStatusKeterisianBulkRT(ctx context.Context, surveyID int64, rtIDs []int64) (map[int64]string, error)
	GetStatusKeterisianBulkRW(ctx context.Context, surveyID int64, rwIDs []int64) (map[int64]string, error)
	GetStatusKeterisianBulkKelurahan(ctx context.Context, surveyID int64, kelurahanIDs []int64) (map[int64]string, error)
	GetRespondentSurveyByRTId(ctx context.Context, surveyID int64, rtID int64) (*models.SurveyRespondent, error)

	WipeAndReplaceAllFieldResponses(ctx context.Context, respondentID int64, newResponses []models.FieldResponse) error
	GetAllOldResponsesByRespondent(ctx context.Context, respondentID int64) ([]models.FieldResponse, error)
	CreateFlaggingEdit(ctx context.Context, flags []models.FlaggingEditPertanyaanSurvey) error
	ResetFlaggingEditStatus(ctx context.Context, respondentID int64) error
	UpdateRespondentApprovalStatus(ctx context.Context, respondentID int64, status int, statusApproval string, isEditPertanyaan string) error
	ClearAndCreateFlaggingEdit(ctx context.Context, respondentID int64, flags []models.FlaggingEditPertanyaanSurvey) error

	UpdateRespondentApprovalState(ctx context.Context, tx *gorm.DB, respondentID int64, status int, statusApproval *string, isEditPertanyaan string) error
	ResetFlaggingEditStatusTx(ctx context.Context, tx *gorm.DB, respondentID int64) error

	GetSurveyCompletionHistory(ctx context.Context, surveyID int64, respondentID int64) (*models.LogSurveyDetail, error)
	GetSurveyVerificationHistory(ctx context.Context, surveyID int64, respondentID int64) (*models.LogSurveyDetail, error)
	GetSurveyValidationHistory(ctx context.Context, surveyID int64, respondentID int64) (*models.LogSurveyDetail, error)

	GetRawJawabanForExport(ctx context.Context, surveyID int64, rtID int64) ([]dto.ExportRawJawabanDTO, error)
	ValidateIsImageExists(ctx context.Context, tx *gorm.DB, surveyRespondentId int64, filePath string) (bool, error)

	GetRawJawabanWilayahForExport(ctx context.Context, surveyID int64, level int, wilayahID int64) ([]dto.ExportRawJawabanWilayah, error)

	GetSurveyWilayahsBySurveyId(surveyId int64) ([]models.SurveyWilayah, error)
	GetSurveyorsBySurveyId(surveyId int64) ([]models.SurveySurveyor, error)
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

func (repository *surveyRepo) GetListSurvey(userLogin models.JwtCustomClaims, respondentLogin *models.Respondent, req payloads.SurveyDatatablePayload) ([]models.SurveyDatatableResponse, int64, error) {
	defer utils.GeneralRecover()
	var data []models.SurveyDatatableResponse
	var totalData int64

	db := repository.dbSlave.Table("surveys").
		Joins("LEFT JOIN users ON users.id = surveys.created_by").
		Joins("LEFT JOIN flow_details ON flow_details.id = surveys.flow_detail_id")

	if req.SurveyDiikuti {
		if respondentLogin.RoleId != nil {
			if *respondentLogin.RoleId != int64(enums.ROLE_ADMIN) {

				var kecId, kelId, rwId int64
				if respondentLogin.KecamatanId != nil {
					kecId = *respondentLogin.KecamatanId
				}
				if respondentLogin.KelurahanId != nil {
					kelId = *respondentLogin.KelurahanId
				}
				if respondentLogin.RWId != nil {
					rwId = *respondentLogin.RWId
				}

				db = db.Where(`
				EXISTS (
					SELECT 1 FROM survey_wilayahs 
					WHERE survey_wilayahs.survey_id = surveys.id 
					AND survey_wilayahs.kecamatan_id = ?
					AND (survey_wilayahs.kelurahan_id IS NULL OR survey_wilayahs.kelurahan_id = ?)
					AND (survey_wilayahs.rw_id IS NULL OR survey_wilayahs.rw_id = ?)
				) 
				OR NOT EXISTS (
					SELECT 1 FROM survey_wilayahs 
					WHERE survey_wilayahs.survey_id = surveys.id
				)`, kecId, kelId, rwId)
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

	for i := range data {
		data[i].No = int64(offset + i + 1)
	}

	return data, totalData, nil
}

func (repository *surveyRepo) GetSurveyById(surveyId int64) (*models.Survey, error) {
	defer utils.GeneralRecover()

	var survey models.Survey
	err := repository.dbSlave.Where("id = ?", surveyId).First(&survey).Error
	if err != nil {
		return nil, err
	}
	return &survey, nil
}

func (repository *surveyRepo) UpdateSurvey(survey *models.Survey) error {
	defer utils.GeneralRecover()

	err := repository.dbMaster.Save(survey).Error
	if err != nil {
		return err
	}
	return nil
}

func (repository *surveyRepo) GetListSurveyWilayah(userLogin models.JwtCustomClaims, respondentLogin *models.Respondent, req payloads.SurveyWilayahDatatablePayload) ([]models.SurveyWilayahDatatableResponse, int64, error) {
	defer utils.GeneralRecover()
	var data []models.SurveyWilayahDatatableResponse
	var totalData int64

	db := repository.dbSlave.Table("surveys").
		Joins("LEFT JOIN users ON users.id = surveys.created_by").
		Joins("LEFT JOIN flow_details ON flow_details.id = surveys.flow_detail_id").
		Where("surveys.approval_survey = 'approved' OR surveys.approval_survey = 'non_approval'")

	if req.StatusSurvey == string(enums.STATUS_SURVEY_UPCOMING) {
		db = db.Where("surveys.start_date > NOW()")
	} else if req.StatusSurvey == string(enums.STATUS_SURVEY_ONGOING) {
		db = db.Where("surveys.start_date <= NOW() AND surveys.end_date >= NOW()")
	} else if req.StatusSurvey == string(enums.STATUS_SURVEY_FINISHED) {
		db = db.Where("surveys.end_date < NOW()")
	}

	if respondentLogin.RoleId != nil {
		if *respondentLogin.RoleId != int64(enums.ROLE_ADMIN) {

			var kecId, kelId, rwId int64
			if respondentLogin.KecamatanId != nil {
				kecId = *respondentLogin.KecamatanId
			}
			if respondentLogin.KelurahanId != nil {
				kelId = *respondentLogin.KelurahanId
			}
			if respondentLogin.RWId != nil {
				rwId = *respondentLogin.RWId
			}

			db = db.Where(`
			EXISTS (
				SELECT 1 FROM survey_wilayahs 
				WHERE survey_wilayahs.survey_id = surveys.id 
				AND survey_wilayahs.kecamatan_id = ?
				AND (survey_wilayahs.kelurahan_id IS NULL OR survey_wilayahs.kelurahan_id = ?)
				AND (survey_wilayahs.rw_id IS NULL OR survey_wilayahs.rw_id = ?)
			) 
			OR NOT EXISTS (
				SELECT 1 FROM survey_wilayahs 
				WHERE survey_wilayahs.survey_id = surveys.id
			)`, kecId, kelId, rwId)
		}
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

	var respID int64 = 0
	if respondentLogin != nil {
		respID = respondentLogin.ID
	}

	selectQuery := fmt.Sprintf(`
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
		surveys.approval_survey,

		(
			SELECT COUNT(DISTINCT form_field_id) 
			FROM flow_fields 
			WHERE flow_detail_id = surveys.flow_detail_id
		) AS jumlah_total_soal,

		(
			SELECT COUNT(DISTINCT fr.form_field_id) 
			FROM field_responses fr 
			INNER JOIN survey_respondents sr ON sr.id = fr.form_response_id 
			WHERE sr.survey_id = surveys.id AND sr.respondent_id = %d
		) AS jumlah_soal_terisi,

		CASE 
			WHEN (
				SELECT COUNT(DISTINCT fr.form_field_id) 
				FROM field_responses fr 
				INNER JOIN survey_respondents sr ON sr.id = fr.form_response_id 
				WHERE sr.survey_id = surveys.id AND sr.respondent_id = %d
			) = 0 THEN 'Belum Terisi'
			
			WHEN (
				SELECT COUNT(DISTINCT fr.form_field_id) 
				FROM field_responses fr 
				INNER JOIN survey_respondents sr ON sr.id = fr.form_response_id 
				WHERE sr.survey_id = surveys.id AND sr.respondent_id = %d
			) >= (
				SELECT COUNT(DISTINCT form_field_id) 
				FROM flow_fields 
				WHERE flow_detail_id = surveys.flow_detail_id
			) THEN 'Sudah Terisi' 

			ELSE 'Sedang Berjalan' 
		END AS status_keterisian,

		(
			SELECT CASE 
				-- Pengecekan status revisi HARUS di atas agar dicek lebih dulu
				WHEN status = 2 AND status_approval = 'revisi_rt' THEN 'revisi'
				WHEN status = 2 THEN 'selesai'
				WHEN status = 1 THEN 'draft'
			END
			FROM survey_respondents 
			WHERE survey_id = surveys.id AND respondent_id = %d
			LIMIT 1
		) AS status_respondent

	`, respID, respID, respID, respID)

	db = db.Select(selectQuery)

	if req.OrderBy != "" {
		finalOrderBy := "surveys.created_at"
		finalOrderDir := "desc"

		allowedOrderCols := map[string]string{
			"id":                 "surveys.id",
			"survey_name":        "surveys.name",
			"start_date":         "surveys.start_date",
			"end_date":           "surveys.end_date",
			"flow_name":          "flow_details.name",
			"created_at":         "surveys.created_at",
			"updated_at":         "surveys.updated_at",
			"created_by_name":    "users.first_name",
			"total_responden":    "total_responden",
			"status":             "surveys.status",
			"approval_survey":    "surveys.approval_survey",
			"status_keterisian":  "status_keterisian",
			"jumlah_soal_terisi": "jumlah_soal_terisi",
			"jumlah_total_soal":  "jumlah_total_soal",
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

	for i := range data {
		data[i].No = int64(offset + i + 1)
	}

	return data, totalData, nil
}

func (repository *surveyRepo) GetRespondentExistsInSurvey(respondentId int64, surveyId int64) (*models.SurveyRespondent, error) {
	defer utils.GeneralRecover()

	var surveyRespondent models.SurveyRespondent
	err := repository.dbSlave.Where("respondent_id = ? AND survey_id = ?", respondentId, surveyId).First(&surveyRespondent).Error
	if err != nil {
		return nil, err
	}
	return &surveyRespondent, nil
}

func (repository *surveyRepo) IsRespondentExistsInSurvey(respondentId int64) (bool, error) {
	defer utils.GeneralRecover()

	var count int64
	err := repository.dbSlave.Model(&models.SurveyRespondent{}).Where("respondent_id = ?", respondentId).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (repository *surveyRepo) BeginTransaction() *gorm.DB {
	return repository.dbMaster.Begin()
}

func (repository *surveyRepo) CreateSurveyRespondent(tx *gorm.DB, surveyRespondent models.SurveyRespondent) (*models.SurveyRespondent, error) {
	defer utils.GeneralRecover()

	err := tx.Create(&surveyRespondent).Error
	if err != nil {
		return nil, err
	}
	return &surveyRespondent, nil
}

func (repository *surveyRepo) BulkInsertFieldResponses(ctx context.Context, tx *gorm.DB, fieldResponses []models.FieldResponse) error {
	defer utils.GeneralRecover()

	err := tx.WithContext(ctx).Create(&fieldResponses).Error
	if err != nil {
		return err
	}

	return nil
}

func (repository *surveyRepo) CheckRespondentEligibility(ctx context.Context, tx *gorm.DB, surveyID int64, respondent *models.Respondent) (bool, error) {
	defer utils.GeneralRecover()

	if respondent == nil {
		return false, errors.New("data respondent tidak valid")
	}

	db := tx
	if db == nil {
		db = repository.dbSlave
	}

	var count int64

	// Ekstrak ID wilayah dengan aman untuk menghindari panic nil pointer
	var kecId, kelId, rwId int64
	if respondent.KecamatanId != nil {
		kecId = *respondent.KecamatanId
	}
	if respondent.KelurahanId != nil {
		kelId = *respondent.KelurahanId
	}
	if respondent.RWId != nil {
		rwId = *respondent.RWId
	}

	// Query validasi hierarkis (Universal vs Spesifik Wilayah)
	err := db.WithContext(ctx).Table("surveys").
		Where("surveys.id = ?", surveyID).
		Where(`
			EXISTS (
				SELECT 1 FROM survey_wilayahs 
				WHERE survey_wilayahs.survey_id = surveys.id 
				AND survey_wilayahs.kecamatan_id = ?
				AND (survey_wilayahs.kelurahan_id IS NULL OR survey_wilayahs.kelurahan_id = ?)
				AND (survey_wilayahs.rw_id IS NULL OR survey_wilayahs.rw_id = ?)
			) 
			OR NOT EXISTS (
				SELECT 1 FROM survey_wilayahs 
				WHERE survey_wilayahs.survey_id = surveys.id
			)
		`, kecId, kelId, rwId).
		Count(&count).Error

	if err != nil {
		return false, err
	}

	// Jika count > 0, berarti responden berhak (eligible)
	return count > 0, nil
}

func (repository *surveyRepo) GetAnswersByResponseID(formResponseID int64) ([]models.FieldResponse, error) {
	defer utils.GeneralRecover()

	var answers []models.FieldResponse
	err := repository.dbSlave.Where("form_response_id = ?", formResponseID).Find(&answers).Error
	if err != nil {
		return nil, err
	}
	return answers, nil
}

func (repository *surveyRepo) GetOldResponsesBySection(tx *gorm.DB, formResponseID int64, fieldIDs []int64) ([]models.FieldResponse, error) {
	if tx == nil {
		tx = repository.dbSlave
	}

	var oldResponses []models.FieldResponse
	err := tx.Where("form_response_id = ? AND form_field_id IN ?", formResponseID, fieldIDs).Find(&oldResponses).Error
	return oldResponses, err
}

func (repository *surveyRepo) DeleteOldResponsesBySection(tx *gorm.DB, formResponseID int64, fieldIDs []int64) error {
	return tx.Where("form_response_id = ? AND form_field_id IN ?", formResponseID, fieldIDs).Delete(&models.FieldResponse{}).Error
}

func (repository *surveyRepo) UpdateStatusSurvey(ctx context.Context, tx *gorm.DB, formResponseID int64, updatedStatus int) error {
	defer utils.GeneralRecover()

	db := tx
	if db == nil {
		db = repository.dbMaster
	}
	if ctx != nil {
		db = db.WithContext(ctx)
	}

	// Hanya meng-update kolom "status" pada record yang spesifik
	return db.Model(&models.SurveyRespondent{}).
		Where("id = ?", formResponseID).
		Update("status", updatedStatus).Error
}

func (r *surveyRepo) GetUnansweredRequiredCount(tx *gorm.DB, respondentID int64, surveyID int64) (int, error) {
	db := tx
	if db == nil {
		db = r.dbSlave
	}

	var totalRequired int64
	err := db.Table("form_fields").
		Joins("JOIN flow_fields ON flow_fields.form_field_id = form_fields.id").
		Joins("JOIN surveys ON surveys.flow_detail_id = flow_fields.flow_detail_id").
		Where("surveys.id = ? AND form_fields.required = true", surveyID).
		Count(&totalRequired).Error

	if err != nil {
		return 0, err
	}

	// 2. Hitung berapa banyak soal wajib yang sudah dijawab oleh responden ini
	var answeredRequired int64
	err = db.Table("field_responses").
		Joins("JOIN form_fields ON field_responses.form_field_id = form_fields.id").
		Joins("JOIN survey_respondents sr ON field_responses.form_response_id = sr.id").
		Where("sr.survey_id = ? AND sr.respondent_id = ? AND form_fields.required = true", surveyID, respondentID).
		Count(&answeredRequired).Error

	if err != nil {
		return 0, err
	}

	// Sisanya adalah soal wajib yang belum dijawab
	return int(totalRequired - answeredRequired), nil
}

func (repository *surveyRepo) MakeLogSurvey(ctx context.Context, tx *gorm.DB, logSurvey models.LogSurvey) error {
	defer utils.GeneralRecover()

	if tx == nil {
		tx = repository.dbMaster
	}

	err := tx.WithContext(ctx).Create(&logSurvey).Error
	if err != nil {
		return err
	}

	return nil
}

func (repository *surveyRepo) MakeLogSurveyBulk(ctx context.Context, tx *gorm.DB, logSurveys []models.LogSurvey) error {
	defer utils.GeneralRecover()

	if tx == nil {
		tx = repository.dbMaster
	}

	err := tx.WithContext(ctx).Create(&logSurveys).Error
	if err != nil {
		return err
	}

	return nil
}

func (repository *surveyRepo) GetHistoryApproval(ctx context.Context, tx *gorm.DB, surveyID int64) ([]models.LogSurveyDetail, error) {
	defer utils.GeneralRecover()

	if tx == nil {
		tx = repository.dbSlave
	}

	var logs []models.LogSurveyDetail
	err := tx.WithContext(ctx).
		Joins("JOIN respondents ON respondents.id = log__surveys.respondent_id").
		Where("log__surveys.survey_id = ?", surveyID).
		Order("log__surveys.created_at DESC").
		Find(&logs).Error
	if err != nil {
		return nil, err
	}

	return logs, nil
}

func (repository *surveyRepo) GetHistoryApprovalPerWilayah(ctx context.Context, tx *gorm.DB, surveyID int64, rtId int64) ([]models.LogSurveyDetail, error) {
	defer utils.GeneralRecover()

	if tx == nil {
		tx = repository.dbSlave
	}

	var logs []models.LogSurveyDetail
	err := tx.WithContext(ctx).
		Joins("JOIN respondents ON respondents.id = log__surveys.respondent_id").
		Where("log__surveys.survey_id = ? AND respondents.rt_id = ?", surveyID, rtId).
		Order("log__surveys.created_at DESC").
		Find(&logs).Error
	if err != nil {
		return nil, err
	}

	return logs, nil
}

func (repository *surveyRepo) GetSurveyKewilayahan(userLogin models.JwtCustomClaims, respondentLogin *models.Respondent, req payloads.SurveyWilayahDatatablePayload) ([]models.SurveyKewilayahanDatatableResponse, int64, error) {
	defer utils.GeneralRecover()
	var data []models.SurveyKewilayahanDatatableResponse
	var totalData int64

	db := repository.dbSlave.Table("surveys").
		Joins("LEFT JOIN users ON users.id = surveys.created_by").
		Joins("LEFT JOIN flow_details ON flow_details.id = surveys.flow_detail_id").
		Where("surveys.approval_survey = 'approved' OR surveys.approval_survey = 'non_approval'")

	if req.StatusSurvey == string(enums.STATUS_SURVEY_UPCOMING) {
		db = db.Where("surveys.start_date > NOW()")
	} else if req.StatusSurvey == string(enums.STATUS_SURVEY_ONGOING) {
		db = db.Where("surveys.start_date <= NOW() AND surveys.end_date >= NOW()")
	} else if req.StatusSurvey == string(enums.STATUS_SURVEY_FINISHED) {
		db = db.Where("surveys.end_date < NOW()")
	}

	if respondentLogin.RoleId != nil {
		if *respondentLogin.RoleId != int64(enums.ROLE_ADMIN) {

			var kecId, kelId, rwId int64
			if respondentLogin.KecamatanId != nil {
				kecId = *respondentLogin.KecamatanId
			}
			if respondentLogin.KelurahanId != nil {
				kelId = *respondentLogin.KelurahanId
			}
			if respondentLogin.RWId != nil {
				rwId = *respondentLogin.RWId
			}

			db = db.Where(`
			EXISTS (
				SELECT 1 FROM survey_wilayahs 
				WHERE survey_wilayahs.survey_id = surveys.id 
				AND survey_wilayahs.kecamatan_id = ?
				AND (survey_wilayahs.kelurahan_id IS NULL OR survey_wilayahs.kelurahan_id = ?)
				AND (survey_wilayahs.rw_id IS NULL OR survey_wilayahs.rw_id = ?)
			) 
			OR NOT EXISTS (
				SELECT 1 FROM survey_wilayahs 
				WHERE survey_wilayahs.survey_id = surveys.id
			)`, kecId, kelId, rwId)
		}
	}

	if req.Search != "" {
		searchTerm := "%" + req.Search + "%"
		searchStr := strings.TrimSpace(req.Search)
		parsedDate, isDate := utils.TryParseIndonesianDate(req.Search)

		if isDate {
			db = db.Where(`
				surveys.name ILIKE ? OR
				DATE(surveys.start_date) = ? OR
				DATE(surveys.end_date) = ?
			`, searchTerm, parsedDate, parsedDate)
		} else if len(searchStr) == 4 {
			db = db.Where(`
				surveys.name ILIKE ? OR
				EXTRACT(YEAR FROM surveys.start_date)::TEXT = ? OR
				EXTRACT(YEAR FROM surveys.end_date)::TEXT = ?
			`, searchTerm, searchStr, searchStr)
		} else {
			db = db.Where(`
				surveys.name ILIKE ?
			`, searchTerm)
		}
	}

	err := db.Count(&totalData).Error
	if err != nil {
		return nil, 0, err
	}

	selectQuery := fmt.Sprintf(`
		surveys.id,
		surveys.name AS survey_name,
		surveys.start_date,
		surveys.end_date,
		surveys.status
	`)

	db = db.Select(selectQuery)

	if req.OrderBy != "" {
		finalOrderBy := "surveys.created_at"
		finalOrderDir := "desc"

		allowedOrderCols := map[string]string{
			"id":          "surveys.id",
			"survey_name": "surveys.name",
			"start_date":  "surveys.start_date",
			"end_date":    "surveys.end_date",
			"status":      "surveys.status",
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

	for i := range data {
		data[i].No = int64(offset + i + 1)
	}

	return data, totalData, nil
}

func (repository *surveyRepo) GetStatusKeterisianBulkRT(ctx context.Context, surveyID int64, rtIDs []int64) (map[int64]string, error) {
	var results []struct {
		RtId           int64
		StatusApproval *string
	}

	var countWilayah int64
	repository.dbSlave.Table("survey_wilayahs").Where("survey_id = ?", surveyID).Count(&countWilayah)

	query := repository.dbSlave.Table("survey_respondents").
		Select("respondents.rt_id, MAX(survey_respondents.status_approval) as status_approval").
		Joins("JOIN respondents ON respondents.id = survey_respondents.respondent_id").
		Where("survey_respondents.survey_id = ?", surveyID).
		Where("respondents.rt_id IN ?", rtIDs).
		Where("respondents.role_id = ?", int64(enums.ROLE_RT))

	if countWilayah > 0 {
		query = query.Where(`
			EXISTS (
				SELECT 1 FROM survey_wilayahs sw 
				WHERE sw.survey_id = survey_respondents.survey_id 
				AND sw.kecamatan_id = respondents.kecamatan_id
				AND (sw.kelurahan_id IS NULL OR sw.kelurahan_id = respondents.kelurahan_id)
				AND (sw.rw_id IS NULL OR sw.rw_id = respondents.rw_id)
			)
		`)
	}

	err := query.Group("respondents.rt_id").Find(&results).Error

	if err != nil {
		return nil, err
	}

	statusMap := make(map[int64]string)

	for _, rtID := range rtIDs {
		statusMap[rtID] = "Tidak ada responden"
	}

	for _, res := range results {
		if res.StatusApproval != nil {
			switch *res.StatusApproval {
			case "verified_rw":
				statusMap[res.RtId] = "Sudah Diverifikasi Oleh RW"
			case "validated_lurah":
				statusMap[res.RtId] = "Sudah Divalidasi Oleh Kelurahan"
			case "revisi_rt":
				statusMap[res.RtId] = "Sedang Proses Revisi Oleh RT"
			case "revisi_rw":
				statusMap[res.RtId] = "Sedang Proses Revisi Oleh RW"
			default:
				statusMap[res.RtId] = *res.StatusApproval
			}
		} else {
			statusMap[res.RtId] = "Menunggu Verifikasi Rw"
		}
	}

	return statusMap, nil
}

func (repository *surveyRepo) GetStatusKeterisianBulkRW(ctx context.Context, surveyID int64, rwIDs []int64) (map[int64]string, error) {
	var results []struct {
		RwId             int64
		TotalRespondents int
		TotalValidated   int
	}

	// 1. Cek apakah survei ini memiliki spesifik wilayah (Targeted) atau tidak (Universal)
	var countWilayah int64
	repository.dbSlave.Table("survey_wilayahs").Where("survey_id = ?", surveyID).Count(&countWilayah)

	// Mulai merakit Query utama
	query := repository.dbSlave.Table("survey_respondents").
		Select(`
			respondents.rw_id, 
			COUNT(survey_respondents.id) as total_respondents, 
			COUNT(CASE WHEN survey_respondents.status_approval = 'validated_lurah' THEN 1 END) as total_validated
		`).
		Joins("JOIN respondents ON respondents.id = survey_respondents.respondent_id").
		Where("survey_respondents.survey_id = ?", surveyID).
		Where("respondents.rw_id IN ?", rwIDs).
		// 🆕 PENGAMANAN ROLE: Hanya hitung jika respondennya sah (misal: Role ID = 2 untuk RT)
		// Sesuaikan "enums.ROLE_RT" dengan role target responden survei Anda.
		Where("respondents.role_id = ?", int64(enums.ROLE_RT))

	// 2. Jika survei bersifat Targeted (ada record di survey_wilayahs)
	// Pastikan hanya menghitung responden yang kecocokan wilayahnya terdaftar di survey_wilayahs
	if countWilayah > 0 {
		query = query.Where(`
			EXISTS (
				SELECT 1 FROM survey_wilayahs sw 
				WHERE sw.survey_id = survey_respondents.survey_id 
				AND sw.kecamatan_id = respondents.kecamatan_id
				AND (sw.kelurahan_id IS NULL OR sw.kelurahan_id = respondents.kelurahan_id)
				AND (sw.rw_id IS NULL OR sw.rw_id = respondents.rw_id)
			)
		`)
	}

	// Eksekusi akhir query dengan Group By
	err := query.Group("respondents.rw_id").Find(&results).Error

	if err != nil {
		return nil, err
	}

	statusMap := make(map[int64]string)

	// 1. SET DEFAULT: Berikan nilai "Tidak ada responden" untuk semua RW di halaman ini
	for _, rwID := range rwIDs {
		statusMap[rwID] = "Tidak ada responden"
	}

	// 2. TIMPA STATUS: Evaluasi berdasarkan perbandingan jumlah aggregasi
	for _, res := range results {
		// Jika jumlah total responden SAMA DENGAN jumlah responden yang sudah divalidasi lurah
		if res.TotalRespondents == res.TotalValidated {
			statusMap[res.RwId] = "Selesai"
		} else {
			// Jika ada selisih (berarti ada minimal 1 yang belum divalidasi lurah)
			statusMap[res.RwId] = "Sedang Proses Verifikasi dan Validasi"
		}
	}

	return statusMap, nil
}

func (repository *surveyRepo) GetStatusKeterisianBulkKelurahan(ctx context.Context, surveyID int64, kelurahanIDs []int64) (map[int64]string, error) {
	// Struct penampung hasil agregasi dari database
	var results []struct {
		KelurahanId      int64
		TotalRespondents int
		TotalValidated   int
	}

	// 1. Cek apakah survei ini memiliki spesifik wilayah (Targeted) atau tidak (Universal)
	var countWilayah int64
	repository.dbSlave.Table("survey_wilayahs").Where("survey_id = ?", surveyID).Count(&countWilayah)

	// Mulai merakit Query utama
	query := repository.dbSlave.Table("survey_respondents").
		Select(`
			respondents.kelurahan_id as kelurahan_id, 
			COUNT(survey_respondents.id) as total_respondents, 
			COUNT(CASE WHEN survey_respondents.status_approval = 'validated_lurah' THEN 1 END) as total_validated
		`).
		Joins("JOIN respondents ON respondents.id = survey_respondents.respondent_id").
		Where("survey_respondents.survey_id = ?", surveyID).
		Where("respondents.kelurahan_id IN ?", kelurahanIDs).
		// 🆕 PENGAMANAN ROLE: Hanya hitung jika respondennya sah menjabat
		// Catatan: Jika survei ini diisi oleh RW, gunakan enums.ROLE_RW.
		// Jika diisi oleh RT, ganti menjadi enums.ROLE_RT.
		Where("respondents.role_id = ?", int64(enums.ROLE_RW))

	// 2. Jika survei bersifat Targeted (ada record di survey_wilayahs)
	// Pastikan hanya menghitung responden yang kecocokan wilayahnya terdaftar di survey_wilayahs
	if countWilayah > 0 {
		query = query.Where(`
			EXISTS (
				SELECT 1 FROM survey_wilayahs sw 
				WHERE sw.survey_id = survey_respondents.survey_id 
				AND sw.kecamatan_id = respondents.kecamatan_id
				AND (sw.kelurahan_id IS NULL OR sw.kelurahan_id = respondents.kelurahan_id)
				AND (sw.rw_id IS NULL OR sw.rw_id = respondents.rw_id)
			)
		`)
	}

	// Eksekusi akhir query dengan Group By
	err := query.Group("respondents.kelurahan_id").Find(&results).Error

	if err != nil {
		return nil, err
	}

	statusMap := make(map[int64]string)

	// 1. SET DEFAULT: Berikan nilai "Tidak ada responden" untuk semua Kelurahan di halaman ini
	for _, kelurahanID := range kelurahanIDs {
		statusMap[kelurahanID] = "Tidak ada responden"
	}

	// 2. TIMPA STATUS: Evaluasi berdasarkan perbandingan jumlah agregasi
	for _, res := range results {
		// Jika jumlah total responden SAMA DENGAN jumlah responden yang sudah divalidasi lurah
		if res.TotalRespondents == res.TotalValidated {
			statusMap[res.KelurahanId] = "Selesai"
		} else {
			// Jika ada selisih (berarti ada minimal 1 yang belum divalidasi lurah)
			statusMap[res.KelurahanId] = "Sedang Proses Verifikasi dan Validasi"
		}
	}

	return statusMap, nil
}

func (repository *surveyRepo) GetRespondentSurveyByRTId(ctx context.Context, surveyID int64, rtID int64) (*models.SurveyRespondent, error) {
	var respondentSurvey models.SurveyRespondent
	err := repository.dbSlave.Table("survey_respondents").
		Joins("JOIN respondents ON respondents.id = survey_respondents.respondent_id").
		Where("survey_respondents.survey_id = ?", surveyID).
		Where("respondents.rt_id = ?", rtID).
		Where("respondents.role_id = ?", int64(enums.ROLE_RT)).
		First(&respondentSurvey).Error

	if err != nil {
		return nil, err
	}

	return &respondentSurvey, nil
}

func (repository *surveyRepo) WipeAndReplaceAllFieldResponses(ctx context.Context, respondentID int64, newResponses []models.FieldResponse) error {
	defer utils.GeneralRecover()

	// Hapus total jawaban survei lama tanpa filter per section
	err := repository.dbMaster.WithContext(ctx).
		Where("form_response_id = ?", respondentID).
		Delete(&models.FieldResponse{}).Error
	if err != nil {
		return err
	}

	if len(newResponses) > 0 {
		return repository.dbMaster.WithContext(ctx).Create(&newResponses).Error
	}
	return nil
}

func (repository *surveyRepo) GetAllOldResponsesByRespondent(ctx context.Context, respondentID int64) ([]models.FieldResponse, error) {
	var responses []models.FieldResponse
	err := repository.dbSlave.WithContext(ctx).Where("form_response_id = ?", respondentID).Find(&responses).Error
	return responses, err
}

func (repository *surveyRepo) CreateFlaggingEdit(ctx context.Context, flags []models.FlaggingEditPertanyaanSurvey) error {
	if len(flags) == 0 {
		return nil
	}
	return repository.dbMaster.WithContext(ctx).Create(&flags).Error
}

func (repository *surveyRepo) ResetFlaggingEditStatus(ctx context.Context, respondentID int64) error {
	defer utils.GeneralRecover()

	// Mengubah is_revisied menjadi 'false' untuk semua flagging milik responden ini
	return repository.dbMaster.WithContext(ctx).
		Table("flagging_edit_pertanyaan_surveys").
		Where("survey_respondent_id = ?", respondentID).
		Update("is_revisied", "false").Error
}

func (repository *surveyRepo) UpdateRespondentApprovalStatus(ctx context.Context, respondentID int64, status int, statusApproval string, isEditPertanyaan string) error {
	return repository.dbMaster.WithContext(ctx).
		Table("survey_respondents").
		Where("id = ?", respondentID).
		Updates(map[string]interface{}{
			"status":             status,
			"status_approval":    statusApproval,
			"is_edit_pertanyaan": isEditPertanyaan,
		}).Error
}

func (repository *surveyRepo) ClearAndCreateFlaggingEdit(ctx context.Context, respondentID int64, flags []models.FlaggingEditPertanyaanSurvey) error {
	defer utils.GeneralRecover()

	// 1. Hapus seluruh flagging lama milik responden ini agar tidak duplikat
	err := repository.dbMaster.WithContext(ctx).
		Where("survey_respondent_id = ?", respondentID).
		Delete(&models.FlaggingEditPertanyaanSurvey{}).Error

	if err != nil {
		return err
	}

	// 2. Masukkan daftar pertanyaan baru yang harus direvisi (Bulk Insert)
	if len(flags) > 0 {
		err = repository.dbMaster.WithContext(ctx).Create(&flags).Error
		if err != nil {
			return err
		}
	}

	return nil
}

func (repository *surveyRepo) UpdateRespondentApprovalState(ctx context.Context, tx *gorm.DB, respondentID int64, status int, statusApproval *string, isEditPertanyaan string) error {
	defer utils.GeneralRecover()

	return tx.WithContext(ctx).
		Table("survey_respondents").
		Where("id = ?", respondentID).
		Updates(map[string]interface{}{
			"status":             status,
			"status_approval":    statusApproval,
			"is_edit_pertanyaan": isEditPertanyaan,
		}).Error
}

func (repository *surveyRepo) ResetFlaggingEditStatusTx(ctx context.Context, tx *gorm.DB, respondentID int64) error {
	defer utils.GeneralRecover()

	return tx.WithContext(ctx).
		Table("flagging_edit_pertanyaan_surveys").
		Where("survey_respondent_id = ?", respondentID).
		Update("is_revisied", "false").Error
}

func (repository *surveyRepo) GetSurveyCompletionHistory(ctx context.Context, surveyID int64, respondentID int64) (*models.LogSurveyDetail, error) {
	defer utils.GeneralRecover()

	var logDetail models.LogSurveyDetail
	err := repository.dbSlave.WithContext(ctx).
		Joins("JOIN respondents ON respondents.id = log__surveys.respondent_id").
		Where("log__surveys.survey_id = ? AND log__surveys.respondent_id = ?", surveyID, respondentID).
		Where("log__surveys.keterangan LIKE ? OR log__surveys.keterangan LIKE ?", "%telah menyelesaikan survey%", "%telah menyelesaikan revisi%").
		Order("log__surveys.created_at DESC").
		First(&logDetail).Error

	if err != nil {
		return nil, err
	}

	return &logDetail, nil
}

func (repository *surveyRepo) GetSurveyVerificationHistory(ctx context.Context, surveyID int64, respondentID int64) (*models.LogSurveyDetail, error) {
	defer utils.GeneralRecover()

	var logDetail models.LogSurveyDetail
	err := repository.dbSlave.WithContext(ctx).
		Joins("JOIN respondents ON respondents.id = log__surveys.respondent_id").
		Where("log__surveys.survey_id = ? AND log__surveys.respondent_id = ?", surveyID, respondentID).
		Where("log__surveys.keterangan LIKE ?", "%telah melakukan verifikasi%").
		Order("log__surveys.created_at DESC").
		First(&logDetail).Error

	if err != nil {
		return nil, err
	}

	return &logDetail, nil
}

func (repository *surveyRepo) GetSurveyValidationHistory(ctx context.Context, surveyID int64, respondentID int64) (*models.LogSurveyDetail, error) {
	defer utils.GeneralRecover()

	var logDetail models.LogSurveyDetail
	err := repository.dbSlave.WithContext(ctx).
		Joins("JOIN respondents ON respondents.id = log__surveys.respondent_id").
		Where("log__surveys.survey_id = ? AND log__surveys.respondent_id = ?", surveyID, respondentID).
		Where("log__surveys.keterangan LIKE ?", "%telah melakukan validasi%").
		Order("log__surveys.created_at DESC").
		First(&logDetail).Error

	if err != nil {
		return nil, err
	}

	return &logDetail, nil
}

func (repository *surveyRepo) GetRawJawabanForExport(ctx context.Context, surveyID int64, rtID int64) ([]dto.ExportRawJawabanDTO, error) {
	defer utils.GeneralRecover()
	var results []dto.ExportRawJawabanDTO

	query := `
		SELECT 
			r.id AS respondent_id,
			r.name AS nama_responden,
			sr.status,
			sr.status_approval,
			sr.updated_at AS waktu_selesai,
			fr.form_field_id,
			fr.answer
		FROM survey_respondents sr
		JOIN respondents r ON sr.respondent_id = r.id
		LEFT JOIN field_responses fr ON fr.form_response_id = sr.id
		WHERE sr.survey_id = ? AND r.rt_id = ?
		ORDER BY sr.created_at ASC
	`
	err := repository.dbSlave.WithContext(ctx).Raw(query, surveyID, rtID).Scan(&results).Error
	return results, err
}

func (repository *surveyRepo) ValidateIsImageExists(ctx context.Context, tx *gorm.DB, surveyRespondentId int64, filePath string) (bool, error) {
	defer utils.GeneralRecover()

	if tx == nil {
		tx = repository.dbSlave
	}

	var count int64
	err := tx.WithContext(ctx).
		Table("field_responses").
		Where("form_response_id = ? AND answer LIKE ?", surveyRespondentId, "%"+filePath+"%").
		Count(&count).Error

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func (repository *surveyRepo) GetRawJawabanWilayahForExport(ctx context.Context, surveyID int64, level int, wilayahID int64) ([]dto.ExportRawJawabanWilayah, error) {
	var results []dto.ExportRawJawabanWilayah

	// Gunakan .Debug() sementara agar Anda bisa melihat raw SQL-nya di terminal jika masih kosong
	query := repository.dbSlave.WithContext(ctx).Debug().Table("field_responses").
		Select(`
			survey_respondents.respondent_id,
			respondents.name as nama_responden,
			kecamatans.sub_district_name as kecamatan_name,
			kelurahans.village_name as kelurahan_name,
			rws.nama_rw as rw_name,
			rts.nama_rt as rt_name,
			rts.id as rt_id,
			survey_respondents.status,
			survey_respondents.status_approval,
			survey_respondents.updated_at as waktu_selesai,
			field_responses.form_field_id,
			field_responses.answer
		`).
		// 1. Relasi dari Jawaban -> Transaksi Survei Responden
		Joins("JOIN survey_respondents ON survey_respondents.id = field_responses.form_response_id").
		// 2. Relasi dari Transaksi Survei -> Akun Responden (RT)
		Joins("JOIN respondents ON respondents.id = survey_respondents.respondent_id").
		// 3. Tarik data wilayah milik Responden tersebut
		Joins("LEFT JOIN kecamatans ON kecamatans.id = respondents.kecamatan_id").
		Joins("LEFT JOIN kelurahans ON kelurahans.id = respondents.kelurahan_id").
		Joins("LEFT JOIN data__rws as rws ON rws.id = respondents.rw_id").
		Joins("LEFT JOIN data__rts as rts ON rts.id = respondents.rt_id").
		// 4. Filter bahwa ini adalah data untuk survei yang sedang di-export
		Where("survey_respondents.survey_id = ?", surveyID).
		Where("survey_respondents.status = ?", 2) // Pastikan tipe data status Anda benar integer/smallint 2

	// 5. Filter Kewilayahan: Karena semua responden adalah RT, kita tinggal mengecek
	// wilayah parent-nya berdasarkan level yang direquest.
	switch level {
	case 5: // KECAMATAN
		query = query.Where("respondents.kecamatan_id = ?", wilayahID)
	case 4: // KELURAHAN
		query = query.Where("respondents.kelurahan_id = ?", wilayahID)
	case 3: // RW
		query = query.Where("respondents.rw_id = ?", wilayahID)
	case 2: // RT
		query = query.Where("respondents.rt_id = ?", wilayahID)
	}

	err := query.Find(&results).Error
	return results, err
}

func (repository *surveyRepo) GetSurveyWilayahsBySurveyId(surveyId int64) ([]models.SurveyWilayah, error) {
	var wilayahs []models.SurveyWilayah
	err := repository.dbSlave.Where("survey_id = ?", surveyId).Find(&wilayahs).Error
	return wilayahs, err
}

func (repository *surveyRepo) GetSurveyorsBySurveyId(surveyId int64) ([]models.SurveySurveyor, error) {
	var surveyors []models.SurveySurveyor
	err := repository.dbSlave.Where("survey_id = ?", surveyId).Find(&surveyors).Error
	return surveyors, err
}
