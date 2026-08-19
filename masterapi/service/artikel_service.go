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

type ArtikelService interface {
	Create(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	Update(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
}

type artikelService struct {
	artikelRepo         repository.ArtikelRepository
	artikelCategoryRepo repository.ArtikelCategoryRepository
}

func NewArtikelService(
	artikelRepo repository.ArtikelRepository,
	artikelCategoryRepo repository.ArtikelCategoryRepository,
) ArtikelService {
	return &artikelService{
		artikelRepo:         artikelRepo,
		artikelCategoryRepo: artikelCategoryRepo,
	}
}

func (service *artikelService) Create(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()

	var payload payloads.ArtikelPayload
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

	getByName, err := service.artikelRepo.GetByName(ctx, payload.Judul)
	if err != nil {
		if err.Error() != gorm.ErrRecordNotFound.Error() {
			return utils.SendError(errors.New("terjadi kesalahan pada server, silahkan coba lagi nanti"), http.StatusInternalServerError)
		}
	}

	if getByName != nil {
		return utils.SendError(errors.New("Artikel dengan nama tersebut sudah ada"), http.StatusBadRequest)
	}

	kategori, err := service.artikelCategoryRepo.GetByID(ctx, int(payload.Kategori))
	if err != nil {
		if err.Error() != gorm.ErrRecordNotFound.Error() {
			return utils.SendError(errors.New("terjadi kesalahan pada server, silahkan coba lagi nanti"), http.StatusInternalServerError)
		}
	}

	if kategori == nil {
		return utils.SendError(errors.New("Kategori artikel tidak ditemukan"), http.StatusBadRequest)
	}

	createData := models.Artikel{
		Judul:             &payload.Judul,
		ArtikelCategoryID: utils.Int64ToPointer(int64(kategori.ID)),
	}

	createdData, err := service.artikelRepo.Create(ctx, &createData)
	if err != nil {
		return utils.SendError(err, http.StatusInternalServerError)
	}

	return utils.SendData(createdData, "Berhasil membuat artikel")
}

func (service *artikelService) Update(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()

	var payload payloads.ArtikelPayload
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

	getByName, err := service.artikelRepo.GetByName(ctx, payload.Judul)
	if err != nil {
		if err.Error() != gorm.ErrRecordNotFound.Error() {
			return utils.SendError(errors.New("terjadi kesalahan pada server, silahkan coba lagi nanti"), http.StatusInternalServerError)
		}
	}

	if getByName != nil {
		return utils.SendError(errors.New("Artikel dengan nama tersebut sudah ada"), http.StatusBadRequest)
	}

	kategori, err := service.artikelCategoryRepo.GetByID(ctx, int(payload.Kategori))
	if err != nil {
		if err.Error() != gorm.ErrRecordNotFound.Error() {
			return utils.SendError(errors.New("terjadi kesalahan pada server, silahkan coba lagi nanti"), http.StatusInternalServerError)
		}
	}

	if kategori == nil {
		return utils.SendError(errors.New("Kategori artikel tidak ditemukan"), http.StatusBadRequest)
	}

	createData := models.Artikel{
		Judul:             &payload.Judul,
		ArtikelCategoryID: utils.Int64ToPointer(int64(kategori.ID)),
	}

	createdData, err := service.artikelRepo.Create(ctx, &createData)
	if err != nil {
		return utils.SendError(err, http.StatusInternalServerError)
	}

	return utils.SendData(createdData, "Berhasil membuat artikel")
}
