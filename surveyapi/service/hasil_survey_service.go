package service

import (
	"backend/siccore/pb"
	"backend/surveyapi/dto"
	"backend/surveyapi/enums"
	"backend/surveyapi/models"
	"backend/surveyapi/payloads"
	"backend/surveyapi/repository"
	"backend/surveyapi/response"
	"backend/surveyapi/utils"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/speps/go-hashids/v2"
	"github.com/xuri/excelize/v2"
)

type HasilSurveyService interface {
	GetListSurvey(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	SurveyResultSummary(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)

	SurveyResultQuestionDetail(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	SurveyResultRespondentList(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	SurveyResultRespondentDetail(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	ExportExcelSurveyResultPerRespondent(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)

	ExportExcelSurveyResultAll(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
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

func (service *hasilSurveyService) SurveyResultRespondentList(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
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
		datatablePayload.Limit = 25
	}

	rawRows, totalData, err := service.surveyRepo.GetSurveyResultRespondents(ctx, surveyID, datatablePayload)
	if err != nil {
		return utils.SendError(errors.New("Gagal memuat daftar responden survey"), http.StatusInternalServerError)
	}

	offset := (datatablePayload.Page - 1) * datatablePayload.Limit

	items := make([]map[string]interface{}, 0, len(rawRows))
	for i, row := range rawRows {
		statusLabel := "Sedang Berlangsung"
		if row.Status != nil && *row.Status == 2 &&
			row.StatusApproval != nil && *row.StatusApproval == "validated_lurah" {
			statusLabel = "Selesai"
		}

		var lastResponseAt interface{}
		if row.LastResponseAt != nil {
			lastResponseAt = row.LastResponseAt.Format("2006-01-02 15:04:05")
		} else {
			lastResponseAt = nil
		}

		items = append(items, map[string]interface{}{
			"id":                  row.SurveyRespondentID,
			"no":                  offset + i + 1,
			"nama_responden":      row.RespondentName,
			"tanggapan_terakhir":  lastResponseAt,
			"pertanyaan_terjawab": row.TotalAnswered,
			"status":              statusLabel,
		})
	}

	totalPages := int(math.Ceil(float64(totalData) / float64(datatablePayload.Limit)))

	result := map[string]interface{}{
		"survey_code":     surveyCodeStr,
		"survey_name":     survey.Name,
		"total_responden": totalData,
		"data":            items,
		"meta": map[string]interface{}{
			"limit":      datatablePayload.Limit,
			"page":       datatablePayload.Page,
			"total":      totalData,
			"totalPages": totalPages,
		},
	}

	return utils.SendData(result, "Berhasil mengambil daftar responden hasil survey")
}

func (service *hasilSurveyService) SurveyResultRespondentDetail(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
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

	survey, err := service.surveyRepo.GetSurveyById(surveyID)
	if err != nil {
		return utils.SendError(errors.New("Survey tidak ditemukan"), http.StatusNotFound)
	}

	respondentIdStr, ok := slug["survey_respondent_id"].(string)
	if !ok {
		return utils.SendError(errors.New("ID responden tidak valid"), http.StatusBadRequest)
	}
	surveyRespondentID, err := strconv.ParseInt(respondentIdStr, 10, 64)
	if err != nil || surveyRespondentID <= 0 {
		return utils.SendError(errors.New("ID responden tidak valid"), http.StatusBadRequest)
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

	surveyRespondent, err := service.surveyRepo.GetSurveyRespondentByID(ctx, surveyRespondentID)
	if err != nil || surveyRespondent == nil {
		return utils.SendError(errors.New("Data responden tidak ditemukan"), http.StatusNotFound)
	}

	// pastikan survey_respondent ini memang milik survey yang diminta (guard IDOR)
	if surveyRespondent.SurveyID != surveyID {
		return utils.SendError(errors.New("Data responden tidak sesuai dengan survey ini"), http.StatusBadRequest)
	}

	respondentDetail, err := service.userRepo.GetRespondentById(ctx, surveyRespondent.RespondentID)
	if err != nil || respondentDetail == nil {
		return utils.SendError(errors.New("Data responden tidak ditemukan"), http.StatusNotFound)
	}

	rawRows, err := service.surveyRepo.GetSurveyResultRespondentDetail(ctx, surveyID, surveyRespondentID)
	if err != nil {
		return utils.SendError(errors.New("Gagal memuat detail jawaban responden"), http.StatusInternalServerError)
	}
	if len(rawRows) == 0 {
		return utils.SendError(errors.New("Survey ini belum memiliki pertanyaan"), http.StatusNotFound)
	}

	// siapkan option label map untuk pertanyaan multiple-choices
	var mcFieldIDs []int
	for _, row := range rawRows {
		if row.Template == "multiple-choices" {
			mcFieldIDs = append(mcFieldIDs, row.FormFieldID)
		}
	}

	optionLabelMap := make(map[int]string)
	if len(mcFieldIDs) > 0 {
		options, errOpt := service.manajemenAlurRepo.GetAnswerOptionsByQuestionIDList(mcFieldIDs)
		if errOpt != nil {
			return utils.SendError(errors.New("Gagal memuat opsi jawaban pertanyaan"), http.StatusInternalServerError)
		}
		for _, opt := range options {
			optionLabelMap[opt.ID] = opt.Option
		}
	}

	questions := make([]map[string]interface{}, 0, len(rawRows))
	for i, row := range rawRows {
		questionLabel := row.Question
		if questionLabel == "" {
			questionLabel = row.Deskripsi
		}

		var answer interface{}
		if row.Answer == "" || row.Answer == "[SKIPPED_BY_LOGIC]" {
			answer = nil // ditandai belum terisi
		} else {
			answer = formatAnswerByTemplate(row.Template, row.Answer, optionLabelMap)
		}

		questions = append(questions, map[string]interface{}{
			"no":       i + 1,
			"question": questionLabel,
			"template": row.Template,
			"answer":   answer, // null kalau belum terisi -> FE tampilkan "Jawaban Belum Terisi"
		})
	}

	result := map[string]interface{}{
		"survey_code":     surveyCodeStr,
		"survey_name":     survey.Name,
		"respondent_name": respondentDetail.Name,
		"questions":       questions,
	}

	return utils.SendData(result, "Berhasil mengambil detail jawaban responden")
}

func (service *hasilSurveyService) ExportExcelSurveyResultPerRespondent(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()

	hd := hashids.NewData()
	hd.Salt = os.Getenv("HASHID_SALT")
	hd.MinLength = 24
	h, _ := hashids.NewWithData(hd)

	surveyCodeStr, ok := slug["survey_code"].(string)
	if !ok {
		return utils.SendError(errors.New("Kode survey tidak valid"), http.StatusBadRequest)
	}
	decodedSurveyIDs, err := h.DecodeWithError(surveyCodeStr)
	if err != nil || len(decodedSurveyIDs) == 0 {
		return utils.SendError(errors.New("Kode survey tidak valid atau dimanipulasi"), http.StatusBadRequest)
	}
	surveyID := int64(decodedSurveyIDs[0])

	survey, err := service.surveyRepo.GetSurveyById(surveyID)
	if err != nil {
		return utils.SendError(errors.New("Survey tidak ditemukan"), http.StatusNotFound)
	}

	respondentIdStr, ok := slug["survey_respondent_id"].(string)
	if !ok {
		return utils.SendError(errors.New("ID responden tidak valid"), http.StatusBadRequest)
	}
	surveyRespondentID, err := strconv.ParseInt(respondentIdStr, 10, 64)
	if err != nil || surveyRespondentID <= 0 {
		return utils.SendError(errors.New("ID responden tidak valid"), http.StatusBadRequest)
	}

	respondentLoginId := usr.RespondentID
	respondentLogin, err := service.userRepo.GetRespondentById(ctx, respondentLoginId)
	if err != nil || respondentLogin == nil {
		return utils.SendError(errors.New("Anda tidak memiliki hak akses"), http.StatusUnauthorized)
	}

	if *respondentLogin.RoleId != int64(enums.ROLE_ADMIN) &&
		*respondentLogin.RoleId != int64(enums.ROLE_SURVEYOR) {
		return utils.SendError(errors.New("Anda tidak memiliki hak akses menindak lanjut survey ini"), http.StatusUnauthorized)
	}

	surveyRespondent, err := service.surveyRepo.GetSurveyRespondentByID(ctx, surveyRespondentID)
	if err != nil || surveyRespondent == nil {
		return utils.SendError(errors.New("Data isian responden tidak ditemukan"), http.StatusNotFound)
	}

	if surveyRespondent.SurveyID != surveyID {
		return utils.SendError(errors.New("Data responden tidak sesuai dengan survey ini"), http.StatusBadRequest)
	}

	respondentDetail, err := service.userRepo.GetRespondentById(ctx, surveyRespondent.RespondentID)
	if err != nil || respondentDetail == nil {
		return utils.SendError(errors.New("Data responden tidak ditemukan"), http.StatusUnauthorized)
	}

	kecamatanName, kelurahanName, rwName, rtName := "-", "-", "-", "-"

	if respondentDetail.KecamatanId != nil {
		if kec, errKec := service.wilayahRepo.GetKecamatanById(ctx, *respondentDetail.KecamatanId); errKec == nil && kec != nil {
			kecamatanName = kec.SubDistrictName
		}
	}
	if respondentDetail.KelurahanId != nil {
		if kel, errKel := service.wilayahRepo.GetKelurahanById(ctx, *respondentDetail.KelurahanId); errKel == nil && kel != nil {
			kelurahanName = kel.VillageName
		}
	}
	if respondentDetail.RWId != nil {
		if rw, errRw := service.wilayahRepo.GetRWById(ctx, *respondentDetail.RWId); errRw == nil && rw != nil {
			rwName = rw.NamaRw
		}
	}
	if respondentDetail.RTId != nil {
		if rt, errRt := service.wilayahRepo.GetRTById(ctx, *respondentDetail.RTId); errRt == nil && rt != nil {
			rtName = rt.NamaRt
		}
	}

	rawRows, err := service.surveyRepo.GetSurveyResultRespondentDetail(ctx, surveyID, surveyRespondentID)
	if err != nil {
		return utils.SendError(errors.New("Gagal memuat detail jawaban responden"), http.StatusInternalServerError)
	}
	if len(rawRows) == 0 {
		return utils.SendError(errors.New("Survey ini belum memiliki pertanyaan"), http.StatusNotFound)
	}

	var mcFieldIDs []int
	for _, row := range rawRows {
		if row.Template == "multiple-choices" || row.Template == "checkboxes" {
			mcFieldIDs = append(mcFieldIDs, row.FormFieldID)
		}
	}
	optionMap := make(map[int]string)
	if len(mcFieldIDs) > 0 {
		options, _ := service.manajemenAlurRepo.GetAnswerOptionsByQuestionIDList(mcFieldIDs)
		for _, opt := range options {
			if opt.Option != "" {
				optionMap[opt.ID] = opt.Option
			}
		}
	}

	f := excelize.NewFile()
	defer f.Close()
	sheetName := "Hasil Survey"
	if errRename := f.SetSheetName("Sheet1", sheetName); errRename != nil {
		sheetName = "Sheet1"
	}

	styleTitle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Size: 14},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center", WrapText: true},
	})
	styleLabel, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true},
	})
	styleHeader, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Color: "FFFFFF"},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"#15406a"}, Pattern: 1},
		Border:    []excelize.Border{{Type: "left", Color: "000000", Style: 1}, {Type: "top", Color: "000000", Style: 1}, {Type: "bottom", Color: "000000", Style: 1}, {Type: "right", Color: "000000", Style: 1}},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center", WrapText: true},
	})
	styleDataCenter, _ := f.NewStyle(&excelize.Style{
		Border:    []excelize.Border{{Type: "left", Color: "000000", Style: 1}, {Type: "top", Color: "000000", Style: 1}, {Type: "bottom", Color: "000000", Style: 1}, {Type: "right", Color: "000000", Style: 1}},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "top"},
	})
	styleDataLeft, _ := f.NewStyle(&excelize.Style{
		Border:    []excelize.Border{{Type: "left", Color: "000000", Style: 1}, {Type: "top", Color: "000000", Style: 1}, {Type: "bottom", Color: "000000", Style: 1}, {Type: "right", Color: "000000", Style: 1}},
		Alignment: &excelize.Alignment{Vertical: "top", WrapText: true},
	})
	styleDataText, _ := f.NewStyle(&excelize.Style{
		NumFmt:    49,
		Border:    []excelize.Border{{Type: "left", Color: "000000", Style: 1}, {Type: "top", Color: "000000", Style: 1}, {Type: "bottom", Color: "000000", Style: 1}, {Type: "right", Color: "000000", Style: 1}},
		Alignment: &excelize.Alignment{Vertical: "top", WrapText: true},
	})

	f.MergeCell(sheetName, "A1", "C3")
	titleText := fmt.Sprintf("Hasil Jawaban %s - %s", survey.Name, respondentDetail.Name)
	f.SetCellValue(sheetName, "A1", titleText)
	f.SetCellStyle(sheetName, "A1", "C3", styleTitle)

	infos := []struct{ row, label, value string }{
		{"5", "Kecamatan", kecamatanName},
		{"6", "Kelurahan", kelurahanName},
		{"7", "RW", rwName},
		{"8", "RT", rtName},
	}
	for _, info := range infos {
		f.MergeCell(sheetName, "A"+info.row, "B"+info.row)
		f.SetCellValue(sheetName, "A"+info.row, info.label)
		f.SetCellValue(sheetName, "C"+info.row, "\u200B"+info.value)
		f.SetCellStyle(sheetName, "A"+info.row, "C"+info.row, styleLabel)
	}

	rowHeader := 10
	f.SetCellValue(sheetName, fmt.Sprintf("A%d", rowHeader), "NO")
	f.SetCellValue(sheetName, fmt.Sprintf("B%d", rowHeader), "PERTANYAAN")
	f.SetCellValue(sheetName, fmt.Sprintf("C%d", rowHeader), "JAWABAN")
	f.SetCellStyle(sheetName, fmt.Sprintf("A%d", rowHeader), fmt.Sprintf("C%d", rowHeader), styleHeader)

	f.SetColWidth(sheetName, "A", "A", 6)
	f.SetColWidth(sheetName, "B", "B", 65)
	f.SetColWidth(sheetName, "C", "C", 55)

	currentRow := 11
	for index, row := range rawRows {
		questionLabel := row.Question
		if questionLabel == "" {
			questionLabel = row.Deskripsi
		}

		f.SetCellValue(sheetName, fmt.Sprintf("A%d", currentRow), index+1)
		f.SetCellValue(sheetName, fmt.Sprintf("B%d", currentRow), questionLabel)

		jawabanText := "-"
		if row.Answer != "" && row.Answer != "[SKIPPED_BY_LOGIC]" {
			jawabanText = row.Answer

			switch row.Template {
			case "number":
				jawabanText = "\u200B" + jawabanText

			case "multiple-choices":
				optID, errConv := strconv.Atoi(jawabanText)
				if errConv == nil {
					if labelText, exists := optionMap[optID]; exists {
						jawabanText = labelText
					}
				}

			case "checkboxes":
				var optIDs []float64
				if errUnm := json.Unmarshal([]byte(jawabanText), &optIDs); errUnm == nil {
					var labelTexts []string
					for _, idFloat := range optIDs {
						if label, exists := optionMap[int(idFloat)]; exists {
							labelTexts = append(labelTexts, "- "+label)
						}
					}
					if len(labelTexts) > 0 {
						jawabanText = strings.Join(labelTexts, "\n")
					}
				}

			case "maps":
				var mapsData []struct {
					Lat float64 `json:"lat"`
					Lng float64 `json:"lng"`
				}
				if errUnm := json.Unmarshal([]byte(jawabanText), &mapsData); errUnm == nil && len(mapsData) > 0 {
					jawabanText = fmt.Sprintf("%f,%f", mapsData[0].Lat, mapsData[0].Lng)
				}

			case "image-template":
				var jsonArray []string
				var finalJawaban string
				if errUnm := json.Unmarshal([]byte(jawabanText), &jsonArray); errUnm == nil {
					for imgIdx := range jsonArray {
						jsonArray[imgIdx] = os.Getenv("API_GATEWAY_URL") + "/view-public-survey-image/" + jsonArray[imgIdx]
						if imgIdx > 0 {
							finalJawaban += "\n\n"
						}
						finalJawaban += jsonArray[imgIdx]
					}
					jawabanText = finalJawaban
				}
			}
		}

		f.SetCellStr(sheetName, fmt.Sprintf("C%d", currentRow), jawabanText)

		f.SetCellStyle(sheetName, fmt.Sprintf("A%d", currentRow), fmt.Sprintf("A%d", currentRow), styleDataCenter)
		f.SetCellStyle(sheetName, fmt.Sprintf("B%d", currentRow), fmt.Sprintf("B%d", currentRow), styleDataLeft)

		if row.Template == "number" {
			f.SetCellStyle(sheetName, fmt.Sprintf("C%d", currentRow), fmt.Sprintf("C%d", currentRow), styleDataText)
		} else {
			f.SetCellStyle(sheetName, fmt.Sprintf("C%d", currentRow), fmt.Sprintf("C%d", currentRow), styleDataLeft)
		}

		currentRow++
	}

	var buffer bytes.Buffer
	if err := f.Write(&buffer); err != nil {
		return utils.SendError(errors.New("Gagal menyusun file excel"), http.StatusInternalServerError)
	}

	mimeType := "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	return utils.SetResponseData(buffer.Bytes(), true, "Data File,"+mimeType, http.StatusOK, nil, ""), nil
}

