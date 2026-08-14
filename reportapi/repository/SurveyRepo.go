package repository

import (
	"backend/reportapi/enums"
	"backend/reportapi/models"
	"backend/reportapi/payloads"
	"backend/reportapi/response"
	"backend/reportapi/utils"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"strconv"

	"gorm.io/gorm"
)

type SurveyRepo interface {
	GetFormFieldTemplateByID(formFieldID int64) (string, error)
	LogSurveys(offset int, limit int, param url.Values) ([]models.LogSurveys, int64, error)
	CheckRespondent(respondentId int) (models.Respondents, error)
	InsertLogSurvey(data models.Log_Surveys) error
	GetRespondentById(ctx context.Context, id int64) (*models.Respondentss, error)
	GetSurveyById(surveyId int64) (*models.Survey, error)
	GetDaftarKelurahan(ctx context.Context, kecamatan_id int64, payload payloads.DatatablePayload) (response.KelurahanDatatableResponse, error)
	GetStatusKeterisianBulkKelurahan(ctx context.Context, surveyID int64, kelurahanIDs []int64) (map[int64]string, error)
	GetDaftarRW(ctx context.Context, kelurahan_id int64, payload payloads.DatatablePayload) (response.RWDatatableResponse, error)
	GetStatusKeterisianBulkRW(ctx context.Context, surveyID int64, rwIDs []int64) (map[int64]string, error)
	GetDaftarRT(ctx context.Context, rw_id int64, payload payloads.DatatablePayload) (response.RTDatatableResponse, error)
	GetStatusKeterisianBulkRT(ctx context.Context, surveyID int64, rtIDs []int64) (map[int64]string, error)

	GetRespondentByID(respondentID int64) (*models.Respondentsss, error)
	GetKecamatanByID(id uint) (*models.Kecamatan, error)
	GetKelurahanByID(id uint) (*models.Kelurahan, error)
	GetRwByKelurahanID(kelurahanID uint) (*models.DataRw, error)
	GetSurveysForRT(kecamatanID, kelurahanID, rwID uint, status *string) ([]models.Surveyss, error)
	GetSurveysForSurveyor(respondentID int64, status *string) ([]models.Surveyss, error)
	GetSurveysForOther(respondent *models.Respondentsss, kecamatanID, kelurahanID, rwID uint, status *string, offset int, limit int, param url.Values) ([]models.Surveyss, int64, error)
	GetFlowDetailByID(id uint) (*models.FlowDetails, error)
	GetFlowFieldsByFlowDetailID(flowDetailID uint) ([]models.FlowFields, error)
	GetBreakdownFlowFields(flowDetailID uint, sectionID *uint) ([]models.FlowFields, error)
	GetFirstFlowField(flowDetailID uint, sectionID *uint, breakdownOnly bool) (*models.FlowFields, error)
	GetLastFlowField(flowDetailID uint, sectionID *uint) (*models.FlowFields, error)
	GetFlowFieldsBetween(firstID, lastID, flowDetailID uint, sectionID *uint) ([]models.FlowFields, error)
	CountFormFields(formFieldIDs []uint, excludeIDs []uint) (int64, error)
}

func NewSurveyRepo(dbSlave, dbMaster *gorm.DB) *surveyRepo {
	defer utils.GeneralRecover()
	return &surveyRepo{dbSlave, dbMaster}
}

type surveyRepo struct {
	dbSlave  *gorm.DB
	dbMaster *gorm.DB
}

func (repository *surveyRepo) GetFormFieldTemplateByID(formFieldID int64) (string, error) {
	defer utils.GeneralRecover()

	var template string
	err := repository.dbSlave.Table("form_fields").
		Select("template").
		Where("id = ?", formFieldID).
		Take(&template).Error

	return template, err
}

