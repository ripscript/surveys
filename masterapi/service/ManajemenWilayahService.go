package service

import (
	"backend/masterapi/models"
	"backend/masterapi/payloads"
	"backend/masterapi/repository"
	"backend/masterapi/response"
	"backend/masterapi/utils"
	"backend/siccore/pb"
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"
)

type ManajemenWilayahService interface {
	CreateKecamatan(usr models.JwtCustomClaims, req map[string]interface{}) (*pb.ProxyResponse, error)
	GetKecamatanDetail(slug map[string]interface{}) (*pb.ProxyResponse, error)
	UpdateKecamatan(usr models.JwtCustomClaims, req map[string]interface{}, slug map[string]interface{}) (*pb.ProxyResponse, error)
	GetListKecamatan(req map[string]interface{}) (*pb.ProxyResponse, error)
	GetKecamatanOptions(param url.Values) (*pb.ProxyResponse, error)
	DeleteKecamatan(slug map[string]interface{}) (*pb.ProxyResponse, error)

	CreateKelurahan(usr models.JwtCustomClaims, req map[string]interface{}) (*pb.ProxyResponse, error)
	GetKelurahanDetail(slug map[string]interface{}) (*pb.ProxyResponse, error)
	UpdateKelurahan(usr models.JwtCustomClaims, req map[string]interface{}, slug map[string]interface{}) (*pb.ProxyResponse, error)
	GetListKelurahan(req map[string]interface{}, slug map[string]interface{}) (*pb.ProxyResponse, error)
	GetKelurahanOptions(param url.Values) (*pb.ProxyResponse, error)
	DeleteKelurahan(slug map[string]interface{}) (*pb.ProxyResponse, error)

	GetRwDetail(slug map[string]interface{}) (*pb.ProxyResponse, error)
	UpdateRw(usr models.JwtCustomClaims, req map[string]interface{}, slug map[string]interface{}) (*pb.ProxyResponse, error)
	GetListRw(req map[string]interface{}, slug map[string]interface{}) (*pb.ProxyResponse, error)
	CreateRw(usr models.JwtCustomClaims, req map[string]interface{}) (*pb.ProxyResponse, error)
	GetRwOptions(param url.Values) (*pb.ProxyResponse, error)
	DeleteRw(slug map[string]interface{}) (*pb.ProxyResponse, error)

	GetRtDetail(slug map[string]interface{}) (*pb.ProxyResponse, error)
	UpdateRt(usr models.JwtCustomClaims, req map[string]interface{}, slug map[string]interface{}) (*pb.ProxyResponse, error)
	GetListRt(req map[string]interface{}, slug map[string]interface{}) (*pb.ProxyResponse, error)
	CreateRt(usr models.JwtCustomClaims, req map[string]interface{}) (*pb.ProxyResponse, error)
	GetRtOptions(param url.Values) (*pb.ProxyResponse, error)
	DeleteRT(slug map[string]interface{}) (*pb.ProxyResponse, error)
}

type manajemenWilayahService struct {
	manajemenWilayahRepo repository.ManajemenWilayahRepo
}

func NewManajemenWilayahService(
	manajemenWilayahRepo repository.ManajemenWilayahRepo,
) ManajemenWilayahService {
	return &manajemenWilayahService{
		manajemenWilayahRepo,
	}
}

func (service *manajemenWilayahService) CreateKecamatan(usr models.JwtCustomClaims, req map[string]interface{}) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()

	var payload payloads.CreateKecamatanPayload

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

	kecamatanByName, err := service.manajemenWilayahRepo.GetKecamatanByNameToLower(payload.NamaKecamatan)
	if err != nil && err.Error() != gorm.ErrRecordNotFound.Error() {
		return utils.SendError(err, http.StatusInternalServerError)
	}

	if kecamatanByName != nil {
		err := errors.New("Nama kecamatan sudah digunakan")
		return utils.SendError(err, http.StatusBadRequest)
	}

	slug := utils.StringToSlug(payload.NamaKecamatan, "-")

	kecamatan := models.Kecamatan{
		SubDistrictName: payload.NamaKecamatan,
		SubDistrictSlug: slug,
		KodeWilayah:     payload.KodeWilayah,
		Lat:             payload.Lat,
		Long:            payload.Long,
	}

	createData, err := service.manajemenWilayahRepo.CreateKecamatan(kecamatan)
	if err != nil {
		return utils.SendError(err, http.StatusInternalServerError)
	}

	return utils.SendData(createData, "Berhasil menambahkan kecamatan")
}

