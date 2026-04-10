package service

import (
	"backend/siccore/pb"
	"backend/surveyapi/enums"
	"backend/surveyapi/models"
	"backend/surveyapi/payloads"
	"backend/surveyapi/repository"
	"backend/surveyapi/response"
	"backend/surveyapi/utils"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"
)

type TemplateFormulirPertanyaanService interface {
	CreateTemplateFormulirPertanyaan(usr models.JwtCustomClaims, req map[string]interface{}) (*pb.ProxyResponse, error)
	GetDetailTemplateFormulirPertanyaan(usr models.JwtCustomClaims, slug map[string]interface{}) (*pb.ProxyResponse, error)
	UpdateTemplateFormulirPertanyaan(usr models.JwtCustomClaims, req map[string]interface{}, slug map[string]interface{}) (*pb.ProxyResponse, error)
	DuplicateTemplateFormulirPertanyaan(usr models.JwtCustomClaims, req map[string]interface{}, slug map[string]interface{}) (*pb.ProxyResponse, error)
	DeleteTemplateFormulirPertanyaan(usr models.JwtCustomClaims, req map[string]interface{}, slug map[string]interface{}) (*pb.ProxyResponse, error)
	GetListTemplateFormulirPertanyaan(usr models.JwtCustomClaims, req map[string]interface{}, slug map[string]interface{}) (*pb.ProxyResponse, error)
	GetQuestionTypeOptions(usr models.JwtCustomClaims, param url.Values) (*pb.ProxyResponse, error)
	GetFormulirPertanyaanOptions(param url.Values) (*pb.ProxyResponse, error)
	GetPertanyaanOptions(param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	DetailPertanyaan(usr models.JwtCustomClaims, slug map[string]interface{}) (*pb.ProxyResponse, error)
	GetMultipleChoiceOptionByFormFieldId(param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
}

type templateFormulirPertanyaanService struct {
	templateFormulirPertanyaanRepo repository.TemplateFormulirPertanyaanRepo
}

func NewTemplateFormulirPertanyaanService(
	templateFormulirPertanyaanRepo repository.TemplateFormulirPertanyaanRepo,
) TemplateFormulirPertanyaanService {
	return &templateFormulirPertanyaanService{
		templateFormulirPertanyaanRepo: templateFormulirPertanyaanRepo,
	}
}

func (service *templateFormulirPertanyaanService) CreateTemplateFormulirPertanyaan(usr models.JwtCustomClaims, req map[string]interface{}) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()

	var payload payloads.TemplateFormulirPertanyaanPayload

	err := utils.DynamicBind(req, &payload)
	if err != nil {
		return utils.SendError(err, http.StatusBadRequest)
	}

	var validate = validator.New()

	err = validate.Struct(payload)
	if err != nil {
		customErrorMsg := utils.FormatValidationError(err)
		return utils.SendError(errors.New(customErrorMsg), http.StatusBadRequest)
	}

	code, err := utils.GenerateUniqueString(32)
	if err != nil {
		return utils.SendError(err, http.StatusInternalServerError)
	}

	codeIsExist, err := service.templateFormulirPertanyaanRepo.TemplateFormCodeIsExist(code)
	if err != nil {
		return utils.SendError(err, http.StatusInternalServerError)
	}

	for codeIsExist {
		code, err = utils.GenerateUniqueString(32)
		if err != nil {
			return utils.SendError(err, http.StatusInternalServerError)
		}

		codeIsExist, err = service.templateFormulirPertanyaanRepo.TemplateFormCodeIsExist(code)
		if err != nil {
			return utils.SendError(err, http.StatusInternalServerError)
		}
	}

	flagTematik := "false"
	if *payload.IsTematikQuestion {
		flagTematik = "true"
	}

	form := models.Form{
		UserId:      1,
		Title:       payload.Title,
		Description: payload.Description,
		Code:        code,
		Status:      enums.OPEN,
		FlagTematik: flagTematik,
	}

	existingTemplateByTitle, err := service.templateFormulirPertanyaanRepo.GetFormByTitle(payload.Title)

	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return utils.SendError(err, http.StatusInternalServerError)
	}

	if existingTemplateByTitle != nil {
		return utils.SendError(errors.New("Nama template sudah digunakan, silakan gunakan nama lain"), http.StatusBadRequest)
	}

	var formFields []models.FormFieldWithOption

	for index, v := range payload.Questions {

		if !enums.QuestionType(v.InputType).IsQuestionTypeValid() {
			return utils.SendError(errors.New("Tipe pertanyaan tidak valid"), http.StatusBadRequest)
		}

		prefix := strings.ReplaceAll(v.InputType, "-", "_")
		newAttribute := fmt.Sprintf("%s.%s", prefix, utils.GenerateShortHash())

		attributeIsExists, err := service.templateFormulirPertanyaanRepo.TemplateFormAttributeIsExists(newAttribute)
		if err != nil {
			return utils.SendError(err, http.StatusInternalServerError)
		}

		for attributeIsExists {
			newAttribute = fmt.Sprintf("%s.%s", prefix, utils.GenerateShortHash())

		}

		var imageQtyStr *string
		if v.ImageQuantity != nil && *v.ImageQuantity != 0 {
			strVal := strconv.Itoa(*v.ImageQuantity)
			imageQtyStr = &strVal
		}

		var formFieldOptions []models.FormAnswerField
		for indexOption, vOption := range v.Options {
			formFieldOptions = append(formFieldOptions, models.FormAnswerField{
				Option:   vOption.Name,
				Sequence: strconv.Itoa(indexOption + 1),
			})
		}

		formField := models.FormFieldWithOption{
			FormField: models.FormField{
				Template:      v.InputType,
				Attribute:     newAttribute,
				Question:      v.Question,
				Deskripsi:     v.Description,
				Required:      v.Required,
				ImageQuantity: imageQtyStr,
				Sequence:      index + 1,
			},
			T_Options: formFieldOptions,
		}
		formFields = append(formFields, formField)
	}

	createData, err := service.templateFormulirPertanyaanRepo.CreateFormWithFields(form, formFields)

	if err != nil {
		return utils.SendError(err, http.StatusInternalServerError)
	}

	return utils.SendData(createData, "Berhasil membuat template formulir pertanyaan")
}

