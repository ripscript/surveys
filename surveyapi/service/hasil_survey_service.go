package service

import (
	"backend/siccore/pb"
	"backend/surveyapi/enums"
	"backend/surveyapi/models"
	"backend/surveyapi/payloads"
	"backend/surveyapi/repository"
	"backend/surveyapi/response"
	"backend/surveyapi/utils"
	"context"
	"encoding/json"
	"errors"
	"math"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/speps/go-hashids/v2"
)

type HasilSurveyService interface {
	GetListSurvey(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	SurveyResultSummary(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)

	SurveyResultQuestionDetail(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
}

type hasilSurveyService struct {
	manajemenAlurRepo              repository.ManajemenAlurRepo
	templateFormulirPertanyaanRepo repository.TemplateFormulirPertanyaanRepo
	templateUcapanRepo             repository.TemplateUcapanRepo
	surveyRepo                     repository.SurveyRepo
	wilayahRepo                    repository.WilayahRepo
	userRepo                       repository.UserRepo
	fileRepo                       repository.FileRepo

	isSyncing atomic.Bool
}

func NewHasilSurveyService(
	manajemenAlurRepo repository.ManajemenAlurRepo,
	templateFormulirPertanyaanRepo repository.TemplateFormulirPertanyaanRepo,
	templateUcapanRepo repository.TemplateUcapanRepo,
	surveyRepo repository.SurveyRepo,
	wilayahRepo repository.WilayahRepo,
	userRepo repository.UserRepo,
	fileRepo repository.FileRepo,
) HasilSurveyService {
	return &hasilSurveyService{
		manajemenAlurRepo:              manajemenAlurRepo,
		templateFormulirPertanyaanRepo: templateFormulirPertanyaanRepo,
		templateUcapanRepo:             templateUcapanRepo,
		surveyRepo:                     surveyRepo,
		wilayahRepo:                    wilayahRepo,
		userRepo:                       userRepo,
		fileRepo:                       fileRepo,
	}
}

func (service *hasilSurveyService) GetListSurvey(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()

	if usr.Role != int(enums.ROLE_ADMIN) && usr.Role != int(enums.ROLE_SURVEYOR) {
		return utils.SendError(errors.New("Anda tidak memiliki hak akses untuk melihat daftar survey ini"), http.StatusForbidden)
	}

	search := param.Get("search")
	page, _ := strconv.Atoi(param.Get("page"))
	limit, _ := strconv.Atoi(param.Get("limit"))
	orderBy := param.Get("order_by")
	orderDir := param.Get("order_dir")

	status := param.Get("status")

	payload := payloads.HasilSurveyDatatablePayload{
		Search:   search,
		Page:     page,
		Limit:    limit,
		OrderBy:  orderBy,
		OrderDir: orderDir,
		Status:   status,
	}

	var validate = validator.New()
	err := validate.Struct(payload)
	if err != nil {
		for _, err := range err.(validator.ValidationErrors) {
			customErrorMsg := utils.TranslateError(err)
			return utils.SendError(errors.New(customErrorMsg), http.StatusBadRequest)
		}
	}

	if payload.Page <= 0 {
		payload.Page = 1
	}
	if payload.Limit <= 0 {
		payload.Limit = 25
	}

	respondentLogin, err := service.userRepo.GetRespondentById(ctx, usr.RespondentID)
	if err != nil {
		return utils.SendError(errors.New("Anda tidak memiliki hak akses"), http.StatusInternalServerError)
	}

	data, totalData, err := service.surveyRepo.GetListHasilSurvey(usr, respondentLogin, payload)
	if err != nil {
		return utils.SendError(err, http.StatusInternalServerError)
	}

	hd := hashids.NewData()
	hd.Salt = os.Getenv("HASHID_SALT")
	hd.MinLength = 24

	h, err := hashids.NewWithData(hd)
	if err != nil {
		return utils.SendError(err, http.StatusInternalServerError)
	}

	now := time.Now()

	for i := range data {
		surveyId := []int{int(data[i].ID)}

		surveyCode, err := h.Encode(surveyId)
		if err != nil {
			return utils.SendError(err, http.StatusInternalServerError)
		}

		data[i].SurveyCode = surveyCode

		surveyDimulai := utils.ParseToWIB(data[i].SurveyDimulai)
		surveyBerakhir := utils.ParseToWIB(data[i].SurveyBerakhir)

		if surveyDimulai.After(now) {
			data[i].Status = string(enums.STATUS_SURVEY_UPCOMING)
		} else if surveyDimulai.Before(now) && surveyBerakhir.After(now) {
			data[i].Status = string(enums.STATUS_SURVEY_ONGOING)
		} else if surveyBerakhir.Before(now) {
			data[i].Status = string(enums.STATUS_SURVEY_FINISHED)
		}
	}

	totalPages := int(math.Ceil(float64(totalData) / float64(payload.Limit)))

	result := map[string]interface{}{
		"data": data,
		"meta": map[string]interface{}{
			"total":      totalData,
			"page":       payload.Page,
			"limit":      payload.Limit,
			"totalPages": totalPages,
		},
	}

	return utils.SendData(result, "Berhasil mengambil list survey")
}

func (service *hasilSurveyService) SurveyResultSummary(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()

	hd := hashids.NewData()
	hd.Salt = os.Getenv("HASHID_SALT")
	hd.MinLength = 24
	h, err := hashids.NewWithData(hd)
	if err != nil {
		return utils.SendError(err, http.StatusInternalServerError)
	}

	surveyCode, ok := slug["survey_code"].(string)
	if !ok {
		return utils.SendError(errors.New("Kode survey tidak valid"), http.StatusBadRequest)
	}

	decodedSurveyIDs, err := h.DecodeWithError(surveyCode)
	if err != nil || len(decodedSurveyIDs) == 0 {
		return utils.SendError(errors.New("Kode survey tidak valid atau dimanipulasi"), http.StatusBadRequest)
	}
	surveyID := int64(decodedSurveyIDs[0])

	survey, err := service.surveyRepo.GetSurveyById(surveyID)
	if err != nil {
		return utils.SendError(errors.New("Survey tidak ditemukan"), http.StatusNotFound)
	}

	respondentLoginId := usr.RespondentID
	if respondentLoginId == 0 {
		return utils.SendError(errors.New("Respondent tidak ditemukan"), http.StatusNotFound)
	}

	respondentLogin, err := service.userRepo.GetRespondentById(ctx, respondentLoginId)
	if err != nil || respondentLogin == nil {
		return utils.SendError(errors.New("Anda tidak memiliki hak akses"), http.StatusUnauthorized)
	}

	if *respondentLogin.RoleId != int64(enums.ROLE_ADMIN) &&
		*respondentLogin.RoleId != int64(enums.ROLE_SURVEYOR) {
		return utils.SendError(errors.New("Anda tidak memiliki hak akses untuk melihat hasil survey ini"), http.StatusUnauthorized)
	}

	rawRows, err := service.surveyRepo.GetSurveyResultRaw(ctx, surveyID)
	if err != nil {
		return utils.SendError(errors.New("Gagal memuat hasil survey"), http.StatusInternalServerError)
	}
	if len(rawRows) == 0 {
		return utils.SendError(errors.New("Survey ini belum memiliki pertanyaan"), http.StatusNotFound)
	}

	totalResponden, err := service.surveyRepo.GetTotalValidatedResponden(ctx, surveyID)
	if err != nil {
		return utils.SendError(errors.New("Gagal menghitung total responden"), http.StatusInternalServerError)
	}

	type questionAcc struct {
		summary       response.SurveyResultQuestionSummary
		answerTallies []models.SurveyResultRawDTO
	}

	order := []int{}
	grouped := make(map[int]*questionAcc)

	for _, row := range rawRows {
		acc, exists := grouped[row.FormFieldID]
		if !exists {
			questionLabel := row.Question
			if questionLabel == "" {
				questionLabel = row.Deskripsi
			}

			acc = &questionAcc{
				summary: response.SurveyResultQuestionSummary{
					QuestionID:    row.FormFieldID,
					Question:      questionLabel,
					Sequence:      row.Sequence,
					Template:      row.Template,
					TemplateLabel: templateLabel(row.Template),
					HasChart:      chartableTemplate(row.Template),
				},
			}
			grouped[row.FormFieldID] = acc
			order = append(order, row.FormFieldID)
		}

		if row.Answer != "" && row.Total > 0 {
			acc.answerTallies = append(acc.answerTallies, row)
			acc.summary.TotalAnswered += row.Total
		}
	}

	var mcFieldIDs []int
	for _, fieldID := range order {
		acc := grouped[fieldID]
		if acc.summary.Template == "multiple-choices" || acc.summary.Template == "checkboxes" {
			mcFieldIDs = append(mcFieldIDs, fieldID)
		}
	}

	optionsByField := make(map[int][]models.FormAnswerField)
	if len(mcFieldIDs) > 0 {
		options, errOpt := service.manajemenAlurRepo.GetAnswerOptionsByQuestionIDList(mcFieldIDs)
		if errOpt != nil {
			return utils.SendError(errors.New("Gagal memuat opsi jawaban pertanyaan"), http.StatusInternalServerError)
		}
		for _, opt := range options {
			optionsByField[opt.FormFieldId] = append(optionsByField[opt.FormFieldId], opt)
		}
	}

	var questionSummaries []response.SurveyResultQuestionSummary

	for _, fieldID := range order {
		acc := grouped[fieldID]

		switch acc.summary.Template {
		case "number":
			if acc.summary.TotalAnswered > 0 {
				acc.summary.ChartData = buildNumberChartData(acc.answerTallies, acc.summary.TotalAnswered)
			}

		case "multiple-choices":
			acc.summary.ChartData = buildMultipleChoiceChartData(acc.answerTallies, optionsByField[fieldID], acc.summary.TotalAnswered)

		case "checkboxes":
			acc.summary.ChartData = buildCheckboxChartData(acc.answerTallies, optionsByField[fieldID], acc.summary.TotalAnswered)
		}

		questionSummaries = append(questionSummaries, acc.summary)
	}

	dataResponse := response.SurveyResultSummaryResponse{
		SurveyCode:     surveyCode,
		SurveyName:     survey.Name,
		TotalResponden: totalResponden,
		Questions:      questionSummaries,
	}

	return utils.SendData(dataResponse, "Berhasil mengambil ringkasan hasil survey")
}

func buildNumberChartData(tallies []models.SurveyResultRawDTO, totalAnswered int64) []response.SurveyResultChartItem {
	var chartItems []response.SurveyResultChartItem

	for _, t := range tallies {
		pct := (float64(t.Total) / float64(totalAnswered)) * 100
		chartItems = append(chartItems, response.SurveyResultChartItem{
			Label:      t.Answer,
			Value:      t.Total,
			Percentage: pct,
		})
	}

	return chartItems
}

func buildMultipleChoiceChartData(tallies []models.SurveyResultRawDTO, options []models.FormAnswerField, totalAnswered int64) []response.SurveyResultChartItem {
	countByOptionID := make(map[int]int64)
	for _, t := range tallies {
		optID, errConv := strconv.Atoi(t.Answer)
		if errConv == nil {
			countByOptionID[optID] += t.Total
		}
	}

	chartItems := make([]response.SurveyResultChartItem, 0, len(options))

	// iterasi dari daftar OPSI (form_answer_fields), bukan dari jawaban,
	// supaya opsi dengan 0 suara tetap muncul di chart
	for _, opt := range options {
		count := countByOptionID[opt.ID]

		var pct float64
		if totalAnswered > 0 {
			pct = (float64(count) / float64(totalAnswered)) * 100
		}

		optIDCopy := opt.ID
		chartItems = append(chartItems, response.SurveyResultChartItem{
			OptionID:   &optIDCopy,
			Label:      opt.Option,
			Value:      count,
			Percentage: pct,
		})
	}

	return chartItems
}

func buildCheckboxChartData(tallies []models.SurveyResultRawDTO, options []models.FormAnswerField, totalAnswered int64) []response.SurveyResultChartItem {
	countByOptionID := make(map[int]int64)
	for _, t := range tallies {
		var optIDs []float64
		if errUnm := json.Unmarshal([]byte(t.Answer), &optIDs); errUnm == nil {
			for _, idFloat := range optIDs {
				countByOptionID[int(idFloat)] += t.Total
			}
		}
	}

	chartItems := make([]response.SurveyResultChartItem, 0, len(options))

	for _, opt := range options {
		count := countByOptionID[opt.ID]

		var pct float64
		if totalAnswered > 0 {
			pct = (float64(count) / float64(totalAnswered)) * 100
		}

		optIDCopy := opt.ID
		chartItems = append(chartItems, response.SurveyResultChartItem{
			OptionID:   &optIDCopy,
			Label:      opt.Option,
			Value:      count,
			Percentage: pct,
		})
	}

	return chartItems
}

func templateLabel(template string) string {
	switch template {
	case "number":
		return "Tipe Angka"
	case "multiple-choices":
		return "Tipe Pilihan Ganda"
	case "checkboxes":
		return "Tipe Checkbox"
	case "long-answer":
		return "Tipe Jawaban Panjang"
	case "short-answer":
		return "Tipe Jawaban Singkat"
	case "maps":
		return "Tipe Lokasi"
	case "image-template":
		return "Tipe Gambar"
	default:
		return "Tipe Lainnya"
	}
}

func chartableTemplate(template string) bool {
	switch template {
	case "number", "multiple-choices", "checkboxes":
		return true
	default:
		return false
	}
}

func (service *hasilSurveyService) SurveyResultQuestionDetail(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()

	hd := hashids.NewData()
	hd.Salt = os.Getenv("HASHID_SALT")
	hd.MinLength = 24
	h, err := hashids.NewWithData(hd)
	if err != nil {
		return utils.SendError(err, http.StatusInternalServerError)
	}

	surveyCodeStr, ok := slug["survey_code"].(string)
	if !ok {
		return utils.SendError(errors.New("Kode survey tidak valid"), http.StatusBadRequest)
	}
	decodedSurveyIDs, err := h.DecodeWithError(surveyCodeStr)
	if err != nil || len(decodedSurveyIDs) == 0 {
		return utils.SendError(errors.New("Kode survey tidak valid atau dimanipulasi"), http.StatusBadRequest)
	}
	surveyID := int64(decodedSurveyIDs[0])

	questionIdStr, ok := slug["question_id"].(string)
	if !ok {
		return utils.SendError(errors.New("ID pertanyaan tidak valid"), http.StatusBadRequest)
	}
	formFieldID, err := strconv.Atoi(questionIdStr)
	if err != nil || formFieldID <= 0 {
		return utils.SendError(errors.New("ID pertanyaan tidak valid"), http.StatusBadRequest)
	}

	respondentLoginId := usr.RespondentID
	if respondentLoginId == 0 {
		return utils.SendError(errors.New("Respondent tidak ditemukan"), http.StatusNotFound)
	}

	respondentLogin, err := service.userRepo.GetRespondentById(ctx, respondentLoginId)
	if err != nil || respondentLogin == nil {
		return utils.SendError(errors.New("Anda tidak memiliki hak akses"), http.StatusUnauthorized)
	}

	if *respondentLogin.RoleId != int64(enums.ROLE_ADMIN) &&
		*respondentLogin.RoleId != int64(enums.ROLE_SURVEYOR) {
		return utils.SendError(errors.New("Anda tidak memiliki hak akses untuk melihat hasil survey ini"), http.StatusUnauthorized)
	}

	formFieldDetail, err := service.templateFormulirPertanyaanRepo.GetFormFieldByID(formFieldID)
	if err != nil || formFieldDetail == nil {
		return utils.SendError(errors.New("Pertanyaan tidak ditemukan"), http.StatusNotFound)
	}

	questionLabel := formFieldDetail.Question
	if questionLabel == "" {
		if formFieldDetail.Deskripsi != nil {
			questionLabel = *formFieldDetail.Deskripsi
		}
	}

	page, _ := strconv.Atoi(param.Get("page"))
	limit, _ := strconv.Atoi(param.Get("limit"))

	datatablePayload := payloads.DatatablePayload{
		Search:   param.Get("search"),
		Page:     page,
		Limit:    limit,
		OrderBy:  param.Get("order_by"),
		OrderDir: param.Get("order_dir"),
	}

	var validate = validator.New()
	err = validate.Struct(datatablePayload)
	if err != nil {
		for _, err := range err.(validator.ValidationErrors) {
			customErrorMsg := utils.TranslateError(err)
			return utils.SendError(errors.New(customErrorMsg), http.StatusBadRequest)
		}
	}

	if datatablePayload.Page <= 0 {
		datatablePayload.Page = 1
	}
	if datatablePayload.Limit <= 0 {
		datatablePayload.Limit = 5
	}

	rawRows, totalData, err := service.surveyRepo.GetSurveyResultDetailByQuestion(ctx, surveyID, int64(formFieldID), datatablePayload)
	if err != nil {
		return utils.SendError(errors.New("Gagal memuat detail jawaban pertanyaan"), http.StatusInternalServerError)
	}

	optionLabelMap := make(map[int]string)
	if formFieldDetail.Template == "multiple-choices" {
		options, errOpt := service.manajemenAlurRepo.GetAnswerOptionsByQuestionIDList([]int{formFieldID})
		if errOpt != nil {
			return utils.SendError(errors.New("Gagal memuat opsi jawaban pertanyaan"), http.StatusInternalServerError)
		}
		for _, opt := range options {
			optionLabelMap[opt.ID] = opt.Option
		}
	}

	offset := (datatablePayload.Page - 1) * datatablePayload.Limit

	items := make([]map[string]interface{}, 0, len(rawRows))
	for i, row := range rawRows {
		items = append(items, map[string]interface{}{
			"no":     offset + i + 1,
			"name":   row.RespondentName,
			"answer": formatAnswerByTemplate(formFieldDetail.Template, row.Answer, optionLabelMap),
		})
	}

	totalPages := int(math.Ceil(float64(totalData) / float64(datatablePayload.Limit)))

	result := map[string]interface{}{
		"question_id": questionIdStr,
		"question":    questionLabel,
		"template":    formFieldDetail.Template,
		"data":        items,
		"meta": map[string]interface{}{
			"limit":      datatablePayload.Limit,
			"page":       datatablePayload.Page,
			"total":      totalData,
			"totalPages": totalPages,
		},
	}

	return utils.SendData(result, "Berhasil mengambil detail jawaban pertanyaan")
}

func formatAnswerByTemplate(template string, rawAnswer string, optionLabelMap map[int]string) interface{} {
	switch template {

	case "multiple-choices":
		optID, errConv := strconv.Atoi(rawAnswer)
		if errConv == nil {
			if label, exists := optionLabelMap[optID]; exists {
				return label
			}
		}
		return rawAnswer

	case "maps":
		var coords []response.SurveyResultMapCoordinate
		if errUnm := json.Unmarshal([]byte(rawAnswer), &coords); errUnm == nil {
			return coords
		}
		return rawAnswer

	case "image-template":
		var paths []string
		if errUnm := json.Unmarshal([]byte(rawAnswer), &paths); errUnm == nil {
			for i, path := range paths {
				paths[i] = os.Getenv("API_GATEWAY_URL") + "/view-public-survey-image/" + path
			}
			return paths
		}

		return rawAnswer

	default:
		return rawAnswer
	}
}