func (service *manajemenWilayahService) GetKecamatanDetail(slug map[string]interface{}) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()
	StrId := slug["kecamatan_id"]
	Id, err := utils.ToInt64(StrId)
	if err != nil {
		return utils.SendError(err, http.StatusBadRequest)
	}

	data, err := service.manajemenWilayahRepo.GetKecamatanByID(int(Id))
	if err != nil {
		return utils.SendError(err, http.StatusInternalServerError)
	}
	return utils.SendData(data, "Berhasil mengambil data")
}

func (service *manajemenWilayahService) UpdateKecamatan(usr models.JwtCustomClaims, req map[string]interface{}, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()

	var payload payloads.UpdateKecamatanPayload

	err := utils.DynamicBind(req, &payload)
	if err != nil {
		return utils.SendError(err, http.StatusBadRequest)
	}

	StrId := slug["kecamatan_id"]
	Id, err := utils.ToInt64(StrId)
	if err != nil {
		return utils.SendError(err, http.StatusBadRequest)
	}

	// if payload.NamaKecamatan == "" {
	// 	err := errors.New("Nama kecamatan tidak boleh kosong")
	// 	return utils.SendError(err, http.StatusBadRequest)
	// }

	getKecamatanById, err := service.manajemenWilayahRepo.GetKecamatanByID(int(Id))
	if err != nil {
		if err.Error() == gorm.ErrRecordNotFound.Error() {
			err := errors.New("Data kecamatan tidak ditemukan")
			return utils.SendError(err, http.StatusNotFound)
		}
		return utils.SendError(err, http.StatusInternalServerError)
	}

	newSlug := getKecamatanById.SubDistrictSlug

	// if getKecamatanById.SubDistrictName != payload.NamaKecamatan {
	// 	getKecamatanByName, err := service.manajemenWilayahRepo.GetKecamatanByName(payload.NamaKecamatan)
	// 	if err != nil && err.Error() != gorm.ErrRecordNotFound.Error() {
	// 		return utils.SendError(err, http.StatusInternalServerError)
	// 	}

	// 	if getKecamatanByName != nil {
	// 		err := errors.New("Nama kecamatan sudah digunakan")
	// 		return utils.SendError(err, http.StatusBadRequest)
	// 	}

	// 	newSlug = utils.StringToSlug(payload.NamaKecamatan, "-")
	// 	getKecamatanBySlug, err := service.manajemenWilayahRepo.GetKecamatanBySlug(newSlug)
	// 	if err != nil && err.Error() != gorm.ErrRecordNotFound.Error() {
	// 		return utils.SendError(err, http.StatusInternalServerError)
	// 	}

	// 	if getKecamatanBySlug != nil && getKecamatanBySlug.ID != Id {
	// 		newSlug = utils.StringToSlug(payload.NamaKecamatan+" "+strconv.FormatInt(time.Now().Unix(), 10), "-")
	// 	}
	// }

	var data = models.Kecamatan{
		ID: getKecamatanById.ID,
		// SubDistrictName: payload.NamaKecamatan,
		SubDistrictName: getKecamatanById.SubDistrictName,
		SubDistrictSlug: newSlug,
		KodeWilayah:     payload.KodeWilayah,
		Lat:             payload.Lat,
		Long:            payload.Long,
		CreatedAt:       getKecamatanById.CreatedAt,
	}

	_, err = service.manajemenWilayahRepo.UpdateKecamatan(data)
	if err != nil {
		return utils.SendError(err, http.StatusInternalServerError)
	}

	return utils.SendData(nil, "Berhasil update data")
}

