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

type DashboardOptionStatusMappingService interface {
	BulkAssign(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	Update(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	Delete(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	ListByMapping(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
}

type dashboardOptionStatusMappingService struct {
	statusRepo    repository.DashboardOptionStatusMappingRepository
	mappingRepo   repository.DashboardMetricMappingRepository
	formFieldRepo repository.FormFieldRepository
}

func NewDashboardOptionStatusMappingService(
	statusRepo repository.DashboardOptionStatusMappingRepository,
	mappingRepo repository.DashboardMetricMappingRepository,
	formFieldRepo repository.FormFieldRepository,
) DashboardOptionStatusMappingService {
	return &dashboardOptionStatusMappingService{
		statusRepo:    statusRepo,
		mappingRepo:   mappingRepo,
		formFieldRepo: formFieldRepo,
	}
}

func (service *dashboardOptionStatusMappingService) BulkAssign(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()
	var payload payloads.BulkAssignDashboardOptionStatusMappingRequest
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

	// validasi #1: mapping harus ada dan expected_template metric-nya wajib multiple-choices
	// (status badge cuma masuk akal untuk jawaban kategorikal, bukan number/long-answer/dst)
	mapping, err := service.mappingRepo.FindByID(ctx, payload.DashboardMetricMappingID)
	if err != nil {
		if err.Error() != gorm.ErrRecordNotFound.Error() {
			return utils.SendError(errors.New("Terjadi kesalahan pada server, silahkan coba lagi nanti"), http.StatusInternalServerError)
		}
	}
	if mapping == nil {
		return utils.SendError(errors.New("dashboard metric mapping tidak ditemukan"), http.StatusNotFound)
	}
	if mapping.DashboardMetric == nil || mapping.DashboardMetric.ExpectedTemplate != "multiple-choices" {
		return utils.SendError(errors.New("status mapping hanya berlaku untuk metric bertipe multiple-choices"), http.StatusBadRequest)
	}

	// validasi #2: setiap answer_option_id yang dikirim harus benar-benar
	// milik form_field yang sama dengan mapping ini (cegah opsi nyasar dari pertanyaan lain)
	toInsert := make([]models.DashboardOptionStatusMapping, 0, len(payload.Assignments))
	for _, item := range payload.Assignments {
		valid, err := service.formFieldRepo.AnswerOptionBelongsToField(ctx, item.AnswerOptionID, mapping.FormFieldID)
		if err != nil {
			return utils.SendError(errors.New("Terjadi kesalahan pada server, silahkan coba lagi nanti"), http.StatusInternalServerError)
		}
		if !valid {
			return utils.SendError(errors.New("salah satu answer_option_id tidak ditemukan pada form_field mapping ini"), http.StatusBadRequest)
		}

		// validasi #3: cegah duplikat assignment untuk opsi yang sama (selaras UNIQUE constraint)
		existing, err := service.statusRepo.FindByMappingAndOption(ctx, payload.DashboardMetricMappingID, item.AnswerOptionID)
		if err != nil {
			return utils.SendError(errors.New("Terjadi kesalahan pada server, silahkan coba lagi nanti"), http.StatusInternalServerError)
		}
		if existing != nil {
			return utils.SendError(errors.New("answer_option_id sudah punya status mapping, gunakan update"), http.StatusBadRequest)
		}

		toInsert = append(toInsert, models.DashboardOptionStatusMapping{
			DashboardMetricMappingID: payload.DashboardMetricMappingID,
			AnswerOptionID:           item.AnswerOptionID,
			StatusKey:                item.StatusKey,
		})
	}

	if err := service.statusRepo.BulkCreate(ctx, toInsert); err != nil {
		return utils.SendError(errors.New("Terjadi kesalahan pada server, silahkan coba lagi nanti"), http.StatusInternalServerError)
	}

	result := make([]response.DashboardOptionStatusMappingResponse, 0, len(toInsert))
	for _, m := range toInsert {
		result = append(result, response.DashboardOptionStatusMappingResponse{
			DashboardMetricMappingID: m.DashboardMetricMappingID,
			AnswerOptionID:           m.AnswerOptionID,
			StatusKey:                m.StatusKey,
		})
	}

	return utils.SendData(result, "Berhasil menyimpan status mapping")
}

func (service *dashboardOptionStatusMappingService) Update(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()
	var payload payloads.UpdateDashboardOptionStatusMappingRequest
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

	m, err := service.statusRepo.FindByID(ctx, id)
	if err != nil {
		if err.Error() != gorm.ErrRecordNotFound.Error() {
			return utils.SendError(errors.New("Terjadi kesalahan pada server, silahkan coba lagi nanti"), http.StatusInternalServerError)
		}
	}
	if m == nil {
		return utils.SendError(errors.New("status mapping tidak ditemukan"), http.StatusNotFound)
	}

	m.StatusKey = payload.StatusKey
	if err := service.statusRepo.Update(ctx, m); err != nil {
		return utils.SendError(errors.New("Terjadi kesalahan pada server, silahkan coba lagi nanti"), http.StatusInternalServerError)
	}

	return utils.SendData(response.DashboardOptionStatusMappingResponse{
		ID:                       m.ID,
		DashboardMetricMappingID: m.DashboardMetricMappingID,
		AnswerOptionID:           m.AnswerOptionID,
		StatusKey:                m.StatusKey,
	}, "Berhasil memperbarui status mapping")
}

func (service *dashboardOptionStatusMappingService) Delete(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()
	idStr, ok := slug["id"].(string)
	if !ok {
		return utils.SendError(errors.New("Invalid ID in slug"), http.StatusBadRequest)
	}
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return utils.SendError(errors.New("gagal parse string ke int64"), http.StatusBadRequest)
	}

	m, err := service.statusRepo.FindByID(ctx, id)
	if err != nil {
		if err.Error() != gorm.ErrRecordNotFound.Error() {
			return utils.SendError(errors.New("Terjadi kesalahan pada server, silahkan coba lagi nanti"), http.StatusInternalServerError)
		}
	}
	if m == nil {
		return utils.SendError(errors.New("status mapping tidak ditemukan"), http.StatusNotFound)
	}

	if err := service.statusRepo.Delete(ctx, id); err != nil {
		return utils.SendError(errors.New("Terjadi kesalahan pada server, silahkan coba lagi nanti"), http.StatusInternalServerError)
	}
	return utils.SendData(nil, "Status mapping berhasil dihapus")
}

func (service *dashboardOptionStatusMappingService) ListByMapping(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()

	mappingIDStr := param.Get("dashboard_metric_mapping_id")
	mappingID, err := strconv.ParseInt(mappingIDStr, 10, 64)
	if err != nil {
		return utils.SendError(errors.New("dashboard_metric_mapping_id wajib diisi dan berupa angka"), http.StatusBadRequest)
	}

	list, err := service.statusRepo.ListByMappingID(ctx, mappingID)
	if err != nil {
		return utils.SendError(errors.New("Terjadi kesalahan pada server, silahkan coba lagi nanti"), http.StatusInternalServerError)
	}

	result := make([]response.DashboardOptionStatusMappingResponse, 0, len(list))
	for _, m := range list {
		result = append(result, response.DashboardOptionStatusMappingResponse{
			ID:                       m.ID,
			DashboardMetricMappingID: m.DashboardMetricMappingID,
			AnswerOptionID:           m.AnswerOptionID,
			StatusKey:                m.StatusKey,
		})
	}

	return utils.SendData(result, "")
}