func (service *templateFormulirPertanyaanService) GetDetailTemplateFormulirPertanyaan(usr models.JwtCustomClaims, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()

	codeStr := slug["template_formulir_pertanyaan_code"]
	code, ok := codeStr.(string)
	if !ok {
		return utils.SendError(errors.New("Kode template formulir pertanyaan tidak valid"), http.StatusBadRequest)
	}

	form, err := service.templateFormulirPertanyaanRepo.GetFormByCode(code)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return utils.SendError(errors.New("Template formulir pertanyaan tidak ditemukan"), http.StatusNotFound)
		}
		return utils.SendError(err, http.StatusInternalServerError)
	}

	formField, err := service.templateFormulirPertanyaanRepo.GetFormFieldByFormId(form.ID)
	if err != nil {
		return utils.SendError(err, http.StatusInternalServerError)
	}

	var formFields []models.FormFieldWithOption
	for _, v := range formField {
		options, err := service.templateFormulirPertanyaanRepo.GetFormFieldOptionsByFormFieldId(v.ID)
		if err != nil {
			return utils.SendError(err, http.StatusInternalServerError)
		}

		formFields = append(formFields, models.FormFieldWithOption{
			FormField: v,
			T_Options: options,
		})
	}

	data := models.FormDetail{
		Form:       *form,
		FormFields: formFields,
	}

	return utils.SendData(data, "Berhasil mendapatkan detail template formulir pertanyaan")
}

