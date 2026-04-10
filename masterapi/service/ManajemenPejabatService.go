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
	"net/http"
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
	GetListPejabat(ctx context.Context, usr models.JwtCustomClaims, req map[string]interface{}) (*pb.ProxyResponse, error)
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

	isPejabatExist, err := service.manajemenPejabatRepo.IsPejabatExistByRespondenId(respondentData.ID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return utils.SendError(err, http.StatusInternalServerError)
	}
	if isPejabatExist {
		return utils.SendError(errors.New("Responden sudah terdaftar sebagai pejabat wilayah"), http.StatusBadRequest)
	}

	tipeWilayah, isValid := enums.GetWilayahFromRole(respondentData.RoleId)
	if !isValid {
		return utils.SendError(errors.New("Role respondent tidak valid untuk dijadikan pejabat wilayah"), http.StatusBadRequest)
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
		return utils.SendError(errors.New("gagal parse string ke int64"), http.StatusBadRequest)
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

	if isRespondenBerubah {

		isPejabatExist, err := service.manajemenPejabatRepo.IsPejabatExistByRespondenId(finalIdResponden)
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return utils.SendError(err, http.StatusInternalServerError)
		}
		if isPejabatExist {
			return utils.SendError(errors.New("Responden sudah terdaftar sebagai pejabat wilayah lain"), http.StatusBadRequest)
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
	}

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

	return utils.SendData(nil, "Berhasil menghapus pejabat")
}

func (service *manajemenPejabatService) GetListPejabat(ctx context.Context, usr models.JwtCustomClaims, req map[string]interface{}) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()

	var payload payloads.DatatablePejabatPayload
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

	data, totalData, err := service.manajemenPejabatRepo.GetListPejabat(payload)
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