func (service *manajemenWilayahService) GetListKecamatan(req map[string]interface{}) (*pb.ProxyResponse, error) {
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

	// Ambil data yang SUDAH digabung dari repo
	data, totalData, err := service.manajemenWilayahRepo.GetListKecamatan(payload)
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

	return utils.SendData(result, "Berhasil mengambil list kecamatan")
}

func (s *manajemenWilayahService) GetKecamatanOptions(param url.Values) (*pb.ProxyResponse, error) {
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

	_req := payloads.KecamatanOptionsPayload{
		Q:     param.Get("q"),
		Page:  page,
		Limit: limit,
		IDs:   parsedIDs,
	}

	data, totalData, err := s.manajemenWilayahRepo.GetKecamatanOptions(_req)
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

	return utils.SendData(responseData, "Berhasil mengambil opsi kecamatan")
}

func (service *manajemenWilayahService) DeleteKecamatan(slug map[string]interface{}) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()
	StrId := slug["kecamatan_id"]
	Id, err := utils.ToInt64(StrId)
	if err != nil {
		return utils.SendError(err, http.StatusBadRequest)
	}

	isUsed, err := service.manajemenWilayahRepo.IsKecamatanUsed(Id)
	if err != nil {
		return utils.SendError(err, http.StatusInternalServerError)
	}

	if isUsed {
		err := errors.New("Data tidak dapat dihapus karena masih digunakan oleh data lain.")
		return utils.SendError(err, http.StatusBadRequest)
	}

	err = service.manajemenWilayahRepo.DeleteKecamatanById(Id)
	if err != nil {
		return utils.SendError(err, http.StatusInternalServerError)
	}

	return utils.SendData(nil, "Berhasil menghapus kecamatan")
}

func (service *manajemenWilayahService) CreateKelurahan(usr models.JwtCustomClaims, req map[string]interface{}) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()

	var payload payloads.CreateKelurahanPayload

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

	kecamatan, err := service.manajemenWilayahRepo.GetKecamatanByID(payload.Kecamatan)
	if err != nil {
		if err.Error() == gorm.ErrRecordNotFound.Error() {
			err := errors.New("Kecamatan tidak ditemukan")
			return utils.SendError(err, http.StatusNotFound)
		}
		return utils.SendError(err, http.StatusInternalServerError)
	}

	kelurahanByName, err := service.manajemenWilayahRepo.GetKelurahanByNameToLower(payload.NamaKelurahan)
	if err != nil && err.Error() != gorm.ErrRecordNotFound.Error() {
		return utils.SendError(err, http.StatusInternalServerError)
	}

	if kelurahanByName != nil {
		err := errors.New("Nama kelurahan sudah digunakan")
		return utils.SendError(err, http.StatusBadRequest)
	}

	slug := utils.StringToSlug(payload.NamaKelurahan, "-")

	kelurahan := models.Kelurahan{
		SubDistrictId:     kecamatan.ID,
		VillageName:       payload.NamaKelurahan,
		VillageNameSlug:   slug,
		VillagePostalCode: payload.KodePos,
		KodeWilayah:       payload.KodeWilayah,
		Lat:               payload.Lat,
		Long:              payload.Long,
	}

	createData, err := service.manajemenWilayahRepo.CreateKelurahan(kelurahan)
	if err != nil {
		return utils.SendError(err, http.StatusInternalServerError)
	}

	return utils.SendData(createData, "Berhasil menambahkan kelurahan")
}

