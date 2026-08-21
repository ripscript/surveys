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
	Create(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	Detail(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	Update(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	Delete(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	// List(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
}

type dashboardMetricService struct {
	repo repository.DashboardMetricRepository
}

func NewDashboardMetricService(repo repository.DashboardMetricRepository) DashboardMetricService {
	return &dashboardMetricService{repo: repo}
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
	}
	if err := service.repo.Create(ctx, m); err != nil {
		return nil, err
	}
	return utils.SendData(toDashboardMetricResponse(m), "Berhasil menambahkan dashboard metric")
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
		return utils.SendError(errors.New("dashboard metric tidak ditemukan"), http.StatusNotFound)
	}

	return utils.SendData(toDashboardMetricResponse(metric), "asdasd")
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
		return utils.SendError(errors.New("dashboard metric tidak ditemukan"), http.StatusNotFound)
	}

	m.Label = payload.Label
	m.ExpectedTemplate = payload.ExpectedTemplate
	m.Category = payload.Category

	if err := s.repo.Update(ctx, m); err != nil {
		return nil, err
	}
	return utils.SendData(toDashboardMetricResponse(m), "Berhasil memperbarui dashboard metric")
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
		return utils.SendError(errors.New("dashboard metric tidak ditemukan"), http.StatusNotFound)
	}

	err = s.repo.Delete(ctx, id)
	if err != nil {
		return utils.SendError(errors.New("Terjadi kesalahan pada server, silahkan coba lagi nanti"), http.StatusInternalServerError)
	}

	return utils.SendData(nil, "Berhasil menghapus dashboard metric")
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
	}
}
