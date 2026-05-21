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
	"errors"
	"math"
	"net/http"
	"net/url"
	"strconv"

	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"
)

type TemplateUcapanService interface {
	GetGeneralTemplate(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	GetTemplateUcapanOptions(param url.Values) (*pb.ProxyResponse, error)
	CreateTemplateUcapan(usr models.JwtCustomClaims, req map[string]interface{}) (*pb.ProxyResponse, error)
	UpdateTemplateUcapan(usr models.JwtCustomClaims, req map[string]interface{}, slug map[string]interface{}) (*pb.ProxyResponse, error)
	DeleteTemplateUcapan(usr models.JwtCustomClaims, slug map[string]interface{}) (*pb.ProxyResponse, error)
	GetListTemplateUcapan(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	GetUcapanOptions(param url.Values) (*pb.ProxyResponse, error)
}

type templateUcapanService struct {
	templateUcapanRepo repository.TemplateUcapanRepo
	userRepo           repository.UserRepo
}

func NewTemplateUcapanService(
	templateUcapanRepo repository.TemplateUcapanRepo,
	userRepo repository.UserRepo,
) TemplateUcapanService {
	return &templateUcapanService{
		templateUcapanRepo: templateUcapanRepo,
		userRepo:           userRepo,
	}
}

func (service *templateUcapanService) GetGeneralTemplate(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()
	StrId := slug["template_ucapan_id"]
	Id, err := utils.ToInt64(StrId)
	if err != nil {
		return utils.SendError(err, http.StatusBadRequest)
	}

	data, err := service.templateUcapanRepo.GetTemplateUcapanById(int(Id))
	if err != nil {
		return utils.SendError(err, http.StatusInternalServerError)
	}

	respondent, err := service.userRepo.GetRespondentDetailById(ctx, usr.RespondentID)
	if err != nil {
		return utils.SendError(err, http.StatusInternalServerError)
	}

	content := data.Content

	if rawOpening, errOp := service.templateUcapanRepo.GetTemplateUcapanById(data.ID); errOp == nil && rawOpening != nil {

		openingStr := utils.ReplaceStringRespondentVariable(rawOpening.Content, respondent)
		content = openingStr
	}

	data.Preview = content

	return utils.SendData(data, "Berhasil mengambil data")
}

func (s *templateUcapanService) GetTemplateUcapanOptions(param url.Values) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()

	page, err := strconv.Atoi(param.Get("page"))
	if err != nil || page <= 0 {
		page = 1
	}

	limit, err := strconv.Atoi(param.Get("limit"))
	if err != nil || limit <= 0 {
		limit = 1000
	}

	var rawIDs []string
	if len(param["id[]"]) > 0 {
		rawIDs = param["id[]"]
	} else if len(param["id"]) > 0 {
		rawIDs = param["id"]
	}

	var parsedIDs []int64
	for _, rawID := range rawIDs {
		if id, err := strconv.ParseInt(rawID, 10, 64); err == nil {
			parsedIDs = append(parsedIDs, id)
		}
	}

	kelurahanId, err := strconv.Atoi(param.Get("kelurahan_id"))
	if err != nil || kelurahanId <= 0 {
		kelurahanId = 0
	}

	_req := payloads.TemplateUcapanOptionsPayload{
		Q:     param.Get("q"),
		Page:  page,
		Limit: limit,
		IDs:   parsedIDs,
	}

	data, totalData, err := s.templateUcapanRepo.GetVariableTemplateUcapanOptions(_req)
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

	return utils.SendData(responseData, "Berhasil mengambil opsi variable template ucapan")
}

func (service *templateUcapanService) CreateTemplateUcapan(usr models.JwtCustomClaims, req map[string]interface{}) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()

	var payload payloads.TemplateUcapanPayload

	err := utils.DynamicBind(req, &payload)
	if err != nil {
		return utils.SendError(err, http.StatusBadRequest)
	}

	var validate = validator.New()

	validate.RegisterValidation("is_type_template", func(fl validator.FieldLevel) bool {
		// Ambil value dari field
		val := fl.Field().String()
		// Konversi ke enum dan cek validasinya
		return enums.TypeTemplateUcapan(val).IsValid()
	})

	err = validate.Struct(payload)
	if err != nil {
		customErrorMsg := utils.FormatValidationError(err)
		return utils.SendError(errors.New(customErrorMsg), http.StatusBadRequest)
	}

	templateUcapan := models.GeneralTemplate{
		Type:      payload.Tipe,
		Name:      payload.NamaTemplate,
		Content:   payload.Konten,
		CreatedBy: 1,
	}

	existingTemplateByName, err := service.templateUcapanRepo.GetTemplateUcapanByName(payload.NamaTemplate)

	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return utils.SendError(err, http.StatusInternalServerError)
	}

	if existingTemplateByName != nil {
		return utils.SendError(errors.New("Nama template ucapan sudah digunakan, silakan gunakan nama lain"), http.StatusBadRequest)
	}

	createdData, err := service.templateUcapanRepo.CreateTemplateUcapan(templateUcapan)
	if err != nil {
		return utils.SendError(err, http.StatusInternalServerError)
	}

	return utils.SendData(createdData, "Berhasil membuat template ucapan baru")
}