func (service *manajemenWilayahService) GetKelurahanDetail(slug map[string]interface{}) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()
	StrId := slug["kelurahan_id"]
	Id, err := utils.ToInt64(StrId)
	if err != nil {
		return utils.SendError(err, http.StatusBadRequest)
	}

	data, err := service.manajemenWilayahRepo.GetKelurahanByID(int(Id))
	if err != nil {
		return utils.SendError(err, http.StatusInternalServerError)
	}
	return utils.SendData(data, "Berhasil mengambil data")
}

func (service *manajemenWilayahService) UpdateKelurahan(usr models.JwtCustomClaims, req map[string]interface{}, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()

	var payload payloads.UpdateKelurahanPayload

	err := utils.DynamicBind(req, &payload)
	if err != nil {
		return utils.SendError(err, http.StatusBadRequest)
	}

	StrId := slug["kelurahan_id"]
	Id, err := utils.ToInt64(StrId)
	if err != nil {
		return utils.SendError(err, http.StatusBadRequest)
	}

	// if payload.NamaKelurahan == "" {
	// 	err := errors.New("Nama kelurahan tidak boleh kosong")
	// 	return utils.SendError(err, http.StatusBadRequest)
	// }

	getKelurahanById, err := service.manajemenWilayahRepo.GetKelurahanByID(int(Id))
	if err != nil {
		if err.Error() == gorm.ErrRecordNotFound.Error() {
			err := errors.New("Data kelurahan tidak ditemukan")
			return utils.SendError(err, http.StatusNotFound)
		}
		return utils.SendError(err, http.StatusInternalServerError)
	}

	newSlug := getKelurahanById.VillageNameSlug

	// if getKelurahanById.VillageName != payload.NamaKelurahan {
	// 	getKelurahanByName, err := service.manajemenWilayahRepo.GetKelurahanByName(payload.NamaKelurahan)
	// 	if err != nil && err.Error() != gorm.ErrRecordNotFound.Error() {
	// 		return utils.SendError(err, http.StatusInternalServerError)
	// 	}

	// 	if getKelurahanByName != nil {
	// 		err := errors.New("Nama kelurahan sudah digunakan")
	// 		return utils.SendError(err, http.StatusBadRequest)
	// 	}

	// 	newSlug = utils.StringToSlug(payload.NamaKelurahan, "-")
	// 	getKelurahanBySlug, err := service.manajemenWilayahRepo.GetKelurahanBySlug(newSlug)
	// 	if err != nil && err.Error() != gorm.ErrRecordNotFound.Error() {
	// 		return utils.SendError(err, http.StatusInternalServerError)
	// 	}

	// 	if getKelurahanBySlug != nil && getKelurahanBySlug.ID != Id {
	// 		newSlug = utils.StringToSlug(payload.NamaKelurahan+" "+strconv.FormatInt(time.Now().Unix(), 10), "-")
	// 	}
	// }

	var data = models.Kelurahan{
		ID:            getKelurahanById.ID,
		SubDistrictId: getKelurahanById.SubDistrictId,
		// VillageName:       payload.NamaKelurahan,
		VillageName:       getKelurahanById.VillageName,
		VillageNameSlug:   newSlug,
		VillagePostalCode: getKelurahanById.VillagePostalCode,
		KodeWilayah:       payload.KodeWilayah,
		Lat:               payload.Lat,
		Long:              payload.Long,
		CreatedAt:         getKelurahanById.CreatedAt,
	}

	_, err = service.manajemenWilayahRepo.UpdateKelurahan(data)
	if err != nil {
		return utils.SendError(err, http.StatusInternalServerError)
	}

	return utils.SendData(nil, "Berhasil update data")
}

func (service *manajemenWilayahService) GetListKelurahan(req map[string]interface{}, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()

	var payload payloads.DatatablePayload
	err := utils.DynamicBind(req, &payload)
	if err != nil {
		return utils.SendError(err, http.StatusBadRequest)
	}

	StrId := slug["kecamatan_id"]
	var kecamatanId *int64
	if StrId != nil {
		Id, err := utils.ToInt64(StrId)
		if err != nil {
			return utils.SendError(err, http.StatusBadRequest)
		}

		kecamatanId = &Id
	}

	if payload.Page <= 0 {
		payload.Page = 1
	}
	if payload.Limit <= 0 {
		payload.Limit = 25
	}

	data, totalData, err := service.manajemenWilayahRepo.GetListKelurahan(payload, kecamatanId)
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

	return utils.SendData(result, "Berhasil mengambil list kelurahan")
}

