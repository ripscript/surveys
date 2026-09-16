package service

import (
	"backend/masterapi/customValidator"
	"backend/masterapi/models"
	"backend/masterapi/payloads"
	"backend/masterapi/repository"
	"backend/masterapi/response"
	"backend/masterapi/utils"
	"backend/siccore/pb"
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"
)

type DashboardMetricService interface {
	GetList(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	Create(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	Detail(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	Update(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	Delete(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	GetCategoryOptions(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
}

type dashboardMetricService struct {
	repo        repository.DashboardMetricRepository
	mappingRepo repository.DashboardMetricMappingRepository
}

func NewDashboardMetricService(repo repository.DashboardMetricRepository, mappingRepo repository.DashboardMetricMappingRepository) DashboardMetricService {
	return &dashboardMetricService{repo: repo, mappingRepo: mappingRepo}
}

func (service *dashboardMetricService) GetList(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()

	var payload payloads.DatatablePayload
	err := utils.DynamicBind(req, &payload)
	if err != nil {
		return utils.SendError(err, http.StatusBadRequest)
	}

	if payload.Page <= 0 {
		payload.Page = 1
	}
	if payload.Limit <= 0 {
		payload.Limit = 25
	}

	data, totalData, err := service.repo.GetList(payload)
	if err != nil {
		return utils.SendError(err, http.StatusInternalServerError)
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

	return utils.SendData(result, "berhasil mengambil daftar metrik")
}

func (service *dashboardMetricService) Create(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()
	var payload payloads.CreateDashboardMetricRequest
	err := utils.DynamicBind(req, &payload)
	if err != nil {
		return utils.SendError(err, http.StatusBadRequest)
	}

	var validate = validator.New()
	validate.RegisterStructValidation(customValidator.CreateDashboardMetricValidator, payloads.CreateDashboardMetricRequest{})
	err = validate.Struct(payload)
	if err != nil {
		for _, err := range err.(validator.ValidationErrors) {
			customErrorMsg := utils.TranslateError(err)
			return utils.SendError(errors.New(customErrorMsg), http.StatusBadRequest)
		}
	}

	existing, err := service.repo.FindByMetricKey(ctx, payload.MetricKey)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return utils.SendError(fmt.Errorf("metric_key '%s' sudah dipakai", payload.MetricKey), http.StatusBadRequest)
	}

	m := &models.DashboardMetric{
		MetricKey:        payload.MetricKey,
		Label:            payload.Label,
		ExpectedTemplate: payload.ExpectedTemplate,
		Category:         payload.Category,
		IsDashboard:      utils.BoolToPointer(false),
	}
	if err := service.repo.Create(ctx, m); err != nil {
		return nil, err
	}
	return utils.SendData(toDashboardMetricResponse(m), "Berhasil menambahkan metrik")
}

func (service *dashboardMetricService) Detail(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()

	idStr, ok := slug["id"].(string)
	if !ok {
		return utils.SendError(errors.New("Invalid ID in slug"), http.StatusBadRequest)
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return utils.SendError(errors.New("gagal parse string ke int64"), http.StatusBadRequest)
	}

	metric, err := service.repo.FindByID(ctx, id)
	if err != nil {
		if err.Error() != gorm.ErrRecordNotFound.Error() {
			return utils.SendError(errors.New("Terjadi kesalahan pada server, silahkan coba lagi nanti"), http.StatusInternalServerError)
		}
	}
	if metric == nil {
		return utils.SendError(errors.New("metrik tidak ditemukan"), http.StatusNotFound)
	}

	return utils.SendData(toDashboardMetricResponse(metric), "data metrik berhasil diambil")
}

func (s *dashboardMetricService) Update(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()
	var payload payloads.UpdateDashboardMetricRequest
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

	m, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if err.Error() != gorm.ErrRecordNotFound.Error() {
			return utils.SendError(errors.New("Terjadi kesalahan pada server, silahkan coba lagi nanti"), http.StatusInternalServerError)
		}
	}
	if m == nil {
		return utils.SendError(errors.New("metrik tidak ditemukan"), http.StatusNotFound)
	}

	if m.IsDashboard != nil && *m.IsDashboard == true {
		return utils.SendError(errors.New("metrik ini tidak dapat diubah, karena digunakan di dashboard"), http.StatusNotFound)
	}

	m.Label = payload.Label
	m.ExpectedTemplate = payload.ExpectedTemplate
	m.Category = payload.Category

	if err := s.repo.Update(ctx, m); err != nil {
		return nil, err
	}
	return utils.SendData(toDashboardMetricResponse(m), "Berhasil memperbarui metrik")
}

func (s *dashboardMetricService) Delete(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()
	idStr, ok := slug["id"].(string)
	if !ok {
		return utils.SendError(errors.New("Invalid ID in slug"), http.StatusBadRequest)
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return utils.SendError(errors.New("gagal parse string ke int64"), http.StatusBadRequest)
	}

	m, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if err.Error() != gorm.ErrRecordNotFound.Error() {
			return utils.SendError(errors.New("Terjadi kesalahan pada server, silahkan coba lagi nanti"), http.StatusInternalServerError)
		}
	}
	if m == nil {
		return utils.SendError(errors.New("metrik tidak ditemukan"), http.StatusNotFound)
	}

	if m.IsDashboard != nil && *m.IsDashboard == true {
		return utils.SendError(errors.New("metrik ini tidak dapat dihapus, karena digunakan di dashboard"), http.StatusNotFound)
	}

	// cek dulu apakah metric ini masih punya mapping ke form manapun,
	// karena FK dashboard_metric_mappings -> dashboard_metrics akan menolak delete kalau masih ada
	mappingCount, err := s.mappingRepo.CountByMetricID(ctx, id)
	if err != nil {
		return utils.SendError(errors.New("Terjadi kesalahan pada server, silahkan coba lagi nanti"), http.StatusInternalServerError)
	}
	if mappingCount > 0 {
		return utils.SendError(errors.New("metrik ini tidak dapat dihapus, karena masih memiliki pemetaan ke form. Hapus pemetaannya terlebih dahulu"), http.StatusBadRequest)
	}

	err = s.repo.Delete(ctx, id)
	if err != nil {
		return utils.SendError(errors.New("Terjadi kesalahan pada server, silahkan coba lagi nanti"), http.StatusInternalServerError)
	}

	return utils.SendData(nil, "berhasil menghapus metric")
}

func (service *dashboardMetricService) GetCategoryOptions(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()

	page, err := strconv.Atoi(param.Get("page"))
	if err != nil || page <= 0 {
		page = 1
	}

	limit, err := strconv.Atoi(param.Get("limit"))
	if err != nil || limit <= 0 {
		limit = 1000
	}

	var rawValues []string
	if len(param["value[]"]) > 0 {
		rawValues = param["value[]"]
	} else if len(param["value"]) > 0 {
		rawValues = param["value"]
	}

	var parsedValues []string
	for _, rawValue := range rawValues {
		parsedValues = append(parsedValues, rawValue)
	}

	payload := payloads.DashboardMetricCategoryOptionsPayload{
		Q:          param.Get("q"),
		Page:       page,
		Limit:      limit,
		Categories: parsedValues,
	}

	categories, totalData, err := service.repo.GetCategoryOptions(ctx, payload.Q, payload.Categories, payload.Page, payload.Limit)
	if err != nil {
		return utils.SendError(errors.New("Terjadi kesalahan pada server, silahkan coba lagi nanti"), http.StatusInternalServerError)
	}

	options := make([]response.StringOptionItem, 0, len(categories))
	for _, c := range categories {
		options = append(options, response.StringOptionItem{
			Value: c,
			Label: utils.ToCamelCase(c),
		})
	}

	hasMore := int64(payload.Page*payload.Limit) < totalData

	result := response.StringOptionsResponse{
		Options: options,
		Meta: response.PaginationMeta{
			CurrentPage: payload.Page,
			PerPage:     payload.Limit,
			Total:       totalData,
			HasMore:     hasMore,
		},
	}

	return utils.SendData(result, "")
}

func toDashboardMetricResponse(m *models.DashboardMetric) *response.DashboardMetricResponse {
	return &response.DashboardMetricResponse{
		ID:               m.ID,
		MetricKey:        m.MetricKey,
		Label:            m.Label,
		ExpectedTemplate: m.ExpectedTemplate,
		Category:         m.Category,
		CreatedAt:        m.CreatedAt,
		UpdatedAt:        m.UpdatedAt,
		IsDashboard:      m.IsDashboard,
	}
}