func (service *hasilSurveyService) ExportExcelSurveyResultAll(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()

	hd := hashids.NewData()
	hd.Salt = os.Getenv("HASHID_SALT")
	hd.MinLength = 24
	h, _ := hashids.NewWithData(hd)

	surveyCodeStr, ok := slug["survey_code"].(string)
	if !ok || surveyCodeStr == "" {
		return utils.SendError(errors.New("Kode survey tidak valid"), http.StatusBadRequest)
	}

	decodedSurveyIDs, err := h.DecodeWithError(surveyCodeStr)
	if err != nil || len(decodedSurveyIDs) == 0 {
		return utils.SendError(errors.New("Kode survey tidak valid atau dimanipulasi"), http.StatusBadRequest)
	}
	surveyID := int64(decodedSurveyIDs[0])

	survey, err := service.surveyRepo.GetSurveyById(surveyID)
	if err != nil {
		return utils.SendError(errors.New("Survey tidak ditemukan"), http.StatusNotFound)
	}

	respondentLoginId := usr.RespondentID
	respondentLogin, err := service.userRepo.GetRespondentById(ctx, respondentLoginId)
	if err != nil || respondentLogin == nil {
		return utils.SendError(errors.New("Anda tidak memiliki hak akses"), http.StatusUnauthorized)
	}

	if *respondentLogin.RoleId != int64(enums.ROLE_ADMIN) &&
		*respondentLogin.RoleId != int64(enums.ROLE_KECAMATAN) &&
		*respondentLogin.RoleId != int64(enums.ROLE_KELURAHAN) &&
		*respondentLogin.RoleId != int64(enums.ROLE_RW) {
		return utils.SendError(errors.New("Anda tidak memiliki hak akses menindak lanjut survey ini"), http.StatusUnauthorized)
	}

	rawNodes, err := service.manajemenAlurRepo.GetRawNodesForPreview(int(survey.FlowDetailID), 0)
	if err != nil {
		return utils.SendError(errors.New("Gagal memuat struktur pertanyaan"), http.StatusInternalServerError)
	}
	if len(rawNodes) == 0 {
		return utils.SendError(errors.New("Survey ini belum memiliki pertanyaan"), http.StatusNotFound)
	}

	var blueprintFieldIDs []int
	for _, node := range rawNodes {
		blueprintFieldIDs = append(blueprintFieldIDs, node.FormFieldId)
	}
	options, _ := service.manajemenAlurRepo.GetAnswerOptionsByQuestionIDList(blueprintFieldIDs)

	optionMap := make(map[int]string)
	for _, opt := range options {
		if opt.Option != "" {
			optionMap[opt.ID] = opt.Option
		}
	}

	rawJawaban, err := service.surveyRepo.GetAllJawabanForExport(ctx, surveyID)
	if err != nil {
		return utils.SendError(errors.New("Gagal memuat hasil jawaban survei"), http.StatusInternalServerError)
	}

	rekapMap := make(map[int64]*dto.RekapRespondenWilayahExcel)
	var orderedRespondentIDs []int64

	for _, raw := range rawJawaban {
		if _, exists := rekapMap[raw.RespondentID]; !exists {
			waktuTeks := "-"
			if raw.WaktuSelesai != nil {
				waktuTeks = raw.WaktuSelesai.Format("02 Jan 2006 15:04:05")
			}
			rekapMap[raw.RespondentID] = &dto.RekapRespondenWilayahExcel{
				NamaResponden: raw.NamaResponden,
				KecamatanName: raw.KecamatanName,
				KelurahanName: raw.KelurahanName,
				RwName:        raw.RwName,
				RtName:        raw.RtName,
				WaktuSelesai:  waktuTeks,
				JawabanMap:    make(map[int]string),
			}
			orderedRespondentIDs = append(orderedRespondentIDs, raw.RespondentID)
		}
		if raw.Answer != nil {
			rekapMap[raw.RespondentID].JawabanMap[raw.FormFieldID] = *raw.Answer
		}
	}

	totalResp := len(orderedRespondentIDs)

	f := excelize.NewFile()
	defer f.Close()

	sheetName1 := "Rekap Hasil Per Responden"
	f.SetSheetName("Sheet1", sheetName1)

	styleHeader, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Color: "FFFFFF"},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"#15406a"}, Pattern: 1},
		Border:    []excelize.Border{{Type: "left", Color: "000000", Style: 1}, {Type: "top", Color: "000000", Style: 1}, {Type: "bottom", Color: "000000", Style: 1}, {Type: "right", Color: "000000", Style: 1}},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center", WrapText: true},
	})
	styleData, _ := f.NewStyle(&excelize.Style{
		Border:    []excelize.Border{{Type: "left", Color: "000000", Style: 1}, {Type: "top", Color: "000000", Style: 1}, {Type: "bottom", Color: "000000", Style: 1}, {Type: "right", Color: "000000", Style: 1}},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center", WrapText: true},
	})
	styleLabel, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true},
		Alignment: &excelize.Alignment{Vertical: "center", Horizontal: "left"},
	})
	styleInfoValue, _ := f.NewStyle(&excelize.Style{
		Alignment: &excelize.Alignment{Vertical: "center", Horizontal: "left"},
	})

	startDateStr, endDateStr := "-", "-"
	if !survey.StartDate.IsZero() {
		startDateStr = survey.StartDate.Format("02 Jan 2006")
	}
	if !survey.EndDate.IsZero() {
		endDateStr = survey.EndDate.Format("02 Jan 2006")
	}

	// --- Info header (row 4-6) ---
	f.MergeCell(sheetName1, "A4", "B4")
	f.MergeCell(sheetName1, "A5", "B5")
	f.MergeCell(sheetName1, "A6", "B6")
	f.MergeCell(sheetName1, "C4", "F4")
	f.MergeCell(sheetName1, "C5", "F5")
	f.MergeCell(sheetName1, "C6", "F6")

	f.SetCellValue(sheetName1, "A4", "Nama Survey")
	f.SetCellStyle(sheetName1, "A4", "A4", styleLabel)
	f.SetCellValue(sheetName1, "C4", survey.Name)
	f.SetCellStyle(sheetName1, "C4", "C4", styleInfoValue)

	f.SetCellValue(sheetName1, "A5", "Tanggal Survey")
	f.SetCellStyle(sheetName1, "A5", "A5", styleLabel)
	f.SetCellValue(sheetName1, "C5", fmt.Sprintf("%s - %s", startDateStr, endDateStr))
	f.SetCellStyle(sheetName1, "C5", "C5", styleInfoValue)

	f.SetCellValue(sheetName1, "A6", "Total Responden")
	f.SetCellStyle(sheetName1, "A6", "A6", styleLabel)
	f.SetCellValue(sheetName1, "C6", fmt.Sprintf("%d Responden", totalResp))
	f.SetCellStyle(sheetName1, "C6", "C6", styleInfoValue)

	// --- Header tabel (row 8): No | Nama Responden | Kecamatan | Kelurahan | RW | RT | Pertanyaan 1..N ---
	rowHeader := 8
	staticHeaders := []string{"No", "Nama Responden", "Kecamatan", "Kelurahan", "RW", "RT"}
	colIndex := 1
	for _, hLabel := range staticHeaders {
		colName, _ := excelize.ColumnNumberToName(colIndex)
		f.SetCellValue(sheetName1, fmt.Sprintf("%s%d", colName, rowHeader), hLabel)
		f.SetCellStyle(sheetName1, fmt.Sprintf("%s%d", colName, rowHeader), fmt.Sprintf("%s%d", colName, rowHeader), styleHeader)
		f.SetColWidth(sheetName1, colName, colName, 18)
		colIndex++
	}

	for i, node := range rawNodes {
		colName, _ := excelize.ColumnNumberToName(colIndex)
		f.SetCellValue(sheetName1, fmt.Sprintf("%s%d", colName, rowHeader), fmt.Sprintf("%d. %s", i+1, node.Label))
		f.SetCellStyle(sheetName1, fmt.Sprintf("%s%d", colName, rowHeader), fmt.Sprintf("%s%d", colName, rowHeader), styleHeader)
		f.SetColWidth(sheetName1, colName, colName, 35)
		colIndex++
	}

	// --- Data rows ---
	currentRow := rowHeader + 1
	no := 1
	for _, respondentID := range orderedRespondentIDs {
		data := rekapMap[respondentID]

		f.SetCellValue(sheetName1, fmt.Sprintf("A%d", currentRow), no)
		f.SetCellValue(sheetName1, fmt.Sprintf("B%d", currentRow), data.NamaResponden)
		f.SetCellValue(sheetName1, fmt.Sprintf("C%d", currentRow), data.KecamatanName)
		f.SetCellValue(sheetName1, fmt.Sprintf("D%d", currentRow), data.KelurahanName)
		f.SetCellValue(sheetName1, fmt.Sprintf("E%d", currentRow), "\u200B"+data.RwName)
		f.SetCellValue(sheetName1, fmt.Sprintf("F%d", currentRow), "\u200B"+data.RtName)

		colIdx := 7
		for _, node := range rawNodes {
			jawabanText := "-"
			if val, ok := data.JawabanMap[node.FormFieldId]; ok {
				jawabanText = val
				if jawabanText != "-" && jawabanText != "[SKIPPED_BY_LOGIC]" {
					switch node.Template {
					case "number":
						jawabanText = "\u200B" + jawabanText
					case "multiple-choices":
						if optID, errConv := strconv.Atoi(jawabanText); errConv == nil {
							if labelText, exists := optionMap[optID]; exists {
								jawabanText = labelText
							}
						}
					case "checkboxes":
						var optIDs []float64
						if errUnm := json.Unmarshal([]byte(jawabanText), &optIDs); errUnm == nil {
							var labelTexts []string
							for _, idFloat := range optIDs {
								if label, exists := optionMap[int(idFloat)]; exists {
									labelTexts = append(labelTexts, "- "+label)
								}
							}
							if len(labelTexts) > 0 {
								jawabanText = strings.Join(labelTexts, "\n")
							}
						}
					case "maps":
						var mapsData []struct {
							Lat float64 `json:"lat"`
							Lng float64 `json:"lng"`
						}
						if errUnm := json.Unmarshal([]byte(jawabanText), &mapsData); errUnm == nil && len(mapsData) > 0 {
							jawabanText = fmt.Sprintf("%f,%f", mapsData[0].Lat, mapsData[0].Lng)
						}
					case "image-template":
						var jsonArray []string
						var finalJawaban string
						if errUnm := json.Unmarshal([]byte(jawabanText), &jsonArray); errUnm == nil {
							for imgIdx := range jsonArray {
								jsonArray[imgIdx] = os.Getenv("API_GATEWAY_URL") + "/view-public-survey-image/" + jsonArray[imgIdx]
								if imgIdx > 0 {
									finalJawaban += "\n\n"
								}
								finalJawaban += jsonArray[imgIdx]
							}
							jawabanText = finalJawaban
						}
					}
				} else {
					jawabanText = "-"
				}
			}

			colName, _ := excelize.ColumnNumberToName(colIdx)
			f.SetCellValue(sheetName1, fmt.Sprintf("%s%d", colName, currentRow), jawabanText)
			colIdx++
		}

		lastColName, _ := excelize.ColumnNumberToName(colIdx - 1)
		f.SetCellStyle(sheetName1, fmt.Sprintf("A%d", currentRow), fmt.Sprintf("%s%d", lastColName, currentRow), styleData)
		currentRow++
		no++
	}

	// --- Sheet 2: Rekap Hasil Survey Per Kolom ---
	sheetStat := "Rekap Hasil Survey Per Kolom"
	f.NewSheet(sheetStat)

	styleTitle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Size: 14},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center", WrapText: true},
	})
	styleStatTitle, _ := f.NewStyle(&excelize.Style{Font: &excelize.Font{Bold: true}})

	f.MergeCell(sheetStat, "A1", "D2")
	f.SetCellValue(sheetStat, "A1", fmt.Sprintf("Hasil Survey %s", survey.Name))
	f.SetCellStyle(sheetStat, "A1", "D2", styleTitle)

	f.SetCellValue(sheetStat, "A4", "Nama Survey")
	f.SetCellValue(sheetStat, "B4", survey.Name)
	f.SetCellValue(sheetStat, "A5", "Tanggal Survey")
	f.SetCellValue(sheetStat, "B5", fmt.Sprintf("%s - %s", startDateStr, endDateStr))
	f.SetCellValue(sheetStat, "A6", "Total Responden")
	f.SetCellValue(sheetStat, "B6", fmt.Sprintf("%d Responden", totalResp))

	f.SetCellStyle(sheetStat, "A4", "A6", styleStatTitle)
	f.SetColWidth(sheetStat, "A", "A", 15)
	f.SetColWidth(sheetStat, "B", "B", 40)
	f.SetColWidth(sheetStat, "C", "C", 12)
	f.SetColWidth(sheetStat, "D", "D", 12)

	currentRowStat := 8

	for i, node := range rawNodes {
		qID := node.FormFieldId

		answeredCount := 0
		optionTallies := make(map[int]int)

		for _, respID := range orderedRespondentIDs {
			if data, ok := rekapMap[respID]; ok {
				if ans, exists := data.JawabanMap[qID]; exists && ans != "-" && ans != "[SKIPPED_BY_LOGIC]" && ans != "" {
					answeredCount++

					switch node.Template {
					case "multiple-choices":
						if optID, errConv := strconv.Atoi(ans); errConv == nil {
							optionTallies[optID]++
						}
					case "checkboxes":
						var optIDs []float64
						if errUnm := json.Unmarshal([]byte(ans), &optIDs); errUnm == nil {
							for _, idFloat := range optIDs {
								optionTallies[int(idFloat)]++
							}
						}
					}
				}
			}
		}

		f.SetCellValue(sheetStat, fmt.Sprintf("A%d", currentRowStat), fmt.Sprintf("Pertanyaan %d", i+1))
		f.SetCellValue(sheetStat, fmt.Sprintf("B%d", currentRowStat), node.Label)
		f.SetCellStyle(sheetStat, fmt.Sprintf("A%d", currentRowStat), fmt.Sprintf("B%d", currentRowStat), styleStatTitle)
		currentRowStat++

		f.SetCellValue(sheetStat, fmt.Sprintf("B%d", currentRowStat), "Responden :")

		respFraction := fmt.Sprintf("'%d/%d", answeredCount, totalResp)
		respPct := "'0%"
		if totalResp > 0 {
			respPct = fmt.Sprintf("'%.0f%%", (float64(answeredCount)/float64(totalResp))*100)
		}
		f.SetCellValue(sheetStat, fmt.Sprintf("C%d", currentRowStat), respFraction)
		f.SetCellValue(sheetStat, fmt.Sprintf("D%d", currentRowStat), respPct)
		currentRowStat++

		if node.Template == "multiple-choices" || node.Template == "checkboxes" {
			for _, opt := range options {
				if opt.FormFieldId == qID {
					optCount := optionTallies[opt.ID]
					optPct := "'0%"
					if totalResp > 0 {
						optPct = fmt.Sprintf("'%.0f%%", (float64(optCount)/float64(totalResp))*100)
					}

					f.SetCellValue(sheetStat, fmt.Sprintf("B%d", currentRowStat), "\u200B"+opt.Option)
					f.SetCellValue(sheetStat, fmt.Sprintf("C%d", currentRowStat), fmt.Sprintf("'%d/%d", optCount, totalResp))
					f.SetCellValue(sheetStat, fmt.Sprintf("D%d", currentRowStat), optPct)
					currentRowStat++
				}
			}
		}

		currentRowStat++
	}

	f.SetActiveSheet(0)

	var buffer bytes.Buffer
	if err := f.Write(&buffer); err != nil {
		return utils.SendError(errors.New("Gagal menyusun file excel massal"), http.StatusInternalServerError)
	}

	mimeType := "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	return utils.SetResponseData(buffer.Bytes(), true, "Data File,"+mimeType, http.StatusOK, nil, ""), nil
}