func (s *manajemenWilayahService) GetKelurahanOptions(param url.Values) (*pb.ProxyResponse, error) {
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

	kecamatanId, err := strconv.Atoi(param.Get("kecamatan_id"))
	if err != nil || kecamatanId <= 0 {
		kecamatanId = 0
	}

	_req := payloads.KelurahanOptionsPayload{
		Q:           param.Get("q"),
		Page:        page,
		Limit:       limit,
		IDs:         parsedIDs,
		KecamatanId: kecamatanId,
	}

	data, totalData, err := s.manajemenWilayahRepo.GetKelurahanOptions(_req)
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

	return utils.SendData(responseData, "Berhasil mengambil opsi kelurahan")
}

func (service *manajemenWilayahService) DeleteKelurahan(slug map[string]interface{}) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()
	StrId := slug["kelurahan_id"]
	Id, err := utils.ToInt64(StrId)
	if err != nil {
		return utils.SendError(err, http.StatusBadRequest)
	}

	isUsed, err := service.manajemenWilayahRepo.IsKelurahanUsed(Id)
	if err != nil {
		return utils.SendError(err, http.StatusInternalServerError)
	}

	if isUsed {
		err := errors.New("Data tidak dapat dihapus karena masih digunakan oleh data lain.")
		return utils.SendError(err, http.StatusBadRequest)
	}

	err = service.manajemenWilayahRepo.DeleteKelurahanById(Id)
	if err != nil {
		return utils.SendError(err, http.StatusInternalServerError)
	}

	return utils.SendData(nil, "Berhasil menghapus kelurahan")
}

func (service *manajemenWilayahService) GetRwDetail(slug map[string]interface{}) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()
	StrId := slug["rw_id"]
	Id, err := utils.ToInt64(StrId)
	if err != nil {
		return utils.SendError(err, http.StatusBadRequest)
	}

	data, err := service.manajemenWilayahRepo.GetRwByID(int(Id))
	if err != nil {
		return utils.SendError(err, http.StatusInternalServerError)
	}
	return utils.SendData(data, "Berhasil mengambil data")
}

func (service *manajemenWilayahService) UpdateRw(usr models.JwtCustomClaims, req map[string]interface{}, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()

	var payload payloads.UpdateRwPayload

	err := utils.DynamicBind(req, &payload)
	if err != nil {
		return utils.SendError(err, http.StatusBadRequest)
	}

	StrId := slug["rw_id"]
	Id, err := utils.ToInt64(StrId)
	if err != nil {
		return utils.SendError(err, http.StatusBadRequest)
	}

	// if payload.NamaRw == "" {
	// 	err := errors.New("Nama rw tidak boleh kosong")
	// 	return utils.SendError(err, http.StatusBadRequest)
	// }

	getRwById, err := service.manajemenWilayahRepo.GetRwByID(int(Id))
	if err != nil {
		if err.Error() == gorm.ErrRecordNotFound.Error() {
			err := errors.New("Data rw tidak ditemukan")
			return utils.SendError(err, http.StatusNotFound)
		}
		return utils.SendError(err, http.StatusInternalServerError)
	}

	if getRwById.NamaRw != payload.NamaRw {
		getRwByName, err := service.manajemenWilayahRepo.GetRwByName(payload.NamaRw, &getRwById.VillageId)
		if err != nil && err.Error() != gorm.ErrRecordNotFound.Error() {
			return utils.SendError(err, http.StatusInternalServerError)
		}

		if getRwByName != nil {
			err := errors.New("Nama rw sudah digunakan")
			return utils.SendError(err, http.StatusBadRequest)
		}
	}

	var data = models.DataRw{
		ID:          getRwById.ID,
		KelurahanId: getRwById.VillageId,
		NamaRw:      getRwById.NamaRw,
		KodeWilayah: payload.KodeWilayah,
		Lat:         payload.Lat,
		Long:        payload.Long,
		CreatedAt:   getRwById.CreatedAt,
	}

	_, err = service.manajemenWilayahRepo.UpdateRw(data)
	if err != nil {
		return utils.SendError(err, http.StatusInternalServerError)
	}

	return utils.SendData(nil, "Berhasil update data")
}

