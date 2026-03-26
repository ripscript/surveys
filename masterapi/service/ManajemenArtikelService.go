package service

import (
	"backend/masterapi/models"
	"backend/masterapi/payloads"
	"backend/masterapi/repository"
	"backend/masterapi/utils"
	"backend/siccore/pb"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"
)

type ManajemenArtikelService interface {
	CreateKategoriArtikel(usr models.JwtCustomClaims, req map[string]interface{}) (*pb.ProxyResponse, error)
	UpdateKategoriArtikel(usr models.JwtCustomClaims, req map[string]interface{}, slug map[string]interface{}) (*pb.ProxyResponse, error)
	DeleteKategoriArtikel(usr models.JwtCustomClaims, slug map[string]interface{}) (*pb.ProxyResponse, error)
	GetKategoriArtikel(usr models.JwtCustomClaims, slug map[string]interface{}) (*pb.ProxyResponse, error)
	GetListKategoriArtikel(req map[string]interface{}, slug map[string]interface{}) (*pb.ProxyResponse, error)
}

type manajemenArtikelService struct {
	manajemenArtikelRepo repository.ManajemenArtikelRepo
}

func NewManajemenArtikelService(
	manajemenArtikelRepo repository.ManajemenArtikelRepo,
) ManajemenArtikelService {
	return &manajemenArtikelService{
		manajemenArtikelRepo,
	}
}

func (service *manajemenArtikelService) CreateKategoriArtikel(usr models.JwtCustomClaims, req map[string]interface{}) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()

	var payload payloads.KategoriArtikelPayload

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

	kategori, err := service.manajemenArtikelRepo.GetKategoriArtikelByName(payload.Nama)
	if err != nil {
		if err.Error() != gorm.ErrRecordNotFound.Error() {
			return utils.SendError(err, http.StatusInternalServerError)
		}
	}

	if kategori != nil {
		return utils.SendError(errors.New("Kategori artikel dengan nama tersebut sudah ada"), http.StatusBadRequest)
	}

	artikelKategori := models.ArtikelKategori{
		Name: payload.Nama,
	}

	newKategori, err := service.manajemenArtikelRepo.CreateKategoriArtikel(artikelKategori)
	if err != nil {
		return utils.SendError(err, http.StatusInternalServerError)
	}

	return utils.SendData(newKategori, "Berhasil create data")
}

func (service *manajemenArtikelService) UpdateKategoriArtikel(usr models.JwtCustomClaims, req map[string]interface{}, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()

	var payload payloads.KategoriArtikelPayload

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

	idStr, ok := slug["id"].(string)
	if !ok {
		return utils.SendError(errors.New("Invalid ID in slug"), http.StatusBadRequest)
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return utils.SendError(errors.New("gagal parse string ke int64"), http.StatusBadRequest)
	}

	kategori, err := service.manajemenArtikelRepo.GetKategoriArtikelById(int64(id))
	if err != nil {
		if err.Error() == gorm.ErrRecordNotFound.Error() {
			return utils.SendError(errors.New("Kategori artikel tidak ditemukan"), http.StatusNotFound)
		}
		return utils.SendError(err, http.StatusInternalServerError)
	}

	if kategori == nil {
		return utils.SendError(errors.New("Kategori artikel tidak ditemukan"), http.StatusNotFound)
	}

	if kategori.Name != payload.Nama {
		existingKategori, err := service.manajemenArtikelRepo.GetKategoriArtikelByName(payload.Nama)
		if err != nil {
			if err.Error() != gorm.ErrRecordNotFound.Error() {
				return utils.SendError(err, http.StatusInternalServerError)
			}
		}

		if existingKategori != nil {
			return utils.SendError(errors.New("Kategori artikel dengan nama tersebut sudah ada"), http.StatusBadRequest)
		}
	}

	kategori.Name = payload.Nama

	updatedKategori, err := service.manajemenArtikelRepo.UpdateKategoriArtikel(*kategori)
	if err != nil {
		return utils.SendError(err, http.StatusInternalServerError)
	}

	return utils.SendData(updatedKategori, "Berhasil update data")
}

func (service *manajemenArtikelService) DeleteKategoriArtikel(usr models.JwtCustomClaims, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()

	idStr, ok := slug["id"].(string)
	if !ok {
		return utils.SendError(errors.New("Invalid ID in slug"), http.StatusBadRequest)
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return utils.SendError(errors.New("gagal parse string ke int64"), http.StatusBadRequest)
	}

	kategori, err := service.manajemenArtikelRepo.GetKategoriArtikelById(int64(id))
	if err != nil {
		if err.Error() == gorm.ErrRecordNotFound.Error() {
			return utils.SendError(errors.New("Kategori artikel tidak ditemukan"), http.StatusNotFound)
		}
		return utils.SendError(err, http.StatusInternalServerError)
	}

	if kategori == nil {
		return utils.SendError(errors.New("Kategori artikel tidak ditemukan"), http.StatusNotFound)
	}

	err = service.manajemenArtikelRepo.DeleteKategoriArtikelById(int64(id))
	if err != nil {
		return utils.SendError(err, http.StatusInternalServerError)
	}

	return utils.SendData(nil, "Berhasil delete data")
}

func (service *manajemenArtikelService) GetKategoriArtikel(usr models.JwtCustomClaims, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()

	idStr, ok := slug["id"].(string)
	if !ok {
		return utils.SendError(errors.New("Invalid ID in slug"), http.StatusBadRequest)
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return utils.SendError(errors.New("gagal parse string ke int64"), http.StatusBadRequest)
	}

	kategori, err := service.manajemenArtikelRepo.GetKategoriArtikelById(id)
	if err != nil {
		return utils.SendError(err, http.StatusInternalServerError)
	}

	return utils.SendData(kategori, "Berhasil get data")
}

func (service *manajemenArtikelService) GetListKategoriArtikel(req map[string]interface{}, slug map[string]interface{}) (*pb.ProxyResponse, error) {
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

	data, totalData, err := service.manajemenArtikelRepo.GetListKategoriArtikel(payload)
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

	return utils.SendData(result, "Berhasil mengambil list kategori artikel")
}
