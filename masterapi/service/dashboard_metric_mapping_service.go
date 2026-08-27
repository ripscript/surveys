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
}

type dashboardMetricMappingService struct {
	metricRepo    repository.DashboardMetricRepository
	mappingRepo   repository.DashboardMetricMappingRepository
	formRepo      repository.FormRepository
	formFieldRepo repository.FormFieldRepository
}

func NewDashboardMetricMappingService(
	metricRepo repository.DashboardMetricRepository,
	mappingRepo repository.DashboardMetricMappingRepository,
	formRepo repository.FormRepository,
	formFieldRepo repository.FormFieldRepository,
) DashboardMetricMappingService {
	return &dashboardMetricMappingService{
		metricRepo:    metricRepo,
		mappingRepo:   mappingRepo,
		formRepo:      formRepo,
		formFieldRepo: formFieldRepo,
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

	// validasi #1: form_field harus milik form yang disebut, dan template-nya harus cocok
	// sama expected_template si metric -- ini yang mencegah "number di-assign ke chart pie"
	if formField.FormId != form.ID {
		return utils.SendError(errors.New("form_field_id tidak termasuk dalam form_id yang disebutkan"), http.StatusBadRequest)
	}
	if formField.Template != metric.ExpectedTemplate {
		return utils.SendError(errors.New("tipe pertanyaan tidak cocok dengan expected_template metric"), http.StatusBadRequest)
	}

	// validasi #2: kalau metric bertipe multiple-choices, answer_option_id WAJIB diisi
	// dan harus milik form_field yang sama (bukan opsi dari pertanyaan lain)
	if metric.ExpectedTemplate == "multiple-choices" {
		if payload.AnswerOptionID == nil {
			return utils.SendError(errors.New("answer_option_id wajib diisi untuk metric bertipe multiple-choices"), http.StatusBadRequest)
		}
		validOption, err := service.formFieldRepo.AnswerOptionBelongsToField(ctx, *payload.AnswerOptionID, payload.FormFieldID)
		if err != nil {
			return nil, err
		}
		if !validOption {
			return utils.SendError(errors.New("answer_option_id tidak ditemukan pada form_field yang dipilih"), http.StatusBadRequest)
		}
	} else if payload.AnswerOptionID != nil {
		return utils.SendError(errors.New("answer_option_id hanya boleh diisi untuk metric bertipe multiple-choices"), http.StatusBadRequest)
	}

	// validasi #3: cegah 1 metric punya 2 mapping untuk form yang sama (selaras UNIQUE constraint di DB,
	// tapi dicek dulu di service supaya errornya jelas, bukan raw DB constraint error)
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
		AnswerOptionID:    payload.AnswerOptionID,
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
		return utils.SendError(errors.New("dashboard metric mapping tidak ditemukan"), http.StatusNotFound)
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
		return utils.SendError(errors.New("dashboard metric mapping tidak ditemukan"), http.StatusNotFound)
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

	if mapping.DashboardMetric.ExpectedTemplate == "multiple-choices" {
		if payload.AnswerOptionID == nil {
			return utils.SendError(errors.New("answer_option_id wajib diisi untuk metric bertipe multiple-choices"), http.StatusBadRequest)
		}
		validOption, err := service.formFieldRepo.AnswerOptionBelongsToField(ctx, *payload.AnswerOptionID, payload.FormFieldID)
		if err != nil {
			return utils.SendError(err, http.StatusInternalServerError)
		}
		if !validOption {
			return utils.SendError(errors.New("answer_option_id tidak ditemukan pada form_field yang dipilih"), http.StatusBadRequest)
		}
	}

	mapping.FormFieldID = payload.FormFieldID
	mapping.AnswerOptionID = payload.AnswerOptionID

	if err := service.mappingRepo.Update(ctx, mapping); err != nil {
		return utils.SendError(err, http.StatusInternalServerError)
	}

	return utils.SendData(toDashboardMetricMappingResponse(mapping), "Berhasil memperbarui dashboard metric mapping")
}

func toDashboardMetricMappingResponse(m *models.DashboardMetricMapping) *response.DashboardMetricMappingResponse {
	return &response.DashboardMetricMappingResponse{
		ID:                m.ID,
		DashboardMetricID: m.DashboardMetricID,
		FormID:            m.FormID,
		FormFieldID:       m.FormFieldID,
		AnswerOptionID:    m.AnswerOptionID,
	}
}
