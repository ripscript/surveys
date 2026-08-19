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
	"math"
	"net/http"
	"net/url"
	"strconv"

	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"
)

type ArtikelCategoryService interface {
	CreateCategory(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	DeleteCategory(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	DetailCategory(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	UpdateCategory(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	GetCategoryOptions(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	GetList(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	PublicList(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
}

type artikelCategoryService struct {
	artikelCategoryRepo repository.ArtikelCategoryRepository
}

func NewArtikelCategoryService(artikelCategoryRepo repository.ArtikelCategoryRepository) ArtikelCategoryService {
	return &artikelCategoryService{artikelCategoryRepo: artikelCategoryRepo}
}

func (service *artikelCategoryService) CreateCategory(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()

	var payload payloads.ArtikelCategoryPayload
	err := utils.DynamicBind(req, &payload)
	if err != nil {
		return utils.SendError(err, http.StatusBadRequest)
	}

	validate := validator.New()
	err = validate.Struct(payload)
	if err != nil {
		for _, err := range err.(validator.ValidationErrors) {
			customErrorMsg := utils.TranslateError(err)
			return utils.SendError(errors.New(customErrorMsg), http.StatusBadRequest)
		}
	}

	getByName, err := service.artikelCategoryRepo.GetByName(ctx, payload.Nama)
	if err != nil {
		if err.Error() != gorm.ErrRecordNotFound.Error() {
			return utils.SendError(errors.New("terjadi kesalahan pada server, silahkan coba lagi nanti"), http.StatusInternalServerError)
		}
	}

	if getByName != nil {
		return utils.SendError(errors.New("kategori artikel dengan nama tersebut sudah ada"), http.StatusBadRequest)
	}

	createData := models.ArtikelCategory{
		Name: payload.Nama,
	}

	createdData, err := service.artikelCategoryRepo.Create(ctx, &createData)
	if err != nil {
		return utils.SendError(err, http.StatusInternalServerError)
	}

	return utils.SendData(createdData, "Berhasil membuat kategori artikel")
}

func (service *artikelCategoryService) DetailCategory(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()
	strId := slug["id"]
	id, err := utils.ToInt(strId)
	if err != nil {
		return utils.SendError(err, http.StatusBadRequest)
	}

	getByID, err := service.artikelCategoryRepo.GetByID(ctx, id)
	if err != nil {
		if err.Error() != gorm.ErrRecordNotFound.Error() {
			return utils.SendError(errors.New("terjadi kesalahan pada server, silahkan coba lagi nanti"), http.StatusInternalServerError)
		}
	}

	if getByID == nil {
		return utils.SendError(errors.New("kategori artikel tidak ditemukan"), http.StatusNotFound)
	}

	return utils.SendData(getByID, "Kategori Berhasil dihapus")
}

func (service *artikelCategoryService) DeleteCategory(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()
	strId := slug["id"]
	id, err := utils.ToInt(strId)
	if err != nil {
		return utils.SendError(err, http.StatusBadRequest)
	}

	getByID, err := service.artikelCategoryRepo.GetByID(ctx, id)
	if err != nil {
		if err.Error() != gorm.ErrRecordNotFound.Error() {
			return utils.SendError(errors.New("terjadi kesalahan pada server, silahkan coba lagi nanti"), http.StatusInternalServerError)
		}
	}

	if getByID == nil {
		return utils.SendError(errors.New("kategori artikel tidak ditemukan"), http.StatusNotFound)
	}

	err = service.artikelCategoryRepo.Delete(ctx, id)
	if err != nil {
		return utils.SendError(err, http.StatusInternalServerError)
	}

	return utils.SendData(nil, "Kategori Berhasil dihapus")
}

func (service *artikelCategoryService) UpdateCategory(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()

	strId := slug["id"]
	id, err := utils.ToInt(strId)
	if err != nil {
		return utils.SendError(err, http.StatusBadRequest)
	}

	getByID, err := service.artikelCategoryRepo.GetByID(ctx, id)
	if err != nil {
		if err.Error() != gorm.ErrRecordNotFound.Error() {
			return utils.SendError(errors.New("terjadi kesalahan pada server, silahkan coba lagi nanti"), http.StatusInternalServerError)
		}
	}

	if getByID == nil {
		return utils.SendError(errors.New("kategori artikel tidak ditemukan"), http.StatusNotFound)
	}

	var payload payloads.ArtikelCategoryPayload
	err = utils.DynamicBind(req, &payload)
	if err != nil {
		return utils.SendError(err, http.StatusBadRequest)
	}

	validate := validator.New()
	err = validate.Struct(payload)
	if err != nil {
		for _, err := range err.(validator.ValidationErrors) {
			customErrorMsg := utils.TranslateError(err)
			return utils.SendError(errors.New(customErrorMsg), http.StatusBadRequest)
		}
	}

	if getByID.Name != payload.Nama {
		getByName, err := service.artikelCategoryRepo.GetByName(ctx, payload.Nama)
		if err != nil {
			if err.Error() != gorm.ErrRecordNotFound.Error() {
				return utils.SendError(errors.New("terjadi kesalahan pada server, silahkan coba lagi nanti"), http.StatusInternalServerError)
			}
		}

		if getByName != nil {
			return utils.SendError(errors.New("kategori artikel dengan nama tersebut sudah ada"), http.StatusBadRequest)
		}
	}

	getByID.Name = payload.Nama

	updatedData, err := service.artikelCategoryRepo.Update(ctx, getByID)
	if err != nil {
		return utils.SendError(err, http.StatusInternalServerError)
	}

	return utils.SendData(updatedData, "Berhasil memperbarui kategori artikel")
}

func (service *artikelCategoryService) GetCategoryOptions(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
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

	_req := payloads.ArtikelCategoryOptionsPayload{
		Q:     param.Get("q"),
		Page:  page,
		Limit: limit,
		IDs:   parsedIDs,
	}

	data, totalData, err := service.artikelCategoryRepo.GetOptions(_req)
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

	return utils.SendData(responseData, "Berhasil mengambil opsi kategori artikel")
}

func (service *artikelCategoryService) GetList(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
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

	data, totalData, err := service.artikelCategoryRepo.GetList(payload)
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

	return utils.SendData(result, "Berhasil mengambil list kategori artikel")
}

func (service *artikelCategoryService) PublicList(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()

	data, err := service.artikelCategoryRepo.GetPublicList()
	if err != nil {
		return utils.SendError(err, http.StatusInternalServerError)
	}

	return utils.SendData(data, "Berhasil mengambil list kategori artikel")
}