func (service *manajemenWilayahService) GetListRw(req map[string]interface{}, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()

	var payload payloads.DatatablePayload
	err := utils.DynamicBind(req, &payload)
	if err != nil {
		return utils.SendError(err, http.StatusBadRequest)
	}

	StrId := slug["kelurahan_id"]
	var kelurahanId *int64
	if StrId != nil {
		Id, err := utils.ToInt64(StrId)
		if err != nil {
			return utils.SendError(err, http.StatusBadRequest)
		}

		kelurahanId = &Id
	}

	if payload.Page <= 0 {
		payload.Page = 1
	}
	if payload.Limit <= 0 {
		payload.Limit = 25
	}

	data, totalData, err := service.manajemenWilayahRepo.GetListRw(payload, kelurahanId)
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

	return utils.SendData(result, "Berhasil mengambil list rw")
}

func (service *manajemenWilayahService) CreateRw(usr models.JwtCustomClaims, req map[string]interface{}) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()

	var payload payloads.CreateRwPayload

	err := utils.DynamicBind(req, &payload)
	if err != nil {
		return utils.SendError(err, http.StatusBadRequest)
	}

	if payload.KelurahanId == 0 {
		err := errors.New("KelurahanId tidak boleh kosong")
		return utils.SendError(err, http.StatusBadRequest)
	}

	if payload.NamaRw == "" {
		err := errors.New("Nama rw tidak boleh kosong")
		return utils.SendError(err, http.StatusBadRequest)
	}

	payload.NamaRw = strings.TrimLeft(payload.NamaRw, "0")
	if payload.NamaRw == "" {
		payload.NamaRw = "0"
	}

	for _, char := range payload.NamaRw {
		if char < '0' || char > '9' {
			err := errors.New("Nama rw harus berupa angka")
			return utils.SendError(err, http.StatusBadRequest)
		}
	}

	getRwByName, err := service.manajemenWilayahRepo.GetRwByName(payload.NamaRw, &payload.KelurahanId)
	if err != nil && err.Error() != gorm.ErrRecordNotFound.Error() {
		return utils.SendError(err, http.StatusInternalServerError)
	}

	if getRwByName != nil {
		err := errors.New("Nama rw sudah digunakan")
		return utils.SendError(err, http.StatusBadRequest)
	}

	var data = models.DataRw{
		KelurahanId: payload.KelurahanId,
		NamaRw:      payload.NamaRw,
		KodeWilayah: payload.KodeWilayah,
		Lat:         payload.Lat,
		Long:        payload.Long,
	}

	dataCreate, err := service.manajemenWilayahRepo.CreateRw(data)
	if err != nil {
		return utils.SendError(err, http.StatusInternalServerError)
	}

	return utils.SendData(dataCreate, "Berhasil menambahkan rw")
}

func (s *manajemenWilayahService) GetRwOptions(param url.Values) (*pb.ProxyResponse, error) {
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

	_req := payloads.RwOptionsPayload{
		Q:           param.Get("q"),
		Page:        page,
		Limit:       limit,
		IDs:         parsedIDs,
		KelurahanId: kelurahanId,
	}

	data, totalData, err := s.manajemenWilayahRepo.GetRwOptions(_req)
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

	return utils.SendData(responseData, "Berhasil mengambil opsi rw")
}