func (service *templateFormulirPertanyaanService) UpdateTemplateFormulirPertanyaan(usr models.JwtCustomClaims, req map[string]interface{}, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()

	var payload payloads.UpdateTemplateFormulirPertanyaanPayload

	err := utils.DynamicBind(req, &payload)
	if err != nil {
		return utils.SendError(err, http.StatusBadRequest)
	}

	var validate = validator.New()
	err = validate.Struct(payload)
	if err != nil {
		customErrorMsg := utils.FormatValidationError(err)
		return utils.SendError(errors.New(customErrorMsg), http.StatusBadRequest)
	}

	codeStr := slug["template_formulir_pertanyaan_code"]
	code, ok := codeStr.(string)
	if !ok {
		return utils.SendError(errors.New("Kode template formulir pertanyaan tidak valid"), http.StatusBadRequest)
	}

	form, err := service.templateFormulirPertanyaanRepo.GetFormByCode(code)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return utils.SendError(errors.New("Template formulir pertanyaan tidak ditemukan"), http.StatusNotFound)
		}
		return utils.SendError(err, http.StatusInternalServerError)
	}

	flagTematik := "false"
	if *payload.IsTematikQuestion {
		flagTematik = "true"
	}

	if payload.Title != form.Title {
		existingTemplateByTitle, err := service.templateFormulirPertanyaanRepo.GetFormByTitle(payload.Title)
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return utils.SendError(err, http.StatusInternalServerError)
		}
		if existingTemplateByTitle != nil {
			return utils.SendError(errors.New("Nama template sudah digunakan, silakan gunakan nama lain"), http.StatusBadRequest)
		}
	}

	formUpdate := models.Form{
		ID:          form.ID,
		UserId:      form.UserId,
		Title:       payload.Title,
		Description: payload.Description,
		Code:        form.Code,
		Status:      enums.OPEN,
		FlagTematik: flagTematik,
		CreatedAt:   form.CreatedAt,
	}

	var attributeToKeep []string
	var formFields []models.UpdateFormFieldWithOption

	var isQuestionUpdate bool
	if payload.IsQuestionsUpdated != nil {
		isQuestionUpdate = *payload.IsQuestionsUpdated
	}

	if isQuestionUpdate == true {
		existingFields, _ := service.templateFormulirPertanyaanRepo.GetAllFieldsByFormID(form.ID)

		fieldMap := make(map[string]models.FormField)
		for _, f := range existingFields {
			fieldMap[f.Attribute] = f
		}

		for seqIndex, q := range payload.Questions {
			var formFieldID int
			var currentAttribute string

			if q.Attribute != "" {
				existingFormField, exists := fieldMap[q.Attribute]
				if !exists {
					return utils.SendError(errors.New("Atribut pertanyaan tidak ditemukan: "+q.Attribute), http.StatusBadRequest)
				}

				currentAttribute = q.Attribute
				formFieldID = existingFormField.ID
				attributeToKeep = append(attributeToKeep, currentAttribute)
			} else {
				prefix := strings.ReplaceAll(q.InputType, "-", "_")
				currentAttribute = fmt.Sprintf("%s.%s", prefix, utils.GenerateShortHash())
			}

			// Susun options jika tipe input adalah multiple-choices
			var options []models.FormAnswerField
			var optionsToKeep []int

			if q.InputType == "multiple-choices" {
				for optIndex, opt := range q.Options {
					if opt.ID != nil && *opt.ID != 0 {
						optionsToKeep = append(optionsToKeep, *opt.ID)
					}

					var optID int
					if opt.ID != nil {
						optID = *opt.ID
					}

					options = append(options, models.FormAnswerField{
						ID:          optID,
						FormFieldId: formFieldID,
						Option:      opt.Name,
						Sequence:    strconv.Itoa(optIndex + 1),
						CreatedAt:   form.CreatedAt,
					})
				}
			}

			var imageQty *string
			if q.ImageQuantity != nil && *q.ImageQuantity != 0 {
				strVal := strconv.Itoa(*q.ImageQuantity)
				imageQty = &strVal
			}

			formFields = append(formFields, models.UpdateFormFieldWithOption{
				FormField: models.FormField{
					ID:            formFieldID,
					FormId:        form.ID,
					Template:      q.InputType,
					Question:      q.Question,
					Deskripsi:     q.Description,
					Required:      q.Required,
					ImageQuantity: imageQty,
					Attribute:     currentAttribute,
					Sequence:      seqIndex + 1,
					CreatedAt:     form.CreatedAt,
				},
				Options:       options,
				OptionsToKeep: optionsToKeep,
			})
		}
	}

	err = service.templateFormulirPertanyaanRepo.UpdateFormTransactionTx(formUpdate, formFields, attributeToKeep, isQuestionUpdate)
	if err != nil {
		return utils.SendError(err, http.StatusInternalServerError)
	}

	return utils.SendData(nil, "Berhasil memperbarui template formulir pertanyaan")
}

