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
	GetOptionsByMapping(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	Sync(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
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

func (service *dashboardOptionStatusMappingService) GetOptionsByMapping(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()

	idStr, ok := slug["id"].(string)
	if !ok {
		return utils.SendError(errors.New("Invalid ID in slug"), http.StatusBadRequest)
	}
	mappingID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return utils.SendError(errors.New("gagal parse string ke int64"), http.StatusBadRequest)
	}

	mapping, err := service.mappingRepo.FindByID(ctx, mappingID) // sudah Preload("DashboardMetric")
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

	formField, err := service.formFieldRepo.FindByID(ctx, mapping.FormFieldID)
	if err != nil {
		return utils.SendError(errors.New("Terjadi kesalahan pada server, silahkan coba lagi nanti"), http.StatusInternalServerError)
	}
	if formField == nil {
		return utils.SendError(errors.New("form field tidak ditemukan"), http.StatusNotFound)
	}

	answerOptions, err := service.formFieldRepo.ListAnswerOptions(ctx, mapping.FormFieldID)
	if err != nil {
		return utils.SendError(errors.New("Terjadi kesalahan pada server, silahkan coba lagi nanti"), http.StatusInternalServerError)
	}

	existing, err := service.statusRepo.ListByMappingID(ctx, mappingID)
	if err != nil {
		return utils.SendError(errors.New("Terjadi kesalahan pada server, silahkan coba lagi nanti"), http.StatusInternalServerError)
	}
	existingByOption := make(map[int64]models.DashboardOptionStatusMapping, len(existing))
	for _, e := range existing {
		existingByOption[e.AnswerOptionID] = e
	}

	options := make([]response.OptionStatusItem, 0, len(answerOptions))
	for _, opt := range answerOptions {
		item := response.OptionStatusItem{
			AnswerOptionID: int64(opt.ID),
			OptionLabel:    opt.Option,
		}
		if assigned, ok := existingByOption[int64(opt.ID)]; ok {
			id := assigned.ID
			key := assigned.StatusKey
			item.DashboardOptionStatusID = &id
			item.StatusKey = &key
		}
		options = append(options, item)
	}

	result := response.DashboardOptionStatusByMappingResponse{
		DashboardMetricMappingID: mappingID,
		FormFieldID:              mapping.FormFieldID,
		FormFieldQuestion:        formField.Question,
		Options:                  options,
	}

	return utils.SendData(result, "daftar opsi status berhasil diambil")
}

func (service *dashboardOptionStatusMappingService) Sync(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()

	var payload payloads.SyncDashboardOptionStatusMappingRequest
	if err := utils.DynamicBind(req, &payload); err != nil {
		return utils.SendError(err, http.StatusBadRequest)
	}

	var validate = validator.New()
	if err := validate.Struct(payload); err != nil {
		for _, err := range err.(validator.ValidationErrors) {
			return utils.SendError(errors.New(utils.TranslateError(err)), http.StatusBadRequest)
		}
	}

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

	// validasi: setiap answer_option_id memang milik form_field mapping ini, dan tidak duplikat dalam 1 request
	seen := make(map[int64]bool, len(payload.Assignments))
	for _, item := range payload.Assignments {
		if seen[item.AnswerOptionID] {
			return utils.SendError(errors.New("answer_option_id tidak boleh duplikat dalam satu request"), http.StatusBadRequest)
		}
		seen[item.AnswerOptionID] = true

		valid, err := service.formFieldRepo.AnswerOptionBelongsToField(ctx, item.AnswerOptionID, mapping.FormFieldID)
		if err != nil {
			return utils.SendError(errors.New("Terjadi kesalahan pada server, silahkan coba lagi nanti"), http.StatusInternalServerError)
		}
		if !valid {
			return utils.SendError(errors.New("salah satu answer_option_id tidak ditemukan pada form_field mapping ini"), http.StatusBadRequest)
		}
	}

	existing, err := service.statusRepo.ListByMappingID(ctx, payload.DashboardMetricMappingID)
	if err != nil {
		return utils.SendError(errors.New("Terjadi kesalahan pada server, silahkan coba lagi nanti"), http.StatusInternalServerError)
	}
	existingByOption := make(map[int64]models.DashboardOptionStatusMapping, len(existing))
	for _, e := range existing {
		existingByOption[e.AnswerOptionID] = e
	}

	var toInsert []models.DashboardOptionStatusMapping
	var toUpdate []models.DashboardOptionStatusMapping
	keepOptionIDs := make(map[int64]bool, len(payload.Assignments))

	for _, item := range payload.Assignments {
		keepOptionIDs[item.AnswerOptionID] = true
		if row, ok := existingByOption[item.AnswerOptionID]; ok {
			if row.StatusKey != item.StatusKey {
				row.StatusKey = item.StatusKey
				toUpdate = append(toUpdate, row)
			}
		} else {
			toInsert = append(toInsert, models.DashboardOptionStatusMapping{
				DashboardMetricMappingID: payload.DashboardMetricMappingID,
				AnswerOptionID:           item.AnswerOptionID,
				StatusKey:                item.StatusKey,
			})
		}
	}

	// opsi yang sebelumnya punya status tapi tidak lagi dikirim FE -> dianggap ingin dihapus
	var toDeleteIDs []int64
	for optionID, row := range existingByOption {
		if !keepOptionIDs[optionID] {
			toDeleteIDs = append(toDeleteIDs, row.ID)
		}
	}

	if len(toInsert) > 0 {
		if err := service.statusRepo.BulkCreate(ctx, toInsert); err != nil {
			return utils.SendError(errors.New("Terjadi kesalahan pada server, silahkan coba lagi nanti"), http.StatusInternalServerError)
		}
	}
	for i := range toUpdate {
		if err := service.statusRepo.Update(ctx, &toUpdate[i]); err != nil {
			return utils.SendError(errors.New("Terjadi kesalahan pada server, silahkan coba lagi nanti"), http.StatusInternalServerError)
		}
	}
	for _, id := range toDeleteIDs {
		if err := service.statusRepo.Delete(ctx, id); err != nil {
			return utils.SendError(errors.New("Terjadi kesalahan pada server, silahkan coba lagi nanti"), http.StatusInternalServerError)
		}
	}

	final, err := service.statusRepo.ListByMappingID(ctx, payload.DashboardMetricMappingID)
	if err != nil {
		return utils.SendError(errors.New("Terjadi kesalahan pada server, silahkan coba lagi nanti"), http.StatusInternalServerError)
	}
	result := make([]response.DashboardOptionStatusMappingResponse, 0, len(final))
	for _, m := range final {
		result = append(result, response.DashboardOptionStatusMappingResponse{
			ID:                       m.ID,
			DashboardMetricMappingID: m.DashboardMetricMappingID,
			AnswerOptionID:           m.AnswerOptionID,
			StatusKey:                m.StatusKey,
		})
	}

	return utils.SendData(result, "Berhasil memperbarui status mapping")
}
