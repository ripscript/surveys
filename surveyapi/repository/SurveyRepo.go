package repository

import (
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

	GetHistoryApproval(ctx context.Context, tx *gorm.DB, surveyID int64) ([]models.LogSurveyDetail, error)
	GetSurveyKewilayahan(userLogin models.JwtCustomClaims, respondentLogin *models.Respondent, req payloads.SurveyWilayahDatatablePayload) ([]models.SurveyKewilayahanDatatableResponse, int64, error)

	GetStatusKeterisianBulkRT(ctx context.Context, surveyID int64, rtIDs []int64) (map[int64]string, error)
	GetStatusKeterisianBulkRW(ctx context.Context, surveyID int64, rwIDs []int64) (map[int64]string, error)
	GetStatusKeterisianBulkKelurahan(ctx context.Context, surveyID int64, kelurahanIDs []int64) (map[int64]string, error)
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

	// 1. Ambil ID Responden untuk Query Isian (Hindari nil pointer)
	var respID int64 = 0
	if respondentLogin != nil {
		respID = respondentLogin.ID
	}

	// 2. Blok Select dengan 3 Subquery presisi tinggi berdasarkan skema baru
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

		-- Subquery 1: Menghitung Total Soal
		(
			SELECT COUNT(DISTINCT form_field_id) 
			FROM flow_fields 
			WHERE flow_detail_id = surveys.flow_detail_id
		) AS jumlah_total_soal,

		-- Subquery 2: Menghitung Soal Terisi oleh Responden ini
		(
			SELECT COUNT(DISTINCT fr.form_field_id) 
			FROM field_responses fr 
			INNER JOIN survey_respondents sr ON sr.id = fr.form_response_id 
			WHERE sr.survey_id = surveys.id AND sr.respondent_id = %d
		) AS jumlah_soal_terisi,

		-- Subquery 3: Tentukan Status Keterisian (Belum, Sedang Berjalan, Sudah Terisi)
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

		-- 🆕 Subquery 4 (FIXED): Mapping Status secara langsung ke String
		(
			SELECT CASE status 
				WHEN 1 THEN 'draft' 
				WHEN 2 THEN 'selesai' 
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

		// 3. Tambahkan 3 kolom baru ke whitelist order by (Agar datatable bisa di sorting)
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

	err := tx.WithContext(ctx).Create(&logSurvey).Error
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
	err := tx.WithContext(ctx).Where("survey_id = ?", surveyID).Order("created_at DESC").Find(&logs).Error
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

	return data, totalData, nil
}

func (repository *surveyRepo) GetStatusKeterisianBulkRT(ctx context.Context, surveyID int64, rtIDs []int64) (map[int64]string, error) {
	var results []struct {
		RtId           int64
		StatusApproval *string
	}

	err := repository.dbSlave.Table("survey_respondents").
		Select("respondents.rt_id, MAX(survey_respondents.status_approval) as status_approval").
		Joins("JOIN respondents ON respondents.id = survey_respondents.respondent_id").
		Where("survey_respondents.survey_id = ?", surveyID).
		Where("respondents.rt_id IN ?", rtIDs).
		Group("respondents.rt_id").
		Find(&results).Error

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

	err := repository.dbSlave.Table("survey_respondents").
		Select(`
			respondents.rw_id, 
			COUNT(survey_respondents.id) as total_respondents, 
			COUNT(CASE WHEN survey_respondents.status_approval = 'validated_lurah' THEN 1 END) as total_validated
		`).
		Joins("JOIN respondents ON respondents.id = survey_respondents.respondent_id").
		Where("survey_respondents.survey_id = ?", surveyID).
		Where("respondents.rw_id IN ?", rwIDs).
		Group("respondents.rw_id").
		Find(&results).Error

	if err != nil {
		return nil, err
	}

	statusMap := make(map[int64]string)

	// 1. SET DEFAULT: Berikan nilai "Tidak ada responden" untuk semua RW di halaman ini
	for _, rwID := range rwIDs {
		statusMap[rwID] = "Tidak ada responden"
	}

	// 2. TIMPA STATUS: Evaluasi berdasarkan perbandingan jumlah
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

	err := repository.dbSlave.Table("survey_respondents").
		Select(`
			respondents.kelurahan_id as kelurahan_id, 
			COUNT(survey_respondents.id) as total_respondents, 
			COUNT(CASE WHEN survey_respondents.status_approval = 'validated_lurah' THEN 1 END) as total_validated
		`).
		Joins("JOIN respondents ON respondents.id = survey_respondents.respondent_id").
		Where("survey_respondents.survey_id = ?", surveyID).
		Where("respondents.kelurahan_id IN ?", kelurahanIDs). // Diperbaiki dari KelurahanId menjadi kelurahanIDs
		Group("respondents.kelurahan_id").
		Find(&results).Error

	if err != nil {
		return nil, err
	}

	statusMap := make(map[int64]string)

	// 1. SET DEFAULT: Berikan nilai "Tidak ada responden" untuk semua Kelurahan di halaman ini
	for _, kelurahanID := range kelurahanIDs { // Diperbaiki dari rwIDs
		statusMap[kelurahanID] = "Tidak ada responden"
	}

	// 2. TIMPA STATUS: Evaluasi berdasarkan perbandingan jumlah
	for _, res := range results {
		// Jika jumlah total responden SAMA DENGAN jumlah responden yang sudah divalidasi lurah
		if res.TotalRespondents == res.TotalValidated {
			statusMap[res.KelurahanId] = "Selesai" // Diperbaiki menggunakan KelurahanId
		} else {
			// Jika ada selisih (berarti ada minimal 1 yang belum divalidasi lurah)
			statusMap[res.KelurahanId] = "Sedang Proses Verifikasi dan Validasi"
		}
	}

	return statusMap, nil
}