func (service *templateFormulirPertanyaanService) DuplicateTemplateFormulirPertanyaan(usr models.JwtCustomClaims, req map[string]interface{}, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()

	codeStr, ok := slug["template_formulir_pertanyaan_code"].(string)
	if !ok {
		return utils.SendError(errors.New("Kode template formulir pertanyaan tidak valid"), http.StatusBadRequest)
	}

	var payload payloads.TemplateFormPayload

	err := utils.DynamicBind(req, &payload)
	if err != nil {
		return utils.SendError(err, http.StatusBadRequest)
	}

	var validate = validator.New()

	err = validate.Struct(payload)
	if err != nil {
		customErrorMsg := utils.FormatValidationError(err)
		return utils.SendError(errors.New(customErrorMsg), http.StatusBadRequest)
	}

	existingTemplateByTitle, err := service.templateFormulirPertanyaanRepo.GetFormByTitle(payload.Title)

	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return utils.SendError(err, http.StatusInternalServerError)
	}

	if existingTemplateByTitle != nil {
		return utils.SendError(errors.New("Nama template sudah digunakan, silakan gunakan nama lain"), http.StatusBadRequest)
	}

	code, err := utils.GenerateUniqueString(32)
	if err != nil {
		return utils.SendError(err, http.StatusInternalServerError)
	}

	codeIsExist, err := service.templateFormulirPertanyaanRepo.TemplateFormCodeIsExist(code)
	if err != nil {
		return utils.SendError(err, http.StatusInternalServerError)
	}

	for codeIsExist {
		code, err = utils.GenerateUniqueString(32)
		if err != nil {
			return utils.SendError(err, http.StatusInternalServerError)
		}

		codeIsExist, err = service.templateFormulirPertanyaanRepo.TemplateFormCodeIsExist(code)
		if err != nil {
			return utils.SendError(err, http.StatusInternalServerError)
		}
	}

	flagTematik := "false"
	if *payload.IsTematikQuestion {
		flagTematik = "true"
	}

	originalForm, err := service.templateFormulirPertanyaanRepo.GetFullFormByCode(codeStr)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return utils.SendError(errors.New("Template formulir tidak ditemukan"), http.StatusNotFound)
		}
		return utils.SendError(err, http.StatusInternalServerError)
	}

	newForm := models.Form{
		UserId:      originalForm.UserId,
		Title:       payload.Title,
		Description: payload.Description,
		Code:        code,
		Status:      enums.OPEN,
		FlagTematik: flagTematik,
	}

	var newFields []models.FormFieldWithOption

	for _, origField := range originalForm.FormFields {
		prefix := strings.ReplaceAll(origField.Template, "-", "_")
		newAttribute := fmt.Sprintf("%s.%s", prefix, utils.GenerateShortHash())

		attributeIsExists, err := service.templateFormulirPertanyaanRepo.TemplateFormAttributeIsExists(newAttribute)
		if err != nil {
			return utils.SendError(err, http.StatusInternalServerError)
		}

		for attributeIsExists {
			newAttribute = fmt.Sprintf("%s.%s", prefix, utils.GenerateShortHash())
		}

		newField := models.FormField{
			Template:      origField.Template,
			Question:      origField.Question,
			Deskripsi:     origField.Deskripsi,
			Required:      origField.Required,
			ImageQuantity: origField.ImageQuantity,
			Attribute:     newAttribute,
			Sequence:      origField.Sequence,
		}

		var newOptions []models.FormAnswerField
		if origField.Template == "multiple-choices" {
			for _, origOpt := range origField.Options {
				newOptions = append(newOptions, models.FormAnswerField{
					Option:   origOpt.Option,
					Sequence: origOpt.Sequence,
				})
			}
		}

		newFields = append(newFields, models.FormFieldWithOption{
			FormField: newField,
			T_Options: newOptions,
		})
	}

	err = service.templateFormulirPertanyaanRepo.CreateFormTransactionTx(newForm, newFields)
	if err != nil {
		return utils.SendError(err, http.StatusInternalServerError)
	}

	return utils.SendData(nil, "Berhasil menduplikasi template formulir pertanyaan")
}

