package service

import (
	"backend/masterapi/models"
	"backend/masterapi/payloads"
	"backend/masterapi/repository"
	"backend/masterapi/response"
	"backend/masterapi/utils"
	"backend/siccore/pb"
	"context"
	"encoding/json"
	"errors"
	"math"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/speps/go-hashids/v2"
	"gorm.io/gorm"
)

type ManajemenWilayahService interface {
	CreateKecamatan(usr models.JwtCustomClaims, req map[string]interface{}) (*pb.ProxyResponse, error)
	GetKecamatanDetail(slug map[string]interface{}) (*pb.ProxyResponse, error)
	UpdateKecamatan(usr models.JwtCustomClaims, req map[string]interface{}, slug map[string]interface{}) (*pb.ProxyResponse, error)
	GetListKecamatan(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	GetKecamatanOptions(param url.Values) (*pb.ProxyResponse, error)
	DeleteKecamatan(slug map[string]interface{}) (*pb.ProxyResponse, error)

	CreateKelurahan(usr models.JwtCustomClaims, req map[string]interface{}) (*pb.ProxyResponse, error)
	GetKelurahanDetail(slug map[string]interface{}) (*pb.ProxyResponse, error)
	UpdateKelurahan(usr models.JwtCustomClaims, req map[string]interface{}, slug map[string]interface{}) (*pb.ProxyResponse, error)
	GetListKelurahan(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	GetKelurahanOptions(param url.Values) (*pb.ProxyResponse, error)
	DeleteKelurahan(slug map[string]interface{}) (*pb.ProxyResponse, error)

	GetRwDetail(slug map[string]interface{}) (*pb.ProxyResponse, error)
	UpdateRw(usr models.JwtCustomClaims, req map[string]interface{}, slug map[string]interface{}) (*pb.ProxyResponse, error)
	GetListRw(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	CreateRw(usr models.JwtCustomClaims, req map[string]interface{}) (*pb.ProxyResponse, error)
	GetRwOptions(param url.Values) (*pb.ProxyResponse, error)
	DeleteRw(slug map[string]interface{}) (*pb.ProxyResponse, error)

	GetRtDetail(slug map[string]interface{}) (*pb.ProxyResponse, error)
	UpdateRt(usr models.JwtCustomClaims, req map[string]interface{}, slug map[string]interface{}) (*pb.ProxyResponse, error)
	GetListRt(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	CreateRt(usr models.JwtCustomClaims, req map[string]interface{}) (*pb.ProxyResponse, error)
	GetRtOptions(param url.Values) (*pb.ProxyResponse, error)
	DeleteRT(slug map[string]interface{}) (*pb.ProxyResponse, error)

	TabelDataKotaBandung(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
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

	go utils.SaveLogActivities("Master", "Manajemen Wilayah", "POST", int(usr.ID), string(usr.Name), string(0), "Membuat Data Wilayah")

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

	go utils.SaveLogActivities("Master", "Manajemen Wilayah", "PUT", int(usr.ID), string(usr.Name), string(Id), "Memperbarui Data Wilayah")

	return utils.SendData(nil, "Berhasil update data")
}

func (service *manajemenWilayahService) GetListKecamatan(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
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

	hostUserAPI := os.Getenv("USERAPI_HOST") + ":" + os.Getenv("USERAPI_PORT")
	_slug := map[string]interface{}{"id": strconv.FormatInt(usr.RespondentID, 10)}

	dataBytes, err := utils.HitBackend(ctx, hostUserAPI, "GET", "/respondent/raw/:id", _slug, nil)
	if err != nil {
		return utils.SendError(err, http.StatusInternalServerError)
	}

	var respondentData models.Respondent
	if err := json.Unmarshal(dataBytes, &respondentData); err != nil {
		return utils.SendError(errors.New("Gagal memparsing data respondent dari UserAPI"), http.StatusInternalServerError)
	}

	data, totalData, err := service.manajemenWilayahRepo.GetListKecamatan(payload, &respondentData)
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

	_, err = service.manajemenWilayahRepo.GetKecamatanByID(int(Id))
	if err != nil {
		if err.Error() == gorm.ErrRecordNotFound.Error() {
			err := errors.New("Kecamatan tidak ditemukan")
			return utils.SendError(err, http.StatusNotFound)
		}
		return utils.SendError(err, http.StatusInternalServerError)
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

	if payload.NamaKelurahan == "" {
		err := errors.New("Nama kelurahan tidak boleh kosong")
		return utils.SendError(err, http.StatusBadRequest)
	}

	getKelurahanById, err := service.manajemenWilayahRepo.GetKelurahanByID(int(Id))
	if err != nil {
		if err.Error() == gorm.ErrRecordNotFound.Error() {
			err := errors.New("Data kelurahan tidak ditemukan")
			return utils.SendError(err, http.StatusNotFound)
		}
		return utils.SendError(err, http.StatusInternalServerError)
	}

	newSlug := getKelurahanById.VillageNameSlug

	if getKelurahanById.VillageName != payload.NamaKelurahan {
		getKelurahanByName, err := service.manajemenWilayahRepo.GetKelurahanByName(payload.NamaKelurahan)
		if err != nil && err.Error() != gorm.ErrRecordNotFound.Error() {
			return utils.SendError(err, http.StatusInternalServerError)
		}

		if getKelurahanByName != nil {
			err := errors.New("Nama kelurahan sudah digunakan")
			return utils.SendError(err, http.StatusBadRequest)
		}

		newSlug = utils.StringToSlug(payload.NamaKelurahan, "-")
		getKelurahanBySlug, err := service.manajemenWilayahRepo.GetKelurahanBySlug(newSlug)
		if err != nil && err.Error() != gorm.ErrRecordNotFound.Error() {
			return utils.SendError(err, http.StatusInternalServerError)
		}

		if getKelurahanBySlug != nil && getKelurahanBySlug.ID != Id {
			newSlug = utils.StringToSlug(payload.NamaKelurahan+" "+strconv.FormatInt(time.Now().Unix(), 10), "-")
		}
	}

	var data = models.Kelurahan{
		ID:            getKelurahanById.ID,
		SubDistrictId: getKelurahanById.SubDistrictId,
		VillageName:   payload.NamaKelurahan,
		// VillageName:       getKelurahanById.VillageName,
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

func (service *manajemenWilayahService) GetListKelurahan(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
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
		payload.Limit = 5
	}

	hostUserAPI := os.Getenv("USERAPI_HOST") + ":" + os.Getenv("USERAPI_PORT")
	_slug := map[string]interface{}{"id": strconv.FormatInt(usr.RespondentID, 10)}

	dataBytes, err := utils.HitBackend(ctx, hostUserAPI, "GET", "/respondent/raw/:id", _slug, nil)
	if err != nil {
		return utils.SendError(err, http.StatusInternalServerError)
	}

	var respondentData models.Respondent
	if err := json.Unmarshal(dataBytes, &respondentData); err != nil {
		return utils.SendError(errors.New("Gagal memparsing data respondent dari UserAPI"), http.StatusInternalServerError)
	}

	data, totalData, err := service.manajemenWilayahRepo.GetListKelurahan(payload, kecamatanId, &respondentData)
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

	var rawKecamatanIDs []string
	if len(param["kecamatan_id[]"]) > 0 {
		rawKecamatanIDs = param["kecamatan_id[]"]
	} else if len(param["kecamatan_id"]) > 0 {
		rawKecamatanIDs = param["kecamatan_id"]
	}

	var parsedKecamatanIDs []int64
	for _, rawID := range rawKecamatanIDs {
		if id, err := strconv.ParseInt(rawID, 10, 64); err == nil {
			parsedKecamatanIDs = append(parsedKecamatanIDs, id)
		}
	}

	kecamatanId, err := strconv.Atoi(param.Get("kecamatan_id"))
	if err != nil || kecamatanId <= 0 {
		kecamatanId = 0
	}

	_req := payloads.KelurahanOptionsPayload{
		Q:            param.Get("q"),
		Page:         page,
		Limit:        limit,
		IDs:          parsedIDs,
		KecamatanIds: parsedKecamatanIDs,
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

	_, err = service.manajemenWilayahRepo.GetKelurahanByID(int(Id))
	if err != nil {
		if err.Error() == gorm.ErrRecordNotFound.Error() {
			err := errors.New("Kelurahan tidak ditemukan")
			return utils.SendError(err, http.StatusNotFound)
		}
		return utils.SendError(err, http.StatusInternalServerError)
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

func (service *manajemenWilayahService) GetListRw(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
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
		payload.Limit = 5
	}

	hostUserAPI := os.Getenv("USERAPI_HOST") + ":" + os.Getenv("USERAPI_PORT")
	_slug := map[string]interface{}{"id": strconv.FormatInt(usr.RespondentID, 10)}

	dataBytes, err := utils.HitBackend(ctx, hostUserAPI, "GET", "/respondent/raw/:id", _slug, nil)
	if err != nil {
		return utils.SendError(err, http.StatusInternalServerError)
	}

	var respondentData models.Respondent
	if err := json.Unmarshal(dataBytes, &respondentData); err != nil {
		return utils.SendError(errors.New("Gagal memparsing data respondent dari UserAPI"), http.StatusInternalServerError)
	}

	data, totalData, err := service.manajemenWilayahRepo.GetListRw(payload, kelurahanId, &respondentData)
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

	getKelurahanById, err := service.manajemenWilayahRepo.GetKelurahanByID(int(payload.KelurahanId))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			err := errors.New("Data kelurahan tidak ditemukan")
			return utils.SendError(err, http.StatusNotFound)
		}
		return utils.SendError(err, http.StatusInternalServerError)
	}

	if getKelurahanById == nil {
		err := errors.New("Data kelurahan tidak ditemukan")
		return utils.SendError(err, http.StatusNotFound)
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

	var rawKelurahanIDs []string
	if len(param["kelurahan_id[]"]) > 0 {
		rawKelurahanIDs = param["kelurahan_id[]"]
	} else if len(param["kelurahan_id"]) > 0 {
		rawKelurahanIDs = param["kelurahan_id"]
	}

	var parsedKelurahanIDs []int64
	for _, rawID := range rawKelurahanIDs {
		if id, err := strconv.ParseInt(rawID, 10, 64); err == nil {
			parsedKelurahanIDs = append(parsedKelurahanIDs, id)
		}
	}

	_req := payloads.RwOptionsPayload{
		Q:            param.Get("q"),
		Page:         page,
		Limit:        limit,
		IDs:          parsedIDs,
		KelurahanIds: parsedKelurahanIDs,
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

	_, err = service.manajemenWilayahRepo.GetRwByID(int(Id))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			err := errors.New("Data rw tidak ditemukan")
			return utils.SendError(err, http.StatusNotFound)
		}
		return utils.SendError(err, http.StatusInternalServerError)
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

func (service *manajemenWilayahService) GetListRt(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
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
		payload.Limit = 5
	}

	data, totalData, err := service.manajemenWilayahRepo.GetListRt(payload, rwId)
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

	_, err = service.manajemenWilayahRepo.GetRwByID(int(payload.RwId))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			err := errors.New("Data Rw tidak ditemukan")
			return utils.SendError(err, http.StatusNotFound)
		}
		return utils.SendError(err, http.StatusInternalServerError)
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

	var rawRwIDs []string
	if len(param["rw_id[]"]) > 0 {
		rawRwIDs = param["rw_id[]"]
	} else if len(param["rw_id"]) > 0 {
		rawRwIDs = param["rw_id"]
	}

	var parsedRwIDs []int64
	for _, rawID := range rawRwIDs {
		if id, err := strconv.ParseInt(rawID, 10, 64); err == nil {
			parsedRwIDs = append(parsedRwIDs, id)
		}
	}

	_req := payloads.RtOptionsPayload{
		Q:     param.Get("q"),
		Page:  page,
		Limit: limit,
		IDs:   parsedIDs,
		RwIds: parsedRwIDs,
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
		if err.Error() == gorm.ErrRecordNotFound.Error() {
			err := errors.New("Data rt tidak ditemukan")
			return utils.SendError(err, http.StatusNotFound)
		}
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

func (service *manajemenWilayahService) TabelDataKotaBandung(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()

	hd := hashids.NewData()
	hd.Salt = os.Getenv("HASHID_SALT")
	hd.MinLength = 24

	h, err := hashids.NewWithData(hd)
	if err != nil {
		return utils.SendError(err, http.StatusInternalServerError)
	}

	search := param.Get("search")
	page, _ := strconv.Atoi(param.Get("page"))
	limit, _ := strconv.Atoi(param.Get("limit"))
	orderBy := param.Get("order_by")
	orderDir := param.Get("order_dir")
	tipe_wilayah, _ := strconv.Atoi(param.Get("tipe_wilayah"))

	kecamatan_idStr := param.Get("kecamatan_code")
	kelurahan_idStr := param.Get("kelurahan_code")
	rw_idStr := param.Get("rw_code")

	var kecamatan_id, kelurahan_id, rw_id int64

	if kecamatan_idStr != "" {
		decoded, err := h.DecodeWithError(kecamatan_idStr)
		if err != nil {
			return utils.SendError(errors.New("Kecamatan tidak valid"), http.StatusBadRequest)
		}
		kecamatan_id = int64(decoded[0])
	}
	if kelurahan_idStr != "" {
		decoded, err := h.DecodeWithError(kelurahan_idStr)
		if err != nil {
			return utils.SendError(errors.New("Kelurahan tidak valid"), http.StatusBadRequest)
		}
		kelurahan_id = int64(decoded[0])
	}
	if rw_idStr != "" {
		decoded, err := h.DecodeWithError(rw_idStr)
		if err != nil {
			return utils.SendError(errors.New("RW tidak valid"), http.StatusBadRequest)
		}
		rw_id = int64(decoded[0])
	}

	payload := payloads.DatatableDataWilayahKotaBandungPayload{
		Search:      search,
		Page:        page,
		Limit:       limit,
		OrderBy:     orderBy,
		OrderDir:    orderDir,
		TipeWilayah: &tipe_wilayah,
		KecamatanId: &kecamatan_id,
		KelurahanId: &kelurahan_id,
		RWId:        &rw_id,
	}

	if payload.Page <= 0 {
		payload.Page = 1
	}
	if payload.Limit <= 0 {
		payload.Limit = 100
	}
	if payload.OrderBy == "" {
		payload.OrderBy = "id"
	}
	if payload.OrderDir == "" {
		payload.OrderDir = "desc"
	}

	offset := (payload.Page - 1) * payload.Limit
	var metaTotal int64
	var finalData = []response.TabelDataKotaBandungResponse{}

	if payload.TipeWilayah != nil {
		switch *payload.TipeWilayah {
		case 5:
			results, total, err := service.manajemenWilayahRepo.GetListKecamatanV2(ctx, payload, offset)
			if err != nil {
				return utils.SendError(err, http.StatusInternalServerError)
			}
			metaTotal = total
			for i, row := range results {
				code, _ := h.Encode([]int{int(row.ID)})
				finalData = append(finalData, response.TabelDataKotaBandungResponse{
					No:              int64(offset + i + 1),
					ID:              row.ID,
					Code:            code,
					NamaKecamatan:   row.SubDistrictName,
					TotalKelurahan:  row.TotalKelurahan,
					TotalRw:         row.TotalRw,
					TotalRt:         row.TotalRt,
					IsPosibleDetail: true,
				})
			}

		case 4:
			results, total, err := service.manajemenWilayahRepo.GetListKelurahanV2(ctx, payload, offset)
			if err != nil {
				return utils.SendError(err, http.StatusInternalServerError)
			}
			metaTotal = total
			for i, row := range results {
				code, _ := h.Encode([]int{int(row.ID)})
				finalData = append(finalData, response.TabelDataKotaBandungResponse{
					No:              int64(offset + i + 1),
					ID:              row.ID,
					Code:            code,
					NamaKecamatan:   row.SubDistrictName,
					NamaKelurahan:   row.VillageName,
					TotalRw:         row.TotalRw,
					TotalRt:         row.TotalRt,
					IsPosibleDetail: true,
				})
			}

		case 3:
			results, total, err := service.manajemenWilayahRepo.GetListRWV2(ctx, payload, offset)
			if err != nil {
				return utils.SendError(err, http.StatusInternalServerError)
			}
			metaTotal = total
			for i, row := range results {
				code, _ := h.Encode([]int{int(row.ID)})
				finalData = append(finalData, response.TabelDataKotaBandungResponse{
					No:              int64(offset + i + 1),
					ID:              row.ID,
					Code:            code,
					NamaKecamatan:   row.SubDistrictName,
					NamaKelurahan:   row.VillageName,
					NamaRw:          row.NamaRw,
					TotalRt:         row.TotalRt,
					IsPosibleDetail: true,
				})
			}

		case 2:
			results, total, err := service.manajemenWilayahRepo.GetListRTV2(ctx, payload, offset)
			if err != nil {
				return utils.SendError(err, http.StatusInternalServerError)
			}
			metaTotal = total
			for i, row := range results {
				code, _ := h.Encode([]int{int(row.ID)})
				finalData = append(finalData, response.TabelDataKotaBandungResponse{
					No:              int64(offset + i + 1),
					ID:              row.ID,
					Code:            code,
					NamaKecamatan:   row.SubDistrictName,
					NamaKelurahan:   row.VillageName,
					NamaRw:          row.NamaRw,
					NamaRt:          row.NamaRt,
					IsPosibleDetail: false,
				})
			}

		case 0:
			return utils.SendError(errors.New("Tipe wilayah harus diisi"), http.StatusBadRequest)
		default:
			return utils.SendError(errors.New("Tipe wilayah tidak valid"), http.StatusBadRequest)
		}
	} else {
		return utils.SendError(errors.New("Tipe wilayah tidak boleh kosong"), http.StatusBadRequest)
	}

	summary, err := service.manajemenWilayahRepo.GetWilayahSummary(ctx, payload)
	if err != nil {
		return utils.SendError(err, http.StatusInternalServerError)
	}

	metaTotalPages := int(math.Ceil(float64(metaTotal) / float64(payload.Limit)))

	result := map[string]interface{}{
		"data": finalData,
		"meta": map[string]interface{}{
			"total":           int(metaTotal),
			"page":            payload.Page,
			"limit":           payload.Limit,
			"totalPages":      metaTotalPages,
			"total_kecamatan": summary.TotalKecamatan,
			"total_kelurahan": summary.TotalKelurahan,
			"total_rw":        summary.TotalRw,
			"total_rt":        summary.TotalRt,
		},
	}

	return utils.SendData(result, "Berhasil mengambil data wilayah")
}
