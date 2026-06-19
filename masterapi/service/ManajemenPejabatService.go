package service

import (
	"backend/masterapi/enums"
	"backend/masterapi/models"
	"backend/masterapi/payloads"
	"backend/masterapi/repository"
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
	"time"

	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"
)

type ManajemenPejabatService interface {
	CreatePejabat(ctx context.Context, usr models.JwtCustomClaims, req map[string]interface{}) (*pb.ProxyResponse, error)
	DetailPejabat(ctx context.Context, usr models.JwtCustomClaims, req map[string]interface{}, slug map[string]interface{}) (*pb.ProxyResponse, error)
	UpdatePejabat(ctx context.Context, usr models.JwtCustomClaims, req map[string]interface{}, reqSlug map[string]interface{}) (*pb.ProxyResponse, error)
	DeletePejabat(ctx context.Context, usr models.JwtCustomClaims, req map[string]interface{}, slug map[string]interface{}) (*pb.ProxyResponse, error)
	GetListPejabat(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
}

type manajemenPejabatService struct {
	manajemenPejabatRepo repository.ManajemenPejabatRepo
}

func NewManajemenPejabatService(
	manajemenPejabatRepo repository.ManajemenPejabatRepo,
) ManajemenPejabatService {
	return &manajemenPejabatService{
		manajemenPejabatRepo,
	}
}

func (service *manajemenPejabatService) CreatePejabat(ctx context.Context, usr models.JwtCustomClaims, req map[string]interface{}) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()

	var payload payloads.PejabatPayload

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

	hostUserAPI := os.Getenv("USERAPI_HOST") + ":" + os.Getenv("USERAPI_PORT")
	slug := map[string]interface{}{"id": strconv.FormatInt(payload.NamaPejabat, 10)}

	dataBytes, err := utils.HitBackend(ctx, hostUserAPI, "GET", "/respondent/raw/:id", slug, nil)
	if err != nil {
		return utils.SendError(err, http.StatusInternalServerError)
	}

	var respondentData models.Respondent
	if err := json.Unmarshal(dataBytes, &respondentData); err != nil {
		return utils.SendError(errors.New("Gagal memparsing data respondent dari UserAPI"), http.StatusInternalServerError)
	}

	tipeWilayah, isValid := enums.GetWilayahFromRole(respondentData.RoleId)
	if !isValid {
		return utils.SendError(errors.New("Role respondent tidak valid untuk dijadikan pejabat wilayah"), http.StatusBadRequest)
	}

	isPejabatExist, err := service.manajemenPejabatRepo.IsPejabatExistByRespondenId(respondentData.ID, tipeWilayah)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return utils.SendError(err, http.StatusInternalServerError)
	}

	if isPejabatExist {
		return utils.SendError(errors.New("Responden sudah terdaftar sebagai pejabat wilayah"), http.StatusBadRequest)
	}

	var IdWilayah int64

	switch tipeWilayah {
	case int64(enums.KECAMATAN):
		if respondentData.KecamatanId == nil {
			return utils.SendError(errors.New("Data respondent tidak memiliki ID Kecamatan"), http.StatusBadRequest)
		}
		IdWilayah = int64(*respondentData.KecamatanId)
	case int64(enums.KELURAHAN):
		if respondentData.KelurahanId == nil {
			return utils.SendError(errors.New("Data respondent tidak memiliki ID Kelurahan"), http.StatusBadRequest)
		}
		IdWilayah = int64(*respondentData.KelurahanId)
	case int64(enums.RW):
		if respondentData.RWId == nil {
			return utils.SendError(errors.New("Data respondent tidak memiliki ID RW"), http.StatusBadRequest)
		}
		IdWilayah = int64(*respondentData.RWId)
	case int64(enums.RT):
		if respondentData.RTId == nil {
			return utils.SendError(errors.New("Data respondent tidak memiliki ID RT"), http.StatusBadRequest)
		}
		IdWilayah = int64(*respondentData.RTId)
	default:
		return utils.SendError(errors.New("Tipe wilayah tidak dikenali"), http.StatusBadRequest)
	}

	if payload.PeriodeAwal == nil || payload.PeriodeAkhir == nil {
		return utils.SendError(errors.New("Periode awal dan periode akhir wajib diisi"), http.StatusBadRequest)
	}

	layout := "2006-01-02"

	periodeAwal, err := time.Parse(layout, *payload.PeriodeAwal)
	if err != nil {
		return utils.SendError(errors.New("Format periode awal tidak valid, gunakan format YYYY-MM-DD"), http.StatusBadRequest)
	}

	periodeAkhir, err := time.Parse(layout, *payload.PeriodeAkhir)
	if err != nil {
		return utils.SendError(errors.New("Format periode akhir tidak valid, gunakan format YYYY-MM-DD"), http.StatusBadRequest)
	}

	if payload.StatusJabat == nil {
		return utils.SendError(errors.New("Status jabatan wajib diisi"), http.StatusBadRequest)
	}

	var statusJabatan int
	if *payload.StatusJabat {
		statusJabatan = 1
	} else {
		statusJabatan = 0
	}

	pejabat := &models.PejabatWilayah{
		TipeWilayah:  &tipeWilayah,
		IdWilayah:    &IdWilayah,
		PeriodeAwal:  &periodeAwal,
		PeriodeAkhir: &periodeAkhir,
		NoSK:         payload.NoSK,
		StatusJabat:  &statusJabatan,
		IdResponden:  &payload.NamaPejabat,
	}

	createData, err := service.manajemenPejabatRepo.CreatePejabat(pejabat)
	if err != nil {
		return utils.SendError(err, http.StatusInternalServerError)
	}

	go utils.SaveLogActivities("Master", "Manajemen Pejabat", "POST", int(usr.ID), string(usr.Name), string(0), "Membuat Data Pejabat")

	return utils.SendData(createData, "Berhasil menambahkan pejabat")
}