func (service *manajemenWilayahService) DeleteRw(slug map[string]interface{}) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()
	StrId := slug["rw_id"]
	Id, err := utils.ToInt64(StrId)
	if err != nil {
		return utils.SendError(err, http.StatusBadRequest)
	}

	isUsed, err := service.manajemenWilayahRepo.IsRwUsed(Id)
	if err != nil {
		return utils.SendError(err, http.StatusInternalServerError)
	}

	if isUsed {
		err := errors.New("Data tidak dapat dihapus karena masih digunakan oleh data lain.")
		return utils.SendError(err, http.StatusBadRequest)
	}

	err = service.manajemenWilayahRepo.DeleteRwById(Id)
	if err != nil {
		return utils.SendError(err, http.StatusInternalServerError)
	}

	return utils.SendData(nil, "Berhasil menghapus rw")
}

func (service *manajemenWilayahService) GetRtDetail(slug map[string]interface{}) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()
	StrId := slug["rt_id"]
	Id, err := utils.ToInt64(StrId)
	if err != nil {
		return utils.SendError(err, http.StatusBadRequest)
	}

	data, err := service.manajemenWilayahRepo.GetRtByID(Id)
	if err != nil {
		return utils.SendError(err, http.StatusInternalServerError)
	}
	return utils.SendData(data, "Berhasil mengambil data")
}

func (service *manajemenWilayahService) UpdateRt(usr models.JwtCustomClaims, req map[string]interface{}, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()

	var payload payloads.UpdateRtPayload

	err := utils.DynamicBind(req, &payload)
	if err != nil {
		return utils.SendError(err, http.StatusBadRequest)
	}

	StrId := slug["rt_id"]
	Id, err := utils.ToInt64(StrId)
	if err != nil {
		return utils.SendError(err, http.StatusBadRequest)
	}

	// if payload.NamaRt == "" {
	// 	err := errors.New("Nama rt tidak boleh kosong")
	// 	return utils.SendError(err, http.StatusBadRequest)
	// }

	getRtById, err := service.manajemenWilayahRepo.GetRtByID(Id)
	if err != nil {
		if err.Error() == gorm.ErrRecordNotFound.Error() {
			err := errors.New("Data rt tidak ditemukan")
			return utils.SendError(err, http.StatusNotFound)
		}
		return utils.SendError(err, http.StatusInternalServerError)
	}

	if getRtById.NamaRt != payload.NamaRt {
		getRtByName, err := service.manajemenWilayahRepo.GetRtByName(payload.NamaRt, &getRtById.RwId)
		if err != nil && err.Error() != gorm.ErrRecordNotFound.Error() {
			return utils.SendError(err, http.StatusInternalServerError)
		}

		if getRtByName != nil {
			err := errors.New("Nama rt sudah digunakan")
			return utils.SendError(err, http.StatusBadRequest)
		}
	}

	var data = models.DataRt{
		ID:          getRtById.ID,
		RwId:        getRtById.RwId,
		NamaRt:      getRtById.NamaRt,
		KodeWilayah: payload.KodeWilayah,
		Lat:         payload.Lat,
		Long:        payload.Long,
		CreatedAt:   getRtById.CreatedAt,
	}

	_, err = service.manajemenWilayahRepo.UpdateRt(data)
	if err != nil {
		return utils.SendError(err, http.StatusInternalServerError)
	}

	return utils.SendData(nil, "Berhasil update data")
}

func (service *manajemenWilayahService) GetListRt(req map[string]interface{}, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()

	var payload payloads.DatatablePayload
	err := utils.DynamicBind(req, &payload)
	if err != nil {
		return utils.SendError(err, http.StatusBadRequest)
	}

	StrId := slug["rw_id"]
	var rwId *int64
	if StrId != nil {
		Id, err := utils.ToInt64(StrId)
		if err != nil {
			return utils.SendError(err, http.StatusBadRequest)
		}

		rwId = &Id
	}

	if payload.Page <= 0 {
		payload.Page = 1
	}
	if payload.Limit <= 0 {
		payload.Limit = 25
	}

	data, totalData, err := service.manajemenWilayahRepo.GetListRt(payload, rwId)
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

	return utils.SendData(result, "Berhasil mengambil list rt")
}