func (service *templateFormulirPertanyaanService) DeleteTemplateFormulirPertanyaan(usr models.JwtCustomClaims, req map[string]interface{}, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()

	codeStr, ok := slug["template_formulir_pertanyaan_code"].(string)
	if !ok {
		return utils.SendError(errors.New("Kode template formulir pertanyaan tidak valid"), http.StatusBadRequest)
	}

	err := service.templateFormulirPertanyaanRepo.DeleteFormTransactionByCode(codeStr)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return utils.SendError(errors.New("Template formulir pertanyaan tidak ditemukan"), http.StatusNotFound)
		}
		return utils.SendError(err, http.StatusInternalServerError)
	}

	return utils.SendData(nil, "Berhasil menghapus template formulir pertanyaan")
}

func (service *templateFormulirPertanyaanService) GetListTemplateFormulirPertanyaan(usr models.JwtCustomClaims, req map[string]interface{}, slug map[string]interface{}) (*pb.ProxyResponse, error) {
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

	data, totalData, err := service.templateFormulirPertanyaanRepo.GetListTemplateFormulirPertanyaan(payload)
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

	return utils.SendData(result, "Berhasil mengambil list template formulir pertanyaan")
}

func (s *templateFormulirPertanyaanService) GetQuestionTypeOptions(usr models.JwtCustomClaims, param url.Values) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()

	page, err := strconv.Atoi(param.Get("page"))
	if err != nil || page <= 0 {
		page = 1
	}

	limit, err := strconv.Atoi(param.Get("limit"))
	if err != nil || limit <= 0 {
		limit = 1000
	}

	// 1. Definisikan sumber data dari Enum
	allOptions := []response.EnumOption{
		{ID: string(enums.LONG_ANSWER), Name: "Teks"},
		{ID: string(enums.NUMBER), Name: "Angka"},
		{ID: string(enums.MULTIPLE_CHOICES), Name: "Pilihan Ganda"},
		{ID: string(enums.IMAGE_TEMPLATE), Name: "Gambar"},
		{ID: string(enums.MAPS), Name: "Lokasi"},
	}

	// 2. Filter berdasarkan ID (jika ada input dari param id/id[])
	var rawIDs []string
	if len(param["id[]"]) > 0 {
		rawIDs = param["id[]"]
	} else if len(param["id"]) > 0 {
		rawIDs = param["id"]
	}

	var filteredOptions []response.EnumOption
	if len(rawIDs) > 0 {
		idMap := make(map[string]bool)
		for _, id := range rawIDs {
			idMap[id] = true
		}
		for _, opt := range allOptions {
			if idMap[opt.ID] {
				filteredOptions = append(filteredOptions, opt)
			}
		}
	} else {
		filteredOptions = allOptions
	}

	// 3. Filter berdasarkan pencarian kata kunci (q)
	q := strings.ToLower(param.Get("q"))
	fmt.Println(q)
	var searchedOptions []response.EnumOption
	if q != "" {
		for _, opt := range filteredOptions {
			// Mencari kecocokan pada Nama atau ID
			if strings.Contains(strings.ToLower(opt.Name), q) || strings.Contains(strings.ToLower(opt.ID), q) {
				searchedOptions = append(searchedOptions, opt)
			}
		}
	} else {
		searchedOptions = filteredOptions
	}

	// 4. Pagination (In-Memory)
	totalData := int64(len(searchedOptions))
	startIndex := (page - 1) * limit
	endIndex := startIndex + limit

	var paginatedOptions []response.EnumOption
	if startIndex < len(searchedOptions) {
		if endIndex > len(searchedOptions) {
			endIndex = len(searchedOptions)
		}
		paginatedOptions = searchedOptions[startIndex:endIndex]
	} else {
		paginatedOptions = []response.EnumOption{}
	}

	// 6. Hitung status hasMore dan bentuk Response
	currentTotalLoaded := startIndex + len(paginatedOptions)
	hasMore := int64(currentTotalLoaded) < totalData

	responseData := response.OptionsStringIdResponse{
		Options: paginatedOptions,
		Meta: response.PaginationMeta{
			CurrentPage: page,
			PerPage:     limit,
			Total:       totalData,
			HasMore:     hasMore,
		},
	}

	return utils.SendData(responseData, "Berhasil mengambil opsi tipe pertanyaan")
}