func (service *manajemenPejabatService) DetailPejabat(ctx context.Context, usr models.JwtCustomClaims, req map[string]interface{}, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()

	idStr, ok := slug["id"].(string)
	if !ok {
		return utils.SendError(errors.New("Invalid ID in slug"), http.StatusBadRequest)
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return utils.SendError(errors.New("gagal parse string ke int64"), http.StatusBadRequest)
	}

	pejabat, err := service.manajemenPejabatRepo.GetPejabatById(id)
	if err != nil {
		if err.Error() == gorm.ErrRecordNotFound.Error() {
			return utils.SendError(errors.New("Pejabat wilayah tidak ditemukan"), http.StatusNotFound)
		}
		return utils.SendError(err, http.StatusInternalServerError)
	}

	return utils.SendData(pejabat, "Detail Pejabat")
}

func (service *manajemenPejabatService) UpdatePejabat(ctx context.Context, usr models.JwtCustomClaims, req map[string]interface{}, reqSlug map[string]interface{}) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()

	idStr, ok := reqSlug["id"].(string)
	if !ok {
		return utils.SendError(errors.New("Invalid ID in slug"), http.StatusBadRequest)
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return utils.SendError(errors.New("Pejabat tidak ditemukan"), http.StatusBadRequest)
	}

	existingPejabat, err := service.manajemenPejabatRepo.GetPejabatById(id)
	if err != nil {
		if err.Error() == gorm.ErrRecordNotFound.Error() {
			return utils.SendError(errors.New("Pejabat wilayah tidak ditemukan"), http.StatusNotFound)
		}
		return utils.SendError(err, http.StatusInternalServerError)
	}

	var payload payloads.PejabatPayload
	err = utils.DynamicBind(req, &payload)
	if err != nil {
		return utils.SendError(err, http.StatusBadRequest)
	}

	var validate = validator.New()
	err = validate.Struct(payload)
	if err != nil {
		customErrorMsg := utils.FormatValidationError(err)
		return utils.SendError(errors.New(customErrorMsg), http.StatusBadRequest)
	}

	var finalTipeWilayah int64 = *existingPejabat.TipeWilayah
	var finalIdWilayah int64 = *existingPejabat.IdWilayah
	var finalIdResponden int64 = payload.NamaPejabat

	if finalIdResponden == 0 && existingPejabat.IdResponden != nil {
		finalIdResponden = *existingPejabat.IdResponden
	}

	isRespondenBerubah := true
	if existingPejabat.IdResponden != nil && *existingPejabat.IdResponden == finalIdResponden {
		isRespondenBerubah = false
	}

	hostUserAPI := os.Getenv("USERAPI_HOST") + ":" + os.Getenv("USERAPI_PORT")
	userApiSlug := map[string]interface{}{"id": strconv.FormatInt(finalIdResponden, 10)}

	dataBytes, err := utils.HitBackend(ctx, hostUserAPI, "GET", "/respondent/raw/:id", userApiSlug, nil)
	if err != nil {
		return utils.SendError(err, http.StatusInternalServerError)
	}

	var respondentData models.Respondent
	if err := json.Unmarshal(dataBytes, &respondentData); err != nil {
		return utils.SendError(errors.New("Gagal memparsing data respondent dari UserAPI"), http.StatusInternalServerError)
	}

	if isRespondenBerubah {
		isPejabatExist, err := service.manajemenPejabatRepo.IsRespondentHaveActivePejabat(finalIdResponden)
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return utils.SendError(err, http.StatusInternalServerError)
		}
		if isPejabatExist {
			return utils.SendError(errors.New("Responden sudah terdaftar sebagai pejabat wilayah lain"), http.StatusBadRequest)
		}
	}

	tipeWilayah, isValid := enums.GetWilayahFromRole(respondentData.RoleId)
	if !isValid {
		return utils.SendError(errors.New("Role respondent tidak valid untuk dijadikan pejabat wilayah"), http.StatusBadRequest)
	}

	switch tipeWilayah {
	case int64(enums.KECAMATAN):
		if respondentData.KecamatanId == nil {
			return utils.SendError(errors.New("Data respondent tidak memiliki ID Kecamatan"), http.StatusBadRequest)
		}
		finalIdWilayah = int64(*respondentData.KecamatanId)
	case int64(enums.KELURAHAN):
		if respondentData.KelurahanId == nil {
			return utils.SendError(errors.New("Data respondent tidak memiliki ID Kelurahan"), http.StatusBadRequest)
		}
		finalIdWilayah = int64(*respondentData.KelurahanId)
	case int64(enums.RW):
		if respondentData.RWId == nil {
			return utils.SendError(errors.New("Data respondent tidak memiliki ID RW"), http.StatusBadRequest)
		}
		finalIdWilayah = int64(*respondentData.RWId)
	case int64(enums.RT):
		if respondentData.RTId == nil {
			return utils.SendError(errors.New("Data respondent tidak memiliki ID RT"), http.StatusBadRequest)
		}
		finalIdWilayah = int64(*respondentData.RTId)
	default:
		return utils.SendError(errors.New("Tipe wilayah tidak dikenali"), http.StatusBadRequest)
	}

	finalTipeWilayah = tipeWilayah

	// 6. Validasi dan Parsing Tanggal
	if payload.PeriodeAwal == nil || payload.PeriodeAkhir == nil {
		return utils.SendError(errors.New("Periode awal dan periode akhir wajib diisi"), http.StatusBadRequest)
	}

	layout := "2006-01-02"

	periodeAwal, err := time.Parse(layout, *payload.PeriodeAwal)
	if err != nil {
		return utils.SendError(errors.New("Format periode awal tidak valid, gunakan format YYYY-MM-DD"), http.StatusBadRequest)
	}

	periodeAkhir, err := time.Parse(layout, *payload.PeriodeAkhir)
	if err != nil {
		return utils.SendError(errors.New("Format periode akhir tidak valid, gunakan format YYYY-MM-DD"), http.StatusBadRequest)
	}

	if payload.StatusJabat == nil {
		return utils.SendError(errors.New("Status jabatan wajib diisi"), http.StatusBadRequest)
	}

	var statusJabatan int
	if *payload.StatusJabat {
		statusJabatan = 1
	} else {
		statusJabatan = 0
	}

	updatePejabat := &models.PejabatWilayah{
		ID:           existingPejabat.ID,
		TipeWilayah:  &finalTipeWilayah,
		IdWilayah:    &finalIdWilayah,
		PeriodeAwal:  &periodeAwal,
		PeriodeAkhir: &periodeAkhir,
		NoSK:         payload.NoSK,
		StatusJabat:  &statusJabatan,
		IdResponden:  &finalIdResponden,
		CreatedAt:    existingPejabat.CreatedAt,
	}

	updatedData, err := service.manajemenPejabatRepo.UpdatePejabat(updatePejabat)
	if err != nil {
		return utils.SendError(err, http.StatusInternalServerError)
	}

	go utils.SaveLogActivities("Master", "Manajemen Pejabat", "PUT", int(usr.ID), string(usr.Name), string(id), "Memperbarui Data Pejabat")

	return utils.SendData(updatedData, "Berhasil mengupdate pejabat")
}