func (service *templateUcapanService) UpdateTemplateUcapan(usr models.JwtCustomClaims, req map[string]interface{}, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()

	StrId := slug["template_ucapan_id"]
	Id, err := utils.ToInt64(StrId)
	if err != nil {
		return utils.SendError(err, http.StatusBadRequest)
	}

	var payload payloads.TemplateUcapanPayload

	err = utils.DynamicBind(req, &payload)
	if err != nil {
		return utils.SendError(err, http.StatusBadRequest)
	}

	var validate = validator.New()

	validate.RegisterValidation("is_type_template", func(fl validator.FieldLevel) bool {
		val := fl.Field().String()
		return enums.TypeTemplateUcapan(val).IsValid()
	})

	err = validate.Struct(payload)
	if err != nil {
		customErrorMsg := utils.FormatValidationError(err)
		return utils.SendError(errors.New(customErrorMsg), http.StatusBadRequest)
	}

	exisitingTemplateUcapan, err := service.templateUcapanRepo.GetTemplateUcapanById(int(Id))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return utils.SendError(errors.New("Template ucapan tidak ditemukan"), http.StatusNotFound)
		}
		return utils.SendError(err, http.StatusInternalServerError)
	}

	if exisitingTemplateUcapan.Name != payload.NamaTemplate {
		existingTemplateByName, err := service.templateUcapanRepo.GetTemplateUcapanByName(payload.NamaTemplate)

		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return utils.SendError(err, http.StatusInternalServerError)
		}

		if existingTemplateByName != nil {
			return utils.SendError(errors.New("Nama template ucapan sudah digunakan, silakan gunakan nama lain"), http.StatusBadRequest)
		}
	}

	templateUcapan := models.GeneralTemplate{
		ID:        exisitingTemplateUcapan.ID,
		Type:      payload.Tipe,
		Name:      payload.NamaTemplate,
		Content:   payload.Konten,
		CreatedAt: exisitingTemplateUcapan.CreatedAt,
		CreatedBy: 1,
	}

	isUsed, err := service.templateUcapanRepo.IsUsedTemplateUcapan(int64(exisitingTemplateUcapan.ID))
	if err != nil {
		return utils.SendError(err, http.StatusInternalServerError)
	}

	if isUsed {
		return utils.SendError(errors.New("Template ucapan tidak dapat diubah karena sedang digunakan"), http.StatusBadRequest)
	}

	updatedData, err := service.templateUcapanRepo.UpdateTemplateUcapan(templateUcapan)
	if err != nil {
		return utils.SendError(err, http.StatusInternalServerError)
	}

	return utils.SendData(updatedData, "Berhasil memperbarui template ucapan")
}

func (service *templateUcapanService) DeleteTemplateUcapan(usr models.JwtCustomClaims, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()

	StrId := slug["template_ucapan_id"]
	Id, err := utils.ToInt64(StrId)
	if err != nil {
		return utils.SendError(err, http.StatusBadRequest)
	}

	_, err = service.templateUcapanRepo.GetTemplateUcapanById(int(Id))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return utils.SendError(errors.New("Template ucapan tidak ditemukan"), http.StatusNotFound)
		}
		return utils.SendError(err, http.StatusInternalServerError)
	}

	isUsed, err := service.templateUcapanRepo.IsUsedTemplateUcapan(Id)
	if err != nil {
		return utils.SendError(err, http.StatusInternalServerError)
	}

	if isUsed {
		return utils.SendError(errors.New("Template ucapan tidak dapat dihapus karena sedang digunakan"), http.StatusBadRequest)
	}

	err = service.templateUcapanRepo.DeleteTemplateUcapan(int(Id))
	if err != nil {
		return utils.SendError(err, http.StatusInternalServerError)
	}

	return utils.SendData(nil, "Berhasil menghapus template ucapan")
}

func (service *templateUcapanService) GetListTemplateUcapan(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()

	search := param.Get("search")
	page, _ := strconv.Atoi(param.Get("page"))
	limit, _ := strconv.Atoi(param.Get("limit"))
	orderBy := param.Get("order_by")
	orderDir := param.Get("order_dir")

	payload := payloads.DatatablePayload{
		Search:   search,
		Page:     page,
		Limit:    limit,
		OrderBy:  orderBy,
		OrderDir: orderDir,
	}

	if payload.Page <= 0 {
		payload.Page = 1
	}
	if payload.Limit <= 0 {
		payload.Limit = 5
	}

	data, totalData, err := service.templateUcapanRepo.GetListTemplateUcapan(payload)
	if err != nil {
		return utils.SendError(err, http.StatusInternalServerError)
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

	return utils.SendData(result, "Berhasil mengambil list template ucapan")
}

func (service *templateUcapanService) GetUcapanOptions(param url.Values) (*pb.ProxyResponse, error) {
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

	typeTemplateStr := param.Get("type")
	if typeTemplateStr != "" && !enums.TypeTemplateUcapan(typeTemplateStr).IsValid() {
		return utils.SendError(errors.New("Tipe template ucapan tidak valid"), http.StatusBadRequest)
	}

	typeTemplate := enums.StringToTypeTemplateUcapan(typeTemplateStr)

	_req := payloads.UcapanOptionsPayload{
		Q:     param.Get("q"),
		Page:  page,
		Limit: limit,
		IDs:   parsedIDs,
		Type:  &typeTemplate,
	}

	data, totalData, err := service.templateUcapanRepo.GetTemplateUcapanOption(_req)
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

	return utils.SendData(responseData, "Berhasil mengambil opsi Template Ucapan")
}