func (s *templateFormulirPertanyaanService) GetFormulirPertanyaanOptions(param url.Values) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()

	page, err := strconv.Atoi(param.Get("page"))
	if err != nil || page <= 0 {
		page = 1
	}

	limit, err := strconv.Atoi(param.Get("limit"))
	if err != nil || limit <= 0 {
		limit = 1000
	}

	var formulirPertanyaanIDs []string
	if len(param["id[]"]) > 0 {
		formulirPertanyaanIDs = param["id[]"]
	} else if len(param["id"]) > 0 {
		formulirPertanyaanIDs = param["id"]
	}

	var parsedIDs []string
	for _, formuliPertanyaanID := range formulirPertanyaanIDs {
		parsedIDs = append(parsedIDs, formuliPertanyaanID)
	}

	_req := payloads.FormulirPertanyaanOptionsPayload{
		Q:     param.Get("q"),
		Page:  page,
		Limit: limit,
		IDs:   parsedIDs,
	}

	data, totalData, err := s.templateFormulirPertanyaanRepo.GetFormulirPertanyaanOptions(_req)
	if err != nil {
		return utils.SendError(err, http.StatusInternalServerError)
	}

	currentTotalLoaded := (page-1)*limit + len(data)
	hasMore := int64(currentTotalLoaded) < totalData

	responseData := response.StringOptionsResponse{
		Options: data,
		Meta: response.PaginationMeta{
			CurrentPage: page,
			PerPage:     limit,
			Total:       totalData,
			HasMore:     hasMore,
		},
	}

	return utils.SendData(responseData, "Berhasil mengambil opsi formulir pertanyaan")
}

