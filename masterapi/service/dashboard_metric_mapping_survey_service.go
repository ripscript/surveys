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
	"os"

	"github.com/go-playground/validator/v10"
	"github.com/speps/go-hashids/v2"
	"gorm.io/gorm"
)

type DashboardMetricMappingSurveyService interface {
	GetMappingSurvey(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	SaveMappingSurvey(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
}

type dashboardMetricMappingSurveyService struct {
	surveyRepo  repository.SurveyRepo
	mappingRepo repository.DashboardMetricMappingRepository
}

func NewDashboardMetricMappingSurveyService(
	mappingRepo repository.DashboardMetricMappingRepository,
	surveyRepo repository.SurveyRepo,
) DashboardMetricMappingSurveyService {
	return &dashboardMetricMappingSurveyService{mappingRepo: mappingRepo, surveyRepo: surveyRepo}
}

func (s *dashboardMetricMappingSurveyService) GetMappingSurvey(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()

	surveyCode, ok := slug["survey_code"].(string)
	if !ok || surveyCode == "" {
		return utils.SendError(errors.New("Kode alur survey tidak valid"), http.StatusBadRequest)
	}

	hd := hashids.NewData()
	hd.Salt = os.Getenv("HASHID_SALT")
	hd.MinLength = 24

	h, err := hashids.NewWithData(hd)
	if err != nil {
		return utils.SendError(err, http.StatusInternalServerError)
	}

	surveyId, err := h.DecodeInt64WithError(surveyCode)
	if err != nil || len(surveyId) == 0 {
		return utils.SendError(errors.New("Kode alur survey tidak valid"), http.StatusBadRequest)
	}

	survey, err := s.surveyRepo.GetSurveyByID(int(surveyId[0]))
	if err != nil {
		if err.Error() != gorm.ErrRecordNotFound.Error() {
			return utils.SendError(errors.New("Terjadi kesalahan pada server, silahkan coba lagi nanti"), http.StatusInternalServerError)
		}
	}

	if survey == nil {
		return utils.SendError(errors.New("Survey tidak ditemukan"), http.StatusNotFound)
	}

	flow, err := s.mappingRepo.GetFlowDetailById(ctx, survey.FlowDetailID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return utils.SendError(errors.New("survey tidak ditemukan"), http.StatusNotFound)
		}
		return utils.SendError(errors.New("Terjadi kesalahan pada server, silahkan coba lagi nanti"), http.StatusInternalServerError)
	}

	rawRows, err := s.mappingRepo.GetSurveyQuestionsForMapping(ctx, flow.ID)
	if err != nil {
		return utils.SendError(errors.New("Terjadi kesalahan pada server, silahkan coba lagi nanti"), http.StatusInternalServerError)
	}

	// PART 1: assemble question map dulu (gabungkan multi-row opsi jadi 1 question)
	questionOrder := make([]int64, 0)
	questionMap := make(map[int64]*response.MappingSurveyQuestionItem)
	questionSectionID := make(map[int64]int)       // form_field_id -> section_id (0 kalau null)
	questionSectionName := make(map[int64]*string) // form_field_id -> section_name

	for _, row := range rawRows {
		q, exists := questionMap[row.FormFieldID]
		if !exists {
			q = &response.MappingSurveyQuestionItem{
				FormFieldID:       row.FormFieldID,
				Question:          row.Question,
				Description:       row.Description,
				Template:          row.Template,
				Sequence:          row.Sequence,
				DashboardMetricID: row.DashboardMetricID,
				MetricKey:         row.MetricKey,
				Options:           []response.MappingSurveyOptionItem{},
			}
			questionMap[row.FormFieldID] = q
			questionOrder = append(questionOrder, row.FormFieldID)

			sectionID := 0
			if row.SectionID != nil {
				sectionID = int(*row.SectionID)
			}
			questionSectionID[row.FormFieldID] = sectionID
			questionSectionName[row.FormFieldID] = row.SectionName
		}

		if row.AnswerOptionID != nil {
			label := ""
			if row.AnswerLabel != nil {
				label = *row.AnswerLabel
			}
			q.Options = append(q.Options, response.MappingSurveyOptionItem{
				AnswerOptionID:          *row.AnswerOptionID,
				AnswerLabel:             label,
				DashboardMetricStatusID: row.StatusID,
			})
		}
	}

	// PART 2: group ke section, tapi urutan section & urutan question DI
	// DALAM section tetap ikut urutan flow_fields.sequence (query sudah
	// ORDER BY sequence ASC), bukan diacak oleh map.
	sectionOrder := make([]int, 0)
	sectionGroups := make(map[int]*response.MappingSurveySectionGroup)

	for _, fieldID := range questionOrder {
		sectionID := questionSectionID[fieldID]
		group, exists := sectionGroups[sectionID]
		if !exists {
			group = &response.MappingSurveySectionGroup{
				SectionID:   sectionID,
				SectionName: questionSectionName[fieldID],
				Questions:   []response.MappingSurveyQuestionItem{},
			}
			sectionGroups[sectionID] = group
			sectionOrder = append(sectionOrder, sectionID)
		}
		group.Questions = append(group.Questions, *questionMap[fieldID])
	}

	sections := make([]response.MappingSurveySectionGroup, 0, len(sectionOrder))
	for _, sid := range sectionOrder {
		sections = append(sections, *sectionGroups[sid])
	}

	result := response.MappingSurveyResponse{
		SurveyID:   int64(survey.ID),
		SurveyCode: surveyCode,
		SurveyName: survey.Name,
		FlowCode:   flow.Code,
		FlowName:   flow.Name,
		FormID:     flow.FormID,
		FormCode:   flow.FormCode,
		Sections:   sections,
	}

	return utils.SendData(result, "Berhasil mengambil data pemetaan survey")
}

func (s *dashboardMetricMappingSurveyService) SaveMappingSurvey(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()

	var payload payloads.SaveMappingSurveyRequest
	if err := utils.DynamicBind(req, &payload); err != nil {
		return utils.SendError(err, http.StatusBadRequest)
	}

	validate := validator.New()
	if err := validate.Struct(payload); err != nil {
		for _, verr := range err.(validator.ValidationErrors) {
			return utils.SendError(errors.New(utils.TranslateError(verr)), http.StatusBadRequest)
		}
	}

	if len(payload.Questions) == 0 {
		return utils.SendError(errors.New("minimal 1 pertanyaan harus dipetakan"), http.StatusBadRequest)
	}

	mappedQuestions := make([]payloads.SaveMappingSurveyQuestionPayload, 0, len(payload.Questions))
	for _, q := range payload.Questions {
		if q.DashboardMetricID == nil || *q.DashboardMetricID <= 0 {
			continue
		}

		validOptions := make([]payloads.SaveMappingSurveyOptionPayload, 0, len(q.Options))
		for _, opt := range q.Options {
			if opt.DashboardMetricStatusID == nil || *opt.DashboardMetricStatusID <= 0 {
				continue
			}
			validOptions = append(validOptions, opt)
		}
		q.Options = validOptions

		mappedQuestions = append(mappedQuestions, q)
	}

	if len(mappedQuestions) == 0 {
		return utils.SendData(nil, "Tidak ada pertanyaan yang dipetakan ke metric")
	}

	surveyCode, ok := slug["survey_code"].(string)
	if !ok || surveyCode == "" {
		return utils.SendError(errors.New("Kode alur survey tidak valid"), http.StatusBadRequest)
	}

	hd := hashids.NewData()
	hd.Salt = os.Getenv("HASHID_SALT")
	hd.MinLength = 24

	h, err := hashids.NewWithData(hd)
	if err != nil {
		return utils.SendError(err, http.StatusInternalServerError)
	}

	surveyId, err := h.DecodeInt64WithError(surveyCode)
	if err != nil || len(surveyId) == 0 {
		return utils.SendError(errors.New("Kode alur survey tidak valid"), http.StatusBadRequest)
	}

	survey, err := s.surveyRepo.GetSurveyByID(int(surveyId[0]))
	if err != nil {
		if err.Error() != gorm.ErrRecordNotFound.Error() {
			return utils.SendError(errors.New("Terjadi kesalahan pada server, silahkan coba lagi nanti"), http.StatusInternalServerError)
		}
	}

	if survey == nil {
		return utils.SendError(errors.New("Survey tidak ditemukan"), http.StatusNotFound)
	}

	flow, err := s.mappingRepo.GetFlowDetailById(ctx, survey.FlowDetailID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return utils.SendError(errors.New("survey tidak ditemukan"), http.StatusNotFound)
		}
		return utils.SendError(errors.New("Terjadi kesalahan pada server, silahkan coba lagi nanti"), http.StatusInternalServerError)
	}

	validRows, err := s.mappingRepo.GetSurveyQuestionsForMapping(ctx, flow.ID)
	if err != nil {
		return utils.SendError(errors.New("Terjadi kesalahan pada server, silahkan coba lagi nanti"), http.StatusInternalServerError)
	}
	validFieldIDs := make(map[int64]bool)
	validOptionIDs := make(map[int64]bool)
	for _, r := range validRows {
		validFieldIDs[r.FormFieldID] = true
		if r.AnswerOptionID != nil {
			validOptionIDs[*r.AnswerOptionID] = true
		}
	}

	for _, q := range mappedQuestions {
		if !validFieldIDs[q.FormFieldID] {
			return utils.SendError(errors.New("pertanyaan tidak ditemukan pada alur ini"), http.StatusBadRequest)
		}

		for _, opt := range q.Options {
			if !validOptionIDs[opt.AnswerOptionID] {
				return utils.SendError(errors.New("opsi jawaban tidak ditemukan pada alur ini"), http.StatusBadRequest)
			}
		}
	}

	if err := s.mappingRepo.SaveSurveyMapping(ctx, flow.FormID, mappedQuestions); err != nil {
		return utils.SendError(errors.New("Terjadi kesalahan pada server, silahkan coba lagi nanti"), http.StatusInternalServerError)
	}

	return utils.SendData(nil, "Berhasil menyimpan pemetaan metrik")
}
