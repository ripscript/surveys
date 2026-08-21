package service

import (
	"backend/masterapi/models"
	"backend/masterapi/payloads"
	"backend/masterapi/repository"
	"backend/masterapi/utils"
	"backend/siccore/pb"
	"context"
	"errors"
	"net/http"
	"net/url"

	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"
)

type DashboardMetricMappingService interface {
}

type dashboardMetricMappingService struct {
	metricRepo  repository.DashboardMetricRepository
	mappingRepo repository.DashboardMetricMappingRepository
}

func NewDashboardMetricMappingService(
	metricRepo repository.DashboardMetricRepository,
	mappingRepo repository.DashboardMetricMappingRepository,
) DashboardMetricMappingService {
	return &dashboardMetricMappingService{
		metricRepo:  metricRepo,
		mappingRepo: mappingRepo,
	}
}

func (s *dashboardMetricMappingService) Create(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
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

	metric, err := s.metricRepo.FindByID(ctx, payload.DashboardMetricID)
	if err != nil {
		if err.Error() != gorm.ErrRecordNotFound.Error() {
			return utils.SendError(errors.New("Terjadi kesalahan pada server, silahkan coba lagi nanti"), http.StatusInternalServerError)
		}
	}
	if metric == nil {
		return utils.SendError(errors.New("dashboard metric tidak ditemukan"), http.StatusNotFound)
	}

	// formField, err := s.formFieldRepo.FindByID(ctx, req.FormFieldID)
	// if err != nil {
	// 	return nil, err
	// }
	// if formField == nil {
	// 	return nil, utils.NewAppError(404, "form_field tidak ditemukan")
	// }

	// // validasi #1: form_field harus milik form yang disebut, dan template-nya harus cocok
	// // sama expected_template si metric -- ini yang mencegah "number di-assign ke chart pie"
	// if formField.FormID != req.FormID {
	// 	return nil, utils.NewAppError(400, "form_field_id tidak termasuk dalam form_id yang disebutkan")
	// }
	// if formField.Template != metric.ExpectedTemplate {
	// 	return nil, utils.NewAppError(400,
	// 		fmt.Sprintf("tipe pertanyaan (%s) tidak cocok dengan expected_template metric (%s)",
	// 			formField.Template, metric.ExpectedTemplate))
	// }

	// // validasi #2: kalau metric bertipe multiple-choices, answer_option_id WAJIB diisi
	// // dan harus milik form_field yang sama (bukan opsi dari pertanyaan lain)
	// if metric.ExpectedTemplate == "multiple-choices" {
	// 	if req.AnswerOptionID == nil {
	// 		return nil, utils.NewAppError(400, "answer_option_id wajib diisi untuk metric bertipe multiple-choices")
	// 	}
	// 	validOption, err := s.formFieldRepo.AnswerOptionBelongsToField(ctx, *req.AnswerOptionID, req.FormFieldID)
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// 	if !validOption {
	// 		return nil, utils.NewAppError(400, "answer_option_id tidak ditemukan pada form_field yang dipilih")
	// 	}
	// } else if req.AnswerOptionID != nil {
	// 	return nil, utils.NewAppError(400, "answer_option_id hanya boleh diisi untuk metric bertipe multiple-choices")
	// }

	// // validasi #3: cegah 1 metric punya 2 mapping untuk form yang sama (selaras UNIQUE constraint di DB,
	// // tapi dicek dulu di service supaya errornya jelas, bukan raw DB constraint error)
	// existing, err := s.mappingRepo.FindByMetricAndForm(ctx, req.DashboardMetricID, req.FormID)
	// if err != nil {
	// 	return nil, err
	// }
	// if existing != nil {
	// 	return nil, utils.NewAppError(400, "metric ini sudah punya mapping untuk form_id tersebut, gunakan update")
	// }

	// mapping := &models.DashboardMetricMapping{
	// 	DashboardMetricID: req.DashboardMetricID,
	// 	FormID:            req.FormID,
	// 	FormFieldID:       req.FormFieldID,
	// 	AnswerOptionID:    req.AnswerOptionID,
	// }
	// if err := s.mappingRepo.Create(ctx, mapping); err != nil {
	// 	return nil, err
	// }

	// return toDashboardMetricMappingResponse(mapping), nil

	return utils.SendData(nil, "test")
}