func (service *manajemenWilayahService) CreateRt(usr models.JwtCustomClaims, req map[string]interface{}) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()

	var payload payloads.CreateRtPayload

	err := utils.DynamicBind(req, &payload)
	if err != nil {
		return utils.SendError(err, http.StatusBadRequest)
	}

	if payload.RwId == 0 {
		err := errors.New("RwId tidak boleh kosong")
		return utils.SendError(err, http.StatusBadRequest)
	}

	if payload.NamaRt == "" {
		err := errors.New("Nama rt tidak boleh kosong")
		return utils.SendError(err, http.StatusBadRequest)
	}

	payload.NamaRt = strings.TrimLeft(payload.NamaRt, "0")
	if payload.NamaRt == "" {
		payload.NamaRt = "0"
	}

	for _, char := range payload.NamaRt {
		if char < '0' || char > '9' {
			err := errors.New("Nama rt harus berupa angka")
			return utils.SendError(err, http.StatusBadRequest)
		}
	}

	getRtByName, err := service.manajemenWilayahRepo.GetRtByName(payload.NamaRt, &payload.RwId)
	if err != nil && err.Error() != gorm.ErrRecordNotFound.Error() {
		return utils.SendError(err, http.StatusInternalServerError)
	}

	if getRtByName != nil {
		err := errors.New("Nama rt sudah digunakan")
		return utils.SendError(err, http.StatusBadRequest)
	}

	var data = models.DataRt{
		RwId:        payload.RwId,
		NamaRt:      payload.NamaRt,
		KodeWilayah: payload.KodeWilayah,
		Lat:         payload.Lat,
		Long:        payload.Long,
	}

	dataCreate, err := service.manajemenWilayahRepo.CreateRt(data)
	if err != nil {
		return utils.SendError(err, http.StatusInternalServerError)
	}

	return utils.SendData(dataCreate, "Berhasil menambahkan rt")
}

func (s *manajemenWilayahService) GetRtOptions(param url.Values) (*pb.ProxyResponse, error) {
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

	rwId, err := strconv.Atoi(param.Get("rw_id"))
	if err != nil || rwId <= 0 {
		rwId = 0
	}

	_req := payloads.RtOptionsPayload{
		Q:     param.Get("q"),
		Page:  page,
		Limit: limit,
		IDs:   parsedIDs,
		RwId:  rwId,
	}

	data, totalData, err := s.manajemenWilayahRepo.GetRtOptions(_req)
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

	return utils.SendData(responseData, "Berhasil mengambil opsi rt")
}

func (service *manajemenWilayahService) DeleteRT(slug map[string]interface{}) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()
	StrId := slug["rt_id"]
	Id, err := utils.ToInt64(StrId)
	if err != nil {
		return utils.SendError(err, http.StatusBadRequest)
	}

	getRtById, err := service.manajemenWilayahRepo.GetRtByID(Id)
	if err != nil {
		return utils.SendError(err, http.StatusInternalServerError)
	}

	isRTUsed, err := service.manajemenWilayahRepo.IsRtUsed(getRtById.ID)
	if err != nil {
		return utils.SendError(err, http.StatusInternalServerError)
	}

	if isRTUsed {
		err := errors.New("Data tidak dapat dihapus karena masih digunakan oleh data lain.")
		return utils.SendError(err, http.StatusBadRequest)
	}

	err = service.manajemenWilayahRepo.DeleteRtById(Id)
	if err != nil {
		return utils.SendError(err, http.StatusInternalServerError)
	}

	return utils.SendData(nil, "Berhasil menghapus rt")
}
