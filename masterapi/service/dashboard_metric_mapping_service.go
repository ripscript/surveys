package service

import (
	"backend/masterapi/models"
	"backend/masterapi/payloads"
	"backend/masterapi/repository"
	"backend/masterapi/response"
	"backend/masterapi/utils"
	"backend/siccore/pb"
	"context"
	"errors"
	"net/http"
	"net/url"
	"strconv"

	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"
)

type DashboardMetricMappingService interface {
	Create(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	Delete(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	Update(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)

	GetList(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	Detail(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
}

type dashboardMetricMappingService struct {
	metricRepo    repository.DashboardMetricRepository
	mappingRepo   repository.DashboardMetricMappingRepository
	formRepo      repository.FormRepository
	formFieldRepo repository.FormFieldRepository
	statusRepo    repository.DashboardOptionStatusMappingRepository
}

func NewDashboardMetricMappingService(
	metricRepo repository.DashboardMetricRepository,
	mappingRepo repository.DashboardMetricMappingRepository,
	formRepo repository.FormRepository,
	formFieldRepo repository.FormFieldRepository,
	statusRepo repository.DashboardOptionStatusMappingRepository,
) DashboardMetricMappingService {
	return &dashboardMetricMappingService{
		metricRepo:    metricRepo,
		mappingRepo:   mappingRepo,
		formRepo:      formRepo,
		formFieldRepo: formFieldRepo,
		statusRepo:    statusRepo,
	}
}

func (service *dashboardMetricMappingService) Create(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()
	var payload payloads.CreateDashboardMetricMappingRequest
	err := utils.DynamicBind(req, &payload)
	if err != nil {
		return utils.SendError(err, http.StatusBadRequest)
	}

	var validate = validator.New()
	err = validate.Struct(payload)
	if err != nil {
		for _, err := range err.(validator.ValidationErrors) {
			customErrorMsg := utils.TranslateError(err)
			return utils.SendError(errors.New(customErrorMsg), http.StatusBadRequest)
		}
	}

	form, err := service.formRepo.FindByCode(ctx, payload.FormCode)
	if err != nil {
		if err.Error() != gorm.ErrRecordNotFound.Error() {
			return utils.SendError(errors.New("Terjadi kesalahan pada server, silahkan coba lagi nanti"), http.StatusInternalServerError)
		}
	}
	if form == nil {
		return utils.SendError(errors.New("form tidak ditemukan"), http.StatusNotFound)
	}

	metric, err := service.metricRepo.FindByID(ctx, payload.DashboardMetricID)
	if err != nil {
		if err.Error() != gorm.ErrRecordNotFound.Error() {
			return utils.SendError(errors.New("Terjadi kesalahan pada server, silahkan coba lagi nanti"), http.StatusInternalServerError)
		}
	}
	if metric == nil {
		return utils.SendError(errors.New("dashboard metric tidak ditemukan"), http.StatusNotFound)
	}

	formField, err := service.formFieldRepo.FindByID(ctx, payload.FormFieldID)
	if err != nil {
		return nil, err
	}
	if formField == nil {
		return utils.SendError(errors.New("form field tidak ditemukan"), http.StatusNotFound)
	}

	if formField.FormId != form.ID {
		return utils.SendError(errors.New("form_field_id tidak termasuk dalam form_id yang disebutkan"), http.StatusBadRequest)
	}
	if formField.Template != metric.ExpectedTemplate {
		return utils.SendError(errors.New("tipe pertanyaan tidak cocok dengan expected_template metric"), http.StatusBadRequest)
	}

	existing, err := service.mappingRepo.FindByMetricAndForm(ctx, payload.DashboardMetricID, int64(form.ID))
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return utils.SendError(errors.New("metric ini sudah punya mapping untuk form_id tersebut, gunakan update"), http.StatusBadRequest)
	}

	mapping := &models.DashboardMetricMapping{
		DashboardMetricID: payload.DashboardMetricID,
		FormID:            int64(form.ID),
		FormFieldID:       payload.FormFieldID,
	}
	if err := service.mappingRepo.Create(ctx, mapping); err != nil {
		return nil, err
	}

	return utils.SendData(toDashboardMetricMappingResponse(mapping), "")
}

func (service *dashboardMetricMappingService) Delete(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()

	idStr, ok := slug["id"].(string)
	if !ok {
		return utils.SendError(errors.New("Invalid ID in slug"), http.StatusBadRequest)
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return utils.SendError(errors.New("gagal parse string ke int64"), http.StatusBadRequest)
	}

	mapping, err := service.mappingRepo.FindByID(ctx, id)
	if err != nil {
		if err.Error() != gorm.ErrRecordNotFound.Error() {
			return utils.SendError(errors.New("Terjadi kesalahan pada server, silahkan coba lagi nanti"), http.StatusInternalServerError)
		}
	}
	if mapping == nil {
		return utils.SendError(errors.New("pemetaan metrik tidak ditemukan"), http.StatusNotFound)
	}

	// hapus dulu status mapping anaknya sebelum hapus parent-nya
	if err := service.statusRepo.DeleteByMappingID(ctx, id); err != nil {
		return utils.SendError(errors.New("Terjadi kesalahan pada server, silahkan coba lagi nanti"), http.StatusInternalServerError)
	}

	err = service.mappingRepo.Delete(ctx, id)
	if err != nil {
		return utils.SendError(errors.New("Terjadi kesalahan pada server, silahkan coba lagi nanti"), http.StatusInternalServerError)
	}
	return utils.SendData(nil, "Dashboard metric mapping berhasil dihapus")
}

func (service *dashboardMetricMappingService) Update(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()
	var payload payloads.UpdateDashboardMetricMappingRequest
	err := utils.DynamicBind(req, &payload)
	if err != nil {
		return utils.SendError(err, http.StatusBadRequest)
	}

	var validate = validator.New()
	err = validate.Struct(payload)
	if err != nil {
		for _, err := range err.(validator.ValidationErrors) {
			customErrorMsg := utils.TranslateError(err)
			return utils.SendError(errors.New(customErrorMsg), http.StatusBadRequest)
		}
	}

	idStr, ok := slug["id"].(string)
	if !ok {
		return utils.SendError(errors.New("Invalid ID in slug"), http.StatusBadRequest)
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return utils.SendError(errors.New("gagal parse string ke int64"), http.StatusBadRequest)
	}

	mapping, err := service.mappingRepo.FindByID(ctx, id)
	if err != nil {
		if err.Error() != gorm.ErrRecordNotFound.Error() {
			return utils.SendError(errors.New("Terjadi kesalahan pada server, silahkan coba lagi nanti"), http.StatusInternalServerError)
		}
	}
	if mapping == nil {
		return utils.SendError(errors.New("pemetaan metrik tidak ditemukan"), http.StatusNotFound)
	}

	formField, err := service.formFieldRepo.FindByID(ctx, payload.FormFieldID)
	if err != nil {
		return utils.SendError(err, http.StatusInternalServerError)
	}
	if formField == nil {
		return utils.SendError(errors.New("form_field tidak ditemukan"), http.StatusNotFound)
	}
	if int64(formField.FormId) != mapping.FormID {
		return utils.SendError(errors.New("form_field_id tidak termasuk dalam form_id mapping ini"), http.StatusBadRequest)
	}
	if formField.Template != mapping.DashboardMetric.ExpectedTemplate {
		return utils.SendError(errors.New("tipe pertanyaan tidak cocok dengan expected_template metric"), http.StatusBadRequest)
	}

	// cek apakah form_field_id benar-benar berubah SEBELUM di-assign ke mapping
	formFieldChanged := mapping.FormFieldID != payload.FormFieldID

	mapping.FormFieldID = payload.FormFieldID

	if err := service.mappingRepo.Update(ctx, mapping); err != nil {
		return utils.SendError(err, http.StatusInternalServerError)
	}

	// form_field_id berubah -> semua answer_option_id di status mapping lama sudah tidak valid,
	// karena opsi jawaban terikat pada form_field tertentu. Hapus supaya tidak ada data nyasar.
	if formFieldChanged {
		if err := service.statusRepo.DeleteByMappingID(ctx, mapping.ID); err != nil {
			return utils.SendError(errors.New("Terjadi kesalahan pada server, silahkan coba lagi nanti"), http.StatusInternalServerError)
		}
	}

	return utils.SendData(toDashboardMetricMappingResponse(mapping), "Berhasil memperbarui pemetaan metrik")
}

func (service *dashboardMetricMappingService) GetList(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()

	var payload payloads.DatatablePayload
	if err := utils.DynamicBind(req, &payload); err != nil {
		return utils.SendError(err, http.StatusBadRequest)
	}
	if payload.Page <= 0 {
		payload.Page = 1
	}
	if payload.Limit <= 0 {
		payload.Limit = 25
	}

	var formCode string
	if v := param.Get("form_code"); v != "" {
		formCode = v
	}

	var dashboardMetricID int64
	if v := param.Get("dashboard_metric_id"); v != "" {
		dashboardMetricID, _ = strconv.ParseInt(v, 10, 64)
	}

	rows, totalData, err := service.mappingRepo.GetList(ctx, payload, formCode, dashboardMetricID)
	if err != nil {
		return utils.SendError(errors.New("Terjadi kesalahan pada server, silahkan coba lagi nanti"), http.StatusInternalServerError)
	}

	data := make([]response.DashboardMetricMappingListItem, 0, len(rows))
	for _, row := range rows {
		data = append(data, response.DashboardMetricMappingListItem{
			ID:                row.ID,
			DashboardMetricID: row.DashboardMetricID,
			MetricKey:         row.MetricKey,
			MetricLabel:       row.MetricLabel,
			ExpectedTemplate:  row.ExpectedTemplate,
			Category:          row.Category,
			FormID:            row.FormID,
			FormTitle:         row.FormTitle,
			FormCode:          row.FormCode,
			FormFieldID:       row.FormFieldID,
			FormFieldQuestion: row.FormFieldQuestion,
			CreatedAt:         row.CreatedAt,
			UpdatedAt:         row.UpdatedAt,
		})
	}

	showingFrom := (payload.Page-1)*payload.Limit + 1
	showingTo := showingFrom + len(data) - 1
	if totalData == 0 {
		showingFrom = 0
		showingTo = 0
	}

	result := map[string]interface{}{
		"data": data,
		"meta": map[string]interface{}{
			"total_entries": totalData,
			"current_page":  payload.Page,
			"per_page":      payload.Limit,
			"showing_from":  showingFrom,
			"showing_to":    showingTo,
		},
	}

	return utils.SendData(result, "berhasil mengambil daftar pemetaan metrik")
}

func (service *dashboardMetricMappingService) Detail(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()

	idStr, ok := slug["id"].(string)
	if !ok {
		return utils.SendError(errors.New("Invalid ID in slug"), http.StatusBadRequest)
	}
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return utils.SendError(errors.New("gagal parse string ke int64"), http.StatusBadRequest)
	}

	mapping, err := service.mappingRepo.FindDetailByID(ctx, id)
	if err != nil {
		if err.Error() != gorm.ErrRecordNotFound.Error() {
			return utils.SendError(errors.New("Terjadi kesalahan pada server, silahkan coba lagi nanti"), http.StatusInternalServerError)
		}
	}
	if mapping == nil {
		return utils.SendError(errors.New("pemetaan metrik tidak ditemukan"), http.StatusNotFound)
	}

	optionStatusCount, err := service.mappingRepo.CountOptionStatusByMappingID(ctx, id)
	if err != nil {
		return utils.SendError(errors.New("Terjadi kesalahan pada server, silahkan coba lagi nanti"), http.StatusInternalServerError)
	}

	detail := response.DashboardMetricMappingDetailResponse{
		ID: mapping.ID,
		DashboardMetric: response.DashboardMetricSummary{
			ID:               mapping.DashboardMetric.ID,
			MetricKey:        mapping.DashboardMetric.MetricKey,
			Label:            mapping.DashboardMetric.Label,
			ExpectedTemplate: mapping.DashboardMetric.ExpectedTemplate,
			Category:         mapping.DashboardMetric.Category,
		},
		Form: response.FormSummary{
			ID:    mapping.FormID,
			Title: mapping.Form.Title,
			Code:  mapping.Form.Code,
		},
		FormField: response.FormFieldSummary{
			ID:       mapping.FormFieldID,
			Question: mapping.FormField.Question,
			Template: mapping.FormField.Template,
		},
		OptionStatusCount: int(optionStatusCount),
		CreatedAt:         mapping.CreatedAt,
		UpdatedAt:         mapping.UpdatedAt,
	}

	return utils.SendData(detail, "data pemetaan metrik berhasil diambil")
}

func toDashboardMetricMappingResponse(m *models.DashboardMetricMapping) *response.DashboardMetricMappingResponse {
	return &response.DashboardMetricMappingResponse{
		ID:                m.ID,
		DashboardMetricID: m.DashboardMetricID,
		FormID:            m.FormID,
		FormFieldID:       m.FormFieldID,
	}
}