func (r *surveyRepo) LogSurveys(offset int, limit int, param url.Values) ([]models.LogSurveys, int64, error) {
	defer utils.GeneralRecover()
	var data []models.LogSurveys
	var total int64
	db := r.dbSlave

	query := db.Model(&models.LogSurveys{}).Where("log_level = ?", "6")

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

func (r *surveyRepo) CheckRespondent(respondentId int) (models.Respondents, error) {
	defer utils.GeneralRecover()
	var user models.Respondents
	db := r.dbSlave

	err := db.Where("id = ?", respondentId).First(&user).Error
	if err != nil {
		return user, err
	}

	return user, nil
}

func (r *surveyRepo) InsertLogSurvey(data models.Log_Surveys) error {
	defer utils.GeneralRecover()
	db := r.dbMaster
	err := db.Create(data).Error
	if err != nil {
		return err
	}

	return nil
}

func (repository *surveyRepo) GetRespondentById(ctx context.Context, id int64) (*models.Respondentss, error) {
	host := os.Getenv("USERAPI_HOST") + ":" + os.Getenv("USERAPI_PORT")

	newSlug := map[string]interface{}{"id": strconv.FormatInt(id, 10)}

	dataBytes, err := utils.HitBackend(ctx, host, "GET", "/respondent/raw/:id", newSlug, nil)
	if err != nil {
		fmt.Println(err.Error())
		return nil, errors.New("Gagal mendapatkan data surveyor dari UserAPI")
	}

	var data models.Respondentss
	if err := json.Unmarshal(dataBytes, &data); err != nil {
		return nil, errors.New("Gagal memparsing data surveyor dari UserAPI")
	}
	return &data, nil
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

func (repository *surveyRepo) GetDaftarKelurahan(ctx context.Context, kecamatan_id int64, payload payloads.DatatablePayload) (response.KelurahanDatatableResponse, error) {
	host := os.Getenv("MASTERAPI_HOST") + ":" + os.Getenv("MASTERAPI_PORT")

	newSlug := map[string]interface{}{
		"kecamatan_id": kecamatan_id,
		"search":       payload.Search,
		"page":         strconv.Itoa(payload.Page),
		"limit":        strconv.Itoa(payload.Limit),
		"order_by":     payload.OrderBy,
		"order_dir":    payload.OrderDir,
	}

	dataBytes, err := utils.HitBackend(ctx, host, "GET", "/manajemen-wilayah/kelurahan/list/:kecamatan_id", newSlug, nil)
	if err != nil {
		return response.KelurahanDatatableResponse{}, errors.New("Gagal mendapatkan daftar Kelurahan dari MasterAPI")
	}

	var data response.KelurahanDatatableResponse
	if err := json.Unmarshal(dataBytes, &data); err != nil {
		return response.KelurahanDatatableResponse{}, errors.New("Gagal memparsing data daftar Kelurahan dari MasterAPI")
	}

	return data, nil
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

func (repository *surveyRepo) GetDaftarRW(ctx context.Context, kelurahan_id int64, payload payloads.DatatablePayload) (response.RWDatatableResponse, error) {
	host := os.Getenv("MASTERAPI_HOST") + ":" + os.Getenv("MASTERAPI_PORT")

	newSlug := map[string]interface{}{
		"kelurahan_id": kelurahan_id,
		"search":       payload.Search,
		"page":         strconv.Itoa(payload.Page),
		"limit":        strconv.Itoa(payload.Limit),
		"order_by":     payload.OrderBy,
		"order_dir":    payload.OrderDir,
	}

	dataBytes, err := utils.HitBackend(ctx, host, "GET", "/manajemen-wilayah/rw/list/:kelurahan_id", newSlug, nil)
	if err != nil {
		return response.RWDatatableResponse{}, errors.New("Gagal mendapatkan daftar RW dari MasterAPI")
	}

	var data response.RWDatatableResponse
	if err := json.Unmarshal(dataBytes, &data); err != nil {
		return response.RWDatatableResponse{}, errors.New("Gagal memparsing data daftar RW dari MasterAPI")
	}

	return data, nil
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

func (repository *surveyRepo) GetDaftarRT(ctx context.Context, rw_id int64, payload payloads.DatatablePayload) (response.RTDatatableResponse, error) {
	host := os.Getenv("MASTERAPI_HOST") + ":" + os.Getenv("MASTERAPI_PORT")

	newSlug := map[string]interface{}{
		"rw_id":     rw_id,
		"search":    payload.Search,
		"page":      strconv.Itoa(payload.Page),
		"limit":     strconv.Itoa(payload.Limit),
		"order_by":  payload.OrderBy,
		"order_dir": payload.OrderDir,
	}

	dataBytes, err := utils.HitBackend(ctx, host, "GET", "/manajemen-wilayah/rt/list/:rw_id", newSlug, nil)
	if err != nil {
		return response.RTDatatableResponse{}, errors.New("Gagal mendapatkan daftar RT dari MasterAPI")
	}

	var data response.RTDatatableResponse
	if err := json.Unmarshal(dataBytes, &data); err != nil {
		return response.RTDatatableResponse{}, errors.New("Gagal memparsing data daftar RT dari MasterAPI")
	}

	return data, nil
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

func (r *surveyRepo) GetRespondentByID(respondentID int64) (*models.Respondentsss, error) {
	var respondent models.Respondentsss
	err := r.dbSlave.
		Preload("Role").
		First(&respondent, respondentID).Error
	return &respondent, err
}

func (r *surveyRepo) GetKecamatanByID(id uint) (*models.Kecamatan, error) {
	var kecamatan models.Kecamatan
	err := r.dbSlave.First(&kecamatan, id).Error
	return &kecamatan, err
}

func (r *surveyRepo) GetKelurahanByID(id uint) (*models.Kelurahan, error) {
	var kelurahan models.Kelurahan
	err := r.dbSlave.First(&kelurahan, id).Error
	return &kelurahan, err
}

func (r *surveyRepo) GetRwByKelurahanID(kelurahanID uint) (*models.DataRw, error) {
	var rw models.DataRw
	err := r.dbSlave.Where("kelurahan_id = ?", kelurahanID).First(&rw).Error
	return &rw, err
}

func (r *surveyRepo) wilayahFilter(kecamatanID, kelurahanID, rwID uint) *gorm.DB {
	wilayahSubQuery := r.dbSlave.Table("survey_wilayahs").
		Select("1").
		Where("survey_id = surveys.id").
		Where(
			r.dbSlave.
				Where("tingkat_wilayah = ? AND kecamatan_id = ?", "5", kecamatanID).
				Or("tingkat_wilayah = ? AND kecamatan_id = ? AND kelurahan_id = ?", "4", kecamatanID, kelurahanID).
				Or("tingkat_wilayah = ? AND kecamatan_id = ? AND kelurahan_id = ? AND rw_id = ?", "3", kecamatanID, kelurahanID, rwID),
		)

	return r.dbSlave.Where(
		r.dbSlave.
			Where("EXISTS (?)", wilayahSubQuery).
			Or("NOT EXISTS (SELECT 1 FROM survey_wilayahs WHERE survey_id = surveys.id)"),
	)
}
func (r *surveyRepo) GetSurveysForRT(kecamatanID, kelurahanID, rwID uint, status *string) ([]models.Surveyss, error) {
	var surveys []models.Surveyss

	query := r.wilayahFilter(kecamatanID, kelurahanID, rwID).
		Select("id, name, flow_detail_id, start_date, end_date, type, created_at, updated_at, status").
		Preload("FlowDetail").
		Preload("SurveyRespondents.FieldResponses")

	if status != nil {
		query = query.Where("status = ?", *status)
	}

	err := query.Order("created_at DESC").Find(&surveys).Error
	return surveys, err
}

func (r *surveyRepo) GetSurveysForSurveyor(respondentID int64, status *string) ([]models.Surveyss, error) {
	var surveys []models.Surveyss

	surveyorSubQuery := r.dbSlave.Table("survey_surveyors").
		Select("1").
		Where("survey_id = surveys.id AND respondent_id = ?", respondentID)

	query := r.dbSlave.Where("EXISTS (?)", surveyorSubQuery)

	if status != nil {
		query = query.Where("status = ?", *status)
	}

	err := query.Order("created_at DESC").Find(&surveys).Error
	return surveys, err
}

func (r *surveyRepo) GetSurveysForOther(respondent *models.Respondentsss, kecamatanID, kelurahanID, rwID uint, status *string, offset int, limit int, param url.Values) ([]models.Surveyss, int64, error) {
	var surveys []models.Surveyss
	var total int64
	respondentSubQuery := r.dbSlave.Table("survey_respondents sr").
		Joins("JOIN respondents res ON res.id = sr.respondent_id").
		Select("1").
		Where("sr.survey_id = surveys.id")

	switch respondent.Role.Name {
	case "rw":
		respondentSubQuery = respondentSubQuery.Where(
			"res.kecamatan_id = ? AND res.kelurahan_id = ? AND res.rw_id = ?",
			respondent.KecamatanID, respondent.KelurahanID, respondent.RwID,
		)
	case "lurah":
		respondentSubQuery = respondentSubQuery.Where(
			"res.kecamatan_id = ? AND res.kelurahan_id = ?",
			respondent.KecamatanID, respondent.KelurahanID,
		)
	case "camat":
		respondentSubQuery = respondentSubQuery.Where(
			"res.kecamatan_id = ?",
			respondent.KecamatanID,
		)
	}

	query := r.wilayahFilter(kecamatanID, kelurahanID, rwID).
		Where("EXISTS (?)", respondentSubQuery)

	if status != nil {
		query = query.Where("status = ?", *status)
	}

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	err := query.Order("created_at DESC").Find(&surveys).Error

	total = int64(len(surveys))

	return surveys, total, err
}

func (r *surveyRepo) GetFlowDetailByID(id uint) (*models.FlowDetails, error) {
	var flowDetail models.FlowDetails
	err := r.dbSlave.First(&flowDetail, id).Error
	return &flowDetail, err
}

func (r *surveyRepo) GetFlowFieldsByFlowDetailID(flowDetailID uint) ([]models.FlowFields, error) {
	var flowFields []models.FlowFields
	err := r.dbSlave.
		Select("form_field_id").
		Where("flow_detail_id = ?", flowDetailID).
		Find(&flowFields).Error
	return flowFields, err
}

func (r *surveyRepo) GetBreakdownFlowFields(flowDetailID uint, sectionID *uint) ([]models.FlowFields, error) {
	var flowFields []models.FlowFields
	query := r.dbSlave.Where("flow_detail_id = ? AND breakdown = ?", flowDetailID, true)
	if sectionID != nil {
		query = query.Where("section_id = ?", *sectionID)
	}
	err := query.Find(&flowFields).Error
	return flowFields, err
}

func (r *surveyRepo) GetFirstFlowField(flowDetailID uint, sectionID *uint, breakdownOnly bool) (*models.FlowFields, error) {
	var flowField models.FlowFields
	query := r.dbSlave.Select("id").Where("flow_detail_id = ?", flowDetailID)
	if sectionID != nil {
		query = query.Where("section_id = ?", *sectionID)
	}
	if breakdownOnly {
		query = query.Where("breakdown = ?", true)
	}
	err := query.First(&flowField).Error
	return &flowField, err
}

func (r *surveyRepo) GetLastFlowField(flowDetailID uint, sectionID *uint) (*models.FlowFields, error) {
	var flowField models.FlowFields
	query := r.dbSlave.Select("id").Where("flow_detail_id = ?", flowDetailID)
	if sectionID != nil {
		query = query.Where("section_id = ?", *sectionID)
	}
	err := query.Last(&flowField).Error
	return &flowField, err
}

func (r *surveyRepo) GetFlowFieldsBetween(firstID, lastID, flowDetailID uint, sectionID *uint) ([]models.FlowFields, error) {
	var flowFields []models.FlowFields
	query := r.dbSlave.Select("form_field_id").
		Where("id BETWEEN ? AND ?", firstID, lastID).
		Where("flow_detail_id = ?", flowDetailID)
	if sectionID != nil {
		query = query.Where("section_id = ?", *sectionID)
	}
	err := query.Find(&flowFields).Error
	return flowFields, err
}

func (r *surveyRepo) CountFormFields(formFieldIDs []uint, excludeIDs []uint) (int64, error) {
	var count int64
	query := r.dbSlave.Model(&models.FormField{}).Where("id IN ?", formFieldIDs)
	if len(excludeIDs) > 0 {
		query = query.Where("id NOT IN ?", excludeIDs)
	}
	err := query.Count(&count).Error
	return count, err
}