func (s *templateFormulirPertanyaanService) GetPertanyaanOptions(param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()

	codeStr := slug["form_code"]
	code, ok := codeStr.(string)
	if !ok {
		return utils.SendError(errors.New("Kode template formulir pertanyaan tidak valid"), http.StatusBadRequest)
	}

	_, err := s.templateFormulirPertanyaanRepo.GetFormByCode(code)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return utils.SendError(errors.New("Template formulir pertanyaan tidak ditemukan"), http.StatusNotFound)
		}
		return utils.SendError(err, http.StatusInternalServerError)
	}

	page, err := strconv.Atoi(param.Get("page"))
	if err != nil || page <= 0 {
		page = 1
	}

	limit, err := strconv.Atoi(param.Get("limit"))
	if err != nil || limit <= 0 {
		limit = 1000
	}

	var formulirPertanyaanIDs []string
	if len(param["id[]"]) > 0 {
		formulirPertanyaanIDs = param["id[]"]
	} else if len(param["id"]) > 0 {
		formulirPertanyaanIDs = param["id"]
	}

	var parsedIDs []string
	for _, formuliPertanyaanID := range formulirPertanyaanIDs {
		parsedIDs = append(parsedIDs, formuliPertanyaanID)
	}

	var excludePertanyaanIDs []string
	if len(param["exclude_id[]"]) > 0 {
		excludePertanyaanIDs = param["exclude_id[]"]
	} else if len(param["exclude_id"]) > 0 {
		excludePertanyaanIDs = param["exclude_id"]
	}

	var parsedExcludeIDs []string
	for _, excludeID := range excludePertanyaanIDs {
		parsedExcludeIDs = append(parsedExcludeIDs, excludeID)
	}

	typeQuestionStr := param.Get("type")

	if typeQuestionStr != "" && !enums.QuestionType(param.Get("type")).IsQuestionTypeValid() {
		return utils.SendError(errors.New("Tipe pertanyaan tidak valid"), http.StatusBadRequest)
	}

	_req := payloads.PertanyaanOptionsPayload{
		Q:          param.Get("q"),
		Page:       page,
		Limit:      limit,
		IDs:        parsedIDs,
		ExcludeIDs: parsedExcludeIDs,
		Type:       &typeQuestionStr,
	}

	data, totalData, err := s.templateFormulirPertanyaanRepo.GetPertanyaanOptionsByFormCode(_req, code)
	if err != nil {
		return utils.SendError(err, http.StatusInternalServerError)
	}

	currentTotalLoaded := (page-1)*limit + len(data)
	hasMore := int64(currentTotalLoaded) < totalData

	responseData := response.FormFieldOptionsResponse{
		Options: data,
		Meta: response.PaginationMeta{
			CurrentPage: page,
			PerPage:     limit,
			Total:       totalData,
			HasMore:     hasMore,
		},
	}

	return utils.SendData(responseData, "Berhasil mengambil opsi Pertanyaan pada template formulir")
}

func (s *templateFormulirPertanyaanService) DetailPertanyaan(usr models.JwtCustomClaims, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()

	StrId := slug["id"]
	Id, err := utils.ToInt64(StrId)
	if err != nil {
		return utils.SendError(err, http.StatusBadRequest)
	}

	data, err := s.templateFormulirPertanyaanRepo.GetPertanyaanById(int(Id))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return utils.SendError(errors.New("Pertanyaan tidak ditemukan"), http.StatusNotFound)
		}
		return utils.SendError(err, http.StatusInternalServerError)
	}

	return utils.SendData(data, "Berhasil mendapatkan detail pertanyaan")
}

func (s *templateFormulirPertanyaanService) GetMultipleChoiceOptionByFormFieldId(param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()

	StrId := slug["form_field_id"]
	Id, err := utils.ToInt64(StrId)
	if err != nil {
		return utils.SendError(err, http.StatusBadRequest)
	}

	isQuestionExists, err := s.templateFormulirPertanyaanRepo.IsQuestionExistsById(int(Id))
	if err != nil {
		return utils.SendError(err, http.StatusInternalServerError)
	}
	if !isQuestionExists {
		return utils.SendError(errors.New("Pertanyaan tidak ditemukan"), http.StatusNotFound)
	}

	page, err := strconv.Atoi(param.Get("page"))
	if err != nil || page <= 0 {
		page = 1
	}

	limit, err := strconv.Atoi(param.Get("limit"))
	if err != nil || limit <= 0 {
		limit = 1000
	}

	_req := payloads.MultipleChoiceOptionsPayload{
		Q:     param.Get("q"),
		Page:  page,
		Limit: limit,
	}

	data, totalData, err := s.templateFormulirPertanyaanRepo.GetMultipleChoiceOptions(_req, Id)
	if err != nil {
		return utils.SendError(err, http.StatusInternalServerError)
	}

	currentTotalLoaded := (page-1)*limit + len(data)
	hasMore := int64(currentTotalLoaded) < totalData

	responseData := response.OptionsResponse{
		Options: data,
		Meta: response.PaginationMeta{
			CurrentPage: page,
			PerPage:     limit,
			Total:       totalData,
			HasMore:     hasMore,
		},
	}

	return utils.SendData(responseData, "Berhasil mendapatkan opsi pilihan ganda")
}