func (service *manajemenPejabatService) DeletePejabat(ctx context.Context, usr models.JwtCustomClaims, req map[string]interface{}, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()

	idStr, ok := slug["id"].(string)
	if !ok {
		return utils.SendError(errors.New("Invalid ID in slug"), http.StatusBadRequest)
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return utils.SendError(errors.New("gagal parse string ke int64"), http.StatusBadRequest)
	}

	existingPejabat, err := service.manajemenPejabatRepo.GetPejabatById(id)
	if err != nil {
		if err.Error() == gorm.ErrRecordNotFound.Error() {
			return utils.SendError(errors.New("Pejabat wilayah tidak ditemukan"), http.StatusNotFound)
		}
		return utils.SendError(err, http.StatusInternalServerError)
	}

	err = service.manajemenPejabatRepo.DeletePejabatById(existingPejabat.ID)
	if err != nil {
		return utils.SendError(err, http.StatusInternalServerError)
	}

	go utils.SaveLogActivities("Master", "Manajemen Pejabat", "DELETE", int(usr.ID), string(usr.Name), string(id), "Menghapus Data Pejabat")

	return utils.SendData(nil, "Berhasil menghapus pejabat")
}

func (service *manajemenPejabatService) GetListPejabat(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()

	search := param.Get("search")
	page, _ := strconv.Atoi(param.Get("page"))
	limit, _ := strconv.Atoi(param.Get("limit"))
	orderBy := param.Get("order_by")
	orderDir := param.Get("order_dir")

	fNama := param.Get("f_nama")
	fPeriodeAwal := param.Get("f_periode_awal")
	fPeriodeAkhir := param.Get("f_periode_akhir")
	FStatusStr := param.Get("f_status")
	var fStatus *int
	if FStatusStr != "" {
		statusInt, err := strconv.Atoi(FStatusStr)
		if err != nil {
			return utils.SendError(errors.New("Invalid value for f_status, must be an integer"), http.StatusBadRequest)
		}
		fStatus = &statusInt
	}

	fKecamatan := param.Get("f_kecamatan")
	var fKecamatanInt *int64
	if fKecamatan != "" {
		kecamatanId, err := strconv.ParseInt(fKecamatan, 10, 64)
		if err != nil {
			return utils.SendError(errors.New("Invalid value for f_kecamatan, must be an integer"), http.StatusBadRequest)
		}
		fKecamatanInt = &kecamatanId
	}

	fKelurahan := param.Get("f_kelurahan")
	var fKelurahanInt *int64
	if fKelurahan != "" {
		kelurahanId, err := strconv.ParseInt(fKelurahan, 10, 64)
		if err != nil {
			return utils.SendError(errors.New("Invalid value for f_kelurahan, must be an integer"), http.StatusBadRequest)
		}
		fKelurahanInt = &kelurahanId
	}

	fRw := param.Get("f_rw")
	var fRwInt *int64
	if fRw != "" {
		rwId, err := strconv.ParseInt(fRw, 10, 64)
		if err != nil {
			return utils.SendError(errors.New("Invalid value for f_rw, must be an integer"), http.StatusBadRequest)
		}
		fRwInt = &rwId
	}

	fRt := param.Get("f_rt")
	var fRtInt *int64
	if fRt != "" {
		rtId, err := strconv.ParseInt(fRt, 10, 64)
		if err != nil {
			return utils.SendError(errors.New("Invalid value for f_rt, must be an integer"), http.StatusBadRequest)
		}
		fRtInt = &rtId
	}

	payload := payloads.DatatablePejabatPayload{
		Search:   search,
		Page:     page,
		Limit:    limit,
		OrderBy:  orderBy,
		OrderDir: orderDir,

		FNama:         &fNama,
		FPeriodeAwal:  &fPeriodeAwal,
		FPeriodeAkhir: &fPeriodeAkhir,
		FStatus:       fStatus,
		FKecamatan:    fKecamatanInt,
		FKelurahan:    fKelurahanInt,
		FRw:           fRwInt,
		FRt:           fRtInt,
	}

	if payload.Page <= 0 {
		payload.Page = 1
	}
	if payload.Limit <= 0 {
		payload.Limit = 5
	}

	data, totalData, err := service.manajemenPejabatRepo.GetListPejabat(payload)
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
