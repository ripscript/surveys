package service

import (
	"backend/siccore/pb"
	"backend/userapi/models"
	"backend/userapi/payloads"
	"backend/userapi/repository"
	"backend/userapi/response"
	"backend/userapi/utils"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/go-playground/validator/v10"
	excelize "github.com/xuri/excelize/v2"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type RespondentService interface {
	CreateRespondent(req map[string]interface{}, usr models.JwtCustomClaims) (*pb.ProxyResponse, error)
	GetRespondent(usr models.JwtCustomClaims, param url.Values) (*pb.ProxyResponse, error)
	GetDetailRespondent(slug map[string]interface{}) (*pb.ProxyResponse, error)
	DeleteRespondent(usr models.JwtCustomClaims, slug map[string]interface{}) (*pb.ProxyResponse, error)
	UpdateRespondent(usr models.JwtCustomClaims, slug map[string]interface{}, req map[string]interface{}) (*pb.ProxyResponse, error)
	GetExampleImport() (*pb.ProxyResponse, error)
	ImportRespondent(usr models.JwtCustomClaims, req map[string]interface{}) (*pb.ProxyResponse, error)
	GetRawDetailRespondent(slug map[string]interface{}) (*pb.ProxyResponse, error)
	SurveyorOption(param url.Values) (*pb.ProxyResponse, error)
	BlockRespondent(usr models.JwtCustomClaims, param url.Values) (*pb.ProxyResponse, error)
	GetOptionsRespondent(param url.Values) (*pb.ProxyResponse, error)

	GetRespondentByKecamatan(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	GetRespondentByKelurahan(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	GetRespondentByRW(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	GetRespondentByRT(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)

	UpdatePasswordRespondent(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)

	RespondentOptions(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
}

type respondentService struct {
	respondentRepo   repository.RespondentRepo
	usersRepo        repository.UsersRepo
	loginAttemptRepo repository.LoginAttemptRepository
}

func NewRespondentService(
	respondentRepo repository.RespondentRepo,
	usersRepo repository.UsersRepo,
	loginAttemptRepo repository.LoginAttemptRepository,

) RespondentService {
	return &respondentService{
		respondentRepo,
		usersRepo,
		loginAttemptRepo,
	}
}

func (service *respondentService) GetOptionsRespondent(param url.Values) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()
	data, err := service.respondentRepo.GetOptionsRespondent(param)
	if err != nil {
		return utils.SendError(err, http.StatusInternalServerError)
	}

	return utils.SendData(data)
}

func (service *respondentService) CreateRespondent(req map[string]interface{}, usr models.JwtCustomClaims) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()

	payload := payloads.CreateRespondent{}

	err := utils.DynamicBind(req, &payload)
	if err != nil {
		return utils.SendError(err, http.StatusInternalServerError)
	}

	if len(payload.Respondent) < 1 {
		return utils.SendError(fmt.Errorf("Data Tidak Boleh Kosong"), http.StatusBadRequest)
	}

	tx := service.respondentRepo.BeginTx()

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	for _, dataPengguna := range payload.Respondent {
		dataRespondent := models.CreateRespondents{}

		err := utils.DynamicBind(dataPengguna, &dataRespondent)
		if err != nil {
			tx.Rollback()
			return utils.SendError(err, http.StatusInternalServerError)
		}

		dataRespondent.CreatedAt = utils.TimeNow()
		dataRespondent.UpdatedAt = utils.TimeNow()
		dataRespondent.BlkId = 1
		var role int
		if dataPengguna.Kecamatan != "" {
			int64Kecamatan, err := utils.ToInt64(dataPengguna.Kecamatan)
			if err != nil {
				tx.Rollback()
				return utils.SendError(fmt.Errorf("Gagal Mendapatkan Data Kecamatan"), http.StatusInternalServerError)
			}
			intKecamatan := int(int64Kecamatan)
			dataRespondent.Kecamatan = &intKecamatan
			role = 5
		}
		if dataPengguna.Kelurahan != "" {
			int64Kelurahan, err := utils.ToInt64(dataPengguna.Kelurahan)
			if err != nil {
				tx.Rollback()
				return utils.SendError(fmt.Errorf("Gagal Mendapatkan Data Kelurahan"), http.StatusInternalServerError)
			}
			intKelurahan := int(int64Kelurahan)
			dataRespondent.Kelurahan = &intKelurahan
			role = 4
		}
		if dataPengguna.RW != "" {
			int64RW, err := utils.ToInt64(dataPengguna.RW)
			if err != nil {
				tx.Rollback()
				return utils.SendError(fmt.Errorf("Gagal Mendapatkan Data RW"), http.StatusInternalServerError)
			}
			intRW := int(int64RW)
			dataRespondent.RW = &intRW
			role = 3
		}
		if dataPengguna.RT != "" {
			int64RT, err := utils.ToInt64(dataPengguna.RT)
			if err != nil {
				tx.Rollback()
				return utils.SendError(fmt.Errorf("Gagal Mendapatkan Data RT"), http.StatusInternalServerError)
			}
			intRT := int(int64RT)
			dataRespondent.RT = &intRT
			role = 2
		}

		fmt.Println(role)

		// checkIsWilayahAvailable, err := service.respondentRepo.CheckIsWilayahAvailable(&role, dataRespondent.Kecamatan, dataRespondent.Kelurahan, dataRespondent.RW, dataRespondent.RT)
		// if err != nil {
		// 	tx.Rollback()
		// 	return utils.SendError(err, http.StatusInternalServerError)
		// }

		// if !checkIsWilayahAvailable {
		// 	return utils.SendError(errors.New("Wilayah sudah digunakan oleh responden lain"), http.StatusBadRequest)
		// }

		checkEmailRespondent, checkEmailUsers, err := service.usersRepo.CheckEmail(dataRespondent.Email)
		if err != nil {
			tx.Rollback()
			return utils.SendError(err, http.StatusInternalServerError)
		}

		checkPhoneNumberRespondent, err := service.usersRepo.CheckPhoneNumber(dataRespondent.PhoneNumber)
		if err != nil {
			tx.Rollback()
			return utils.SendError(err, http.StatusInternalServerError)
		}

		checkNikRespondent, checkNikUsers, err := service.usersRepo.CheckNik(dataRespondent.NIK)
		if err != nil {
			tx.Rollback()
			return utils.SendError(err, http.StatusInternalServerError)
		}

		spew.Dump(checkNikRespondent, checkNikUsers)

		if checkEmailRespondent != 0 || checkEmailUsers != 0 {
			tx.Rollback()
			return utils.SendError(fmt.Errorf("email %s sudah digunakan", dataRespondent.Email), http.StatusBadRequest)
		}

		if checkNikRespondent != 0 || checkNikUsers != 0 {
			tx.Rollback()
			return utils.SendError(fmt.Errorf("NIK %s sudah digunakan", dataRespondent.NIK), http.StatusBadRequest)
		}

		if checkPhoneNumberRespondent != 0 {
			tx.Rollback()
			return utils.SendError(fmt.Errorf("nomor telepon %s sudah digunakan", dataRespondent.PhoneNumber), http.StatusBadRequest)
		}
		err = service.respondentRepo.StoreUsers(tx, dataRespondent)
		if err != nil {
			tx.Rollback()
			return utils.SendError(err, http.StatusInternalServerError)
		}
	}

	err = tx.Commit().Error
	if err != nil {
		tx.Rollback()
		return utils.SendError(err, http.StatusInternalServerError)
	}

	err = utils.SaveLogActivities("User", "respondent", "POST", int(usr.ID), string(usr.Name), "-", "Menambahkan Data Responden")
	if err != nil {
		return utils.SendError(err, http.StatusInternalServerError)
	}

	return utils.SendData("Semua data berhasil disimpan")
}

func (service *respondentService) GetRespondent(usr models.JwtCustomClaims, param url.Values) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()
	page, limit, offset, _, err := utils.SetPagination(param)
	if err != nil {
		return utils.SendError(fmt.Errorf("terjadi kesalahan saat memproses pagination: %w", err), http.StatusBadRequest)
	}
	respondent, total, err := service.respondentRepo.GetRespondent(offset, limit, param)
	if err != nil || respondent == nil {
		return utils.SendError(fmt.Errorf("data ditemukan: %w", err), http.StatusNotFound)
	}
	var totalPages int
	if limit == 1 {
		totalPages = 1
	} else {
		totalPages = int(math.Ceil(float64(total) / float64(limit)))
	}
	meta := map[string]interface{}{
		"total":      total,
		"page":       page,
		"limit":      limit,
		"totalPages": totalPages,
	}
	response := map[string]interface{}{
		"data": respondent,
		"meta": meta,
	}

	return utils.SendData(response)
}

func (service *respondentService) GetDetailRespondent(slug map[string]interface{}) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()
	id := slug["id"].(string)
	idInt, err := utils.ToInt64(id)
	if err != nil {
		return utils.SendError(err, http.StatusInternalServerError)
	}
	detailRespondent, err := service.respondentRepo.GetDetailRespondent(int(idInt))
	if err != nil {
		return utils.SendError(fmt.Errorf("data ditemukan: %w", err), http.StatusNotFound)
	}
	return utils.SendData(detailRespondent)
}

func (service *respondentService) DeleteRespondent(usr models.JwtCustomClaims, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()
	id := slug["id"].(string)
	idInt, err := utils.ToInt64(id)
	if err != nil {
		return utils.SendError(err, http.StatusInternalServerError)
	}
	_, err = service.respondentRepo.GetDetailRespondent(int(idInt))
	if err != nil {
		return utils.SendError(fmt.Errorf("data ditemukan: %w", err), http.StatusNotFound)
	}

	updateData := models.DeleteRespondent{}
	updateData.Id = int(idInt)
	updateData.DeletedAt = time.Now()

	err = service.respondentRepo.DeleteUsers(updateData)
	if err != nil {
		return utils.SendError(err, http.StatusInternalServerError)
	}

	err = utils.SaveLogActivities("User", "respondent", "DELETE", int(usr.ID), string(usr.Name), id, "Melakukan Penghapusan Data Respondent")
	if err != nil {
		return utils.SendError(err, http.StatusInternalServerError)
	}
	return utils.SendData("Data Berhasil Dihapus")
}

func (service *respondentService) UpdateRespondent(usr models.JwtCustomClaims, slug map[string]interface{}, req map[string]interface{}) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()
	id := slug["id"].(string)
	idInt, err := utils.ToInt64(id)
	if err != nil {
		return utils.SendError(err, http.StatusInternalServerError)
	}
	_, err = service.respondentRepo.GetDetailRespondent(int(idInt))
	if err != nil {
		return utils.SendError(fmt.Errorf("data ditemukan: %w", err), http.StatusNotFound)
	}

	payload := payloads.UpdateRespondents{}

	err = utils.DynamicBind(req, &payload)
	if err != nil {
		return utils.SendError(err, http.StatusBadRequest)
	}

	updateData := models.UpdateRespondents{}
	err = utils.DynamicBind(payload, &updateData)
	if err != nil {
		return utils.SendError(err, http.StatusBadRequest)
	}
	updateData.Id = int(idInt)
	updateData.UpdatedAt = time.Now()

	err = service.respondentRepo.UpdateUsers(int(idInt), updateData)
	if err != nil {
		return utils.SendError(err, http.StatusInternalServerError)
	}

	err = utils.SaveLogActivities("User", "respondent", "UPDATE", int(usr.ID), string(usr.Name), id, "Memperbaharui Data Responden")
	if err != nil {
		return utils.SendError(err, http.StatusInternalServerError)
	}

	return utils.SendData("Data Berhasil Diperbarui")
}

func (service *respondentService) GetExampleImport() (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()

	filePath := "./storage/template/Template_Import_Responden.xlsx"
	fileBytes, err := ioutil.ReadFile(filePath)
	if err != nil {
		return utils.SendError(err, http.StatusNotFound)
	}

	mimeType := "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	return utils.SetResponseData(fileBytes, true, "Data File,"+mimeType, http.StatusOK, nil, ""), nil
}

func (service *respondentService) ImportRespondent(usr models.JwtCustomClaims, req map[string]interface{}) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()
	tx := service.respondentRepo.BeginTx()
	var file string
	var fileExtension string
	if req["file"] != nil {
		file = req["file"].(string)
	} else {
		return utils.SendError(fmt.Errorf("File Tidak Boleh Kosong"), http.StatusBadRequest)
	}
	if req["file_extension"] != nil {
		fileExtension = req["file_extension"].(string)
	}

	if fileExtension != "xlsx" && fileExtension != ".xlsx" {
		return utils.SendError(fmt.Errorf("Format File Tidak Valid"), http.StatusBadRequest)
	}

	result, err := service.readDataInExcel(file)
	if err != nil {
		tx.Rollback()
		return utils.SendError(err, http.StatusInternalServerError)
	}
	for _, dataResult := range result {
		payload := payloads.ImportRespondents{}
		err = utils.DynamicBind(dataResult, &payload)
		if err != nil {
			tx.Rollback()
			return utils.SendError(err, http.StatusInternalServerError)
		}
		data := models.CreateRespondents{}
		err = utils.DynamicBind(payload, &data)
		if err != nil {
			tx.Rollback()
			return utils.SendError(err, http.StatusInternalServerError)
		}

		checkEmailRespondent, checkEmailUsers, err := service.usersRepo.CheckEmail(data.Email)
		if err != nil {
			tx.Rollback()
			return utils.SendError(err, http.StatusInternalServerError)
		}

		if checkEmailRespondent != 0 || checkEmailUsers != 0 {
			tx.Rollback()
			return utils.SendError(fmt.Errorf("email %s sudah digunakan", data.Email), http.StatusBadRequest)
		}

		checkNikRespondent, checkNikUsers, err := service.usersRepo.CheckNik(data.NIK)
		if err != nil {
			tx.Rollback()
			return utils.SendError(err, http.StatusInternalServerError)
		}

		if checkNikRespondent != 0 || checkNikUsers != 0 {
			tx.Rollback()
			return utils.SendError(fmt.Errorf("NIK %s sudah digunakan", data.NIK), http.StatusBadRequest)
		}

		checkPhoneNumberRespondent, err := service.usersRepo.CheckPhoneNumber(data.PhoneNumber)
		if err != nil {
			tx.Rollback()
			return utils.SendError(err, http.StatusInternalServerError)
		}

		if checkPhoneNumberRespondent != 0 {
			tx.Rollback()
			return utils.SendError(fmt.Errorf("nomor telepon %s sudah digunakan", data.PhoneNumber), http.StatusBadRequest)
		}

		data.CreatedAt = utils.TimeNow()
		data.UpdatedAt = utils.TimeNow()
		var role int

		kecamatan, err := service.respondentRepo.GetKecamatanByName(payload.Kecamatan)
		if err != nil {
			tx.Rollback()
			return utils.SendError(err, http.StatusInternalServerError)
		}
		kelurahan, err := service.respondentRepo.GetKelurahanByName(payload.Kelurahan)
		if err != nil {
			tx.Rollback()
			return utils.SendError(err, http.StatusInternalServerError)
		}
		data.Kelurahan = &kelurahan.ID
		data.Kecamatan = &kecamatan.ID
		data.BlkId = 1

		switch payload.Role {
		case "rt":
			role = 2
		case "rw":
			role = 3
		case "lurah":
			role = 4
		default:
			role = 5
		}

		data.RoleID = role

		err = service.respondentRepo.StoreUsers(tx, data)
		if err != nil {
			tx.Rollback()
			return utils.SendError(err, http.StatusInternalServerError)
		}
	}

	err = tx.Commit().Error
	if err != nil {
		tx.Rollback()
		return utils.SendError(err, http.StatusInternalServerError)
	}

	err = utils.SaveLogActivities("User", "respondent", "POST", int(usr.ID), string(usr.Name), "-", "Melakukan Import Data Responden")
	if err != nil {
		return utils.SendError(err, http.StatusInternalServerError)
	}

	return utils.SendData("Data Berhasil Disimpan")
}

func (service *respondentService) readDataInExcel(file string) ([]map[string]interface{}, error) {
	var result []map[string]interface{}
	fileBytes, err := base64.StdEncoding.DecodeString(file)
	if err != nil {
		return result, err
	}

	tempFile := "temp_import.xlsx"
	err = os.WriteFile(tempFile, fileBytes, 0644)
	if err != nil {
		return result, err
	}
	defer os.Remove(tempFile)

	f, err := excelize.OpenFile(tempFile)
	if err != nil {
		return result, err
	}

	sheetName := f.GetSheetName(0)

	rows, err := f.GetRows(sheetName)
	if err != nil {
		return result, err
	}

	for i, row := range rows {
		if i == 0 {
			continue
		}
		data := map[string]interface{}{
			"nik":             getColumn(row, 0),
			"name":            getColumn(row, 1),
			"place_of_birth":  getColumn(row, 2),
			"date_of_birth":   getColumn(row, 3),
			"address":         getColumn(row, 4),
			"phone_number":    getColumn(row, 5),
			"email":           getColumn(row, 6),
			"roleString":      getColumn(row, 7),
			"kecamatan":       getColumn(row, 8),
			"kelurahan":       getColumn(row, 9),
			"rw":              getColumn(row, 10),
			"rt":              getColumn(row, 11),
			"start_sk_period": getColumn(row, 12),
			"end_sk_period":   getColumn(row, 13),
		}

		result = append(result, data)
	}
	return result, nil
}

func getColumn(row []string, index int) string {
	if len(row) > index {
		return row[index]
	}
	return ""
}

func (service *respondentService) GetRawDetailRespondent(slug map[string]interface{}) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()
	id := slug["id"].(string)
	idInt, err := utils.ToInt64(id)
	if err != nil {
		return utils.SendError(err, http.StatusInternalServerError)
	}
	detailRespondent, err := service.respondentRepo.GetRespondenById(int(idInt))
	if err != nil {
		return utils.SendError(fmt.Errorf("data ditemukan: %w", err), http.StatusNotFound)
	}
	return utils.SendData(detailRespondent)
}

func (service *respondentService) SurveyorOption(param url.Values) (*pb.ProxyResponse, error) {
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

	_req := payloads.SurveyorOptionsPayload{
		Q:     param.Get("q"),
		Page:  page,
		Limit: limit,
		IDs:   parsedIDs,
	}

	data, totalData, err := service.respondentRepo.GetSurveyorOptions(_req)
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

	return utils.SendData(responseData, "Berhasil mengambil opsi surveyor")
}

func (service *respondentService) BlockRespondent(usr models.JwtCustomClaims, param url.Values) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()

	respondentId := param.Get("id")
	blockStatus := param.Get("block")

	if respondentId == "" {
		return utils.SendError(fmt.Errorf("ID Respondent Tidak Ditemukan"), http.StatusBadRequest)
	}

	if blockStatus == "" {
		return utils.SendError(fmt.Errorf("Block Status Tidak Ditemukan"), http.StatusBadRequest)
	}

	intRespondentId, err := utils.ToInt64(respondentId)
	if err != nil {
		return utils.SendError(err, http.StatusInternalServerError)
	}

	checkRespondent, err := service.respondentRepo.CheckRespondent(intRespondentId)
	if err != nil {
		if err.Error() == gorm.ErrRecordNotFound.Error() {
			return utils.SendError(fmt.Errorf("Respondent Tidak Ditemukan"), http.StatusBadRequest)
		}
		return utils.SendError(fmt.Errorf("Terjadi Kesalahan Saat Melakukan Pengecekan Data User"), http.StatusInternalServerError)
	}

	if checkRespondent.IsBlocked == "true" && blockStatus == "true" {
		return utils.SendError(fmt.Errorf("Repondent Terblokir"), http.StatusBadRequest)
	} else if checkRespondent.IsBlocked == "false" && blockStatus == "false" {
		return utils.SendError(fmt.Errorf("Repondent Tidak Terblokir"), http.StatusBadRequest)
	}

	var data models.BlockRespondent

	getUserByRespondentId, err := service.usersRepo.GetUserByRespondentId(checkRespondent.ID)
	if err != nil && err.Error() != gorm.ErrRecordNotFound.Error() {
		return utils.SendError(fmt.Errorf("Terjadi Kesalahan Saat Melakukan Pengecekan Data User"), http.StatusInternalServerError)
	}

	if getUserByRespondentId == nil {
		return utils.SendError(fmt.Errorf("User tidak ditemukan"), http.StatusBadRequest)
	}

	if blockStatus == "false" {
		if err := service.loginAttemptRepo.ResetFailed(getUserByRespondentId.ID); err != nil {
			return utils.SendError(err, http.StatusInternalServerError)
		}
	}

	data.ID = int(intRespondentId)
	data.IsBlocked = blockStatus
	data.UpdatedAt = utils.TimeNow()

	err = service.respondentRepo.RespondentBlock(data)
	if err != nil {
		return utils.SendError(fmt.Errorf("Terjadi Kesalahan Saat Melakukan Block Respondent"), http.StatusInternalServerError)
	}

	err = utils.SaveLogActivities("User", "respondent", "POST", int(usr.ID), string(usr.Name), string(data.ID), "Melakukan Block Data Responden")
	if err != nil {
		return utils.SendError(err, http.StatusInternalServerError)
	}

	return utils.SendData(nil)
}

func (service *respondentService) GetRespondentByKecamatan(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()
	kecamatanId := slug["kecamatan_id"].(string)
	intKecamatanId, err := utils.ToInt64(kecamatanId)
	if err != nil {
		return utils.SendError(err, http.StatusInternalServerError)
	}

	respondent, err := service.respondentRepo.GetRespondentByKecamatanId(ctx, intKecamatanId)
	if err != nil {
		return utils.SendError(fmt.Errorf("data ditemukan: %w", err), http.StatusNotFound)
	}

	return utils.SendData(respondent)
}

func (service *respondentService) GetRespondentByKelurahan(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()
	kelurahanId := slug["kelurahan_id"].(string)
	intKelurahanId, err := utils.ToInt64(kelurahanId)
	if err != nil {
		return utils.SendError(err, http.StatusInternalServerError)
	}

	respondent, err := service.respondentRepo.GetRespondentByKelurahanId(ctx, intKelurahanId)
	if err != nil {
		return utils.SendError(fmt.Errorf("data ditemukan: %w", err), http.StatusNotFound)
	}

	return utils.SendData(respondent)
}

func (service *respondentService) GetRespondentByRW(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()
	rwId := slug["rw_id"].(string)
	intRWId, err := utils.ToInt64(rwId)
	if err != nil {
		return utils.SendError(err, http.StatusInternalServerError)
	}

	respondent, err := service.respondentRepo.GetRespondentByRWId(ctx, intRWId)
	if err != nil {
		return utils.SendError(fmt.Errorf("data ditemukan: %w", err), http.StatusNotFound)
	}

	return utils.SendData(respondent)
}

func (service *respondentService) GetRespondentByRT(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()
	rtId := slug["rt_id"].(string)
	intRTId, err := utils.ToInt64(rtId)
	if err != nil {
		return utils.SendError(err, http.StatusInternalServerError)
	}

	respondent, err := service.respondentRepo.GetRespondentByRTId(ctx, intRTId)
	if err != nil {
		return utils.SendError(fmt.Errorf("data ditemukan: %w", err), http.StatusNotFound)
	}

	return utils.SendData(respondent)
}

func (service *respondentService) UpdatePasswordRespondent(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	tx := service.respondentRepo.BeginTx()
	defer utils.GeneralRecoverWithTrx(tx)

	StrId := slug["id"].(string)
	id, err := utils.ToInt64(StrId)
	if err != nil {
		tx.Rollback()
		return utils.SendError(err, http.StatusBadRequest)
	}

	var payload payloads.UpdatePasswordRespondent

	err = utils.DynamicBind(req, &payload)
	if err != nil {
		tx.Rollback()
		return utils.SendError(err, http.StatusBadRequest)
	}

	var validate = validator.New()

	validate.RegisterValidation("password_rule", utils.PasswordRuleValidation)

	err = validate.Struct(payload)
	if err != nil {
		tx.Rollback()
		for _, err := range err.(validator.ValidationErrors) {
			customErrorMsg := utils.TranslateError(err)
			return utils.SendError(errors.New(customErrorMsg), http.StatusBadRequest)
		}
	}

	oldRespondent, err := service.respondentRepo.GetRespondentById(int(id))
	if err != nil {
		if err.Error() != gorm.ErrRecordNotFound.Error() {
			tx.Rollback()
			return utils.SendError(errors.New("Terjadi kesalahan pada server, Silakan coba lagi nanti"), http.StatusNotFound)
		}
	}

	if oldRespondent == nil {
		tx.Rollback()
		return utils.SendError(errors.New("Responden tidak ditemukan"), http.StatusNotFound)
	}

	if oldRespondent.RoleID != 5 && oldRespondent.RoleID != 4 && oldRespondent.RoleID != 3 && oldRespondent.RoleID != 2 {
		tx.Rollback()
		return utils.SendError(errors.New("Anda tidak memiliki izin untuk mengubah password responden ini"), http.StatusForbidden)
	}

	oldUser, err := service.usersRepo.GetUserByRespondentId(int(id))
	if err != nil {
		if err.Error() != gorm.ErrRecordNotFound.Error() {
			tx.Rollback()
			return utils.SendError(errors.New("Terjadi kesalahan pada server, Silakan coba lagi nanti"), http.StatusNotFound)
		}
	}

	if oldUser == nil {
		tx.Rollback()
		return utils.SendError(errors.New("Pengguna tidak ditemukan"), http.StatusNotFound)
	}

	var hashedPassword string
	hashedBytes, _ := bcrypt.GenerateFromPassword([]byte(payload.PasswordBaru), bcrypt.DefaultCost)
	hashedPassword = string(hashedBytes)

	oldUser.Password = &hashedPassword

	_, err = service.usersRepo.UpdateUser(oldUser)
	if err != nil {
		tx.Rollback()
		return utils.SendError(errors.New("Terjadi kesalahan pada server, Silakan coba lagi nanti"), http.StatusInternalServerError)
	}

	err = utils.SaveLogActivities("User", "respondent", "PUT", int(usr.ID), string(usr.Name), "-", "Mengubah Password Responden dengan ID "+StrId)
	if err != nil {
		tx.Rollback()
		return utils.SendError(err, http.StatusInternalServerError)
	}

	err = tx.Commit().Error
	if err != nil {
		tx.Rollback()
		return utils.SendError(err, http.StatusInternalServerError)
	}

	return utils.SendData("Password berhasil diperbarui")
}

func (service *respondentService) RespondentOptions(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()

	page, err := strconv.Atoi(param.Get("page"))
	if err != nil || page <= 0 {
		page = 1
	}

	limit, err := strconv.Atoi(param.Get("limit"))
	if err != nil || limit <= 0 {
		limit = 1000
	}

	var respondentIds []string
	if len(param["id[]"]) > 0 {
		respondentIds = param["id[]"]
	} else if len(param["id"]) > 0 {
		respondentIds = param["id"]
	}

	_req := payloads.RespondentOptionsPayload{
		Q:           param.Get("q"),
		Page:        page,
		Limit:       limit,
		IDs:         respondentIds,
		TipeWilayah: param.Get("tipe_wilayah"),
		KecamatanId: param.Get("kecamatan_id"),
		KelurahanId: param.Get("kelurahan_id"),
		RWId:        param.Get("rw_id"),
		RTId:        param.Get("rt_id"),
		Status:      param.Get("status"),
		HasJabatan:  param.Get("has_jabatan"),
	}

	data, totalData, err := service.respondentRepo.RespondentOptions(_req)
	if err != nil {
		return utils.SendError(err, http.StatusInternalServerError)
	}

	currentTotalLoaded := (page-1)*limit + len(data)
	hasMore := int64(currentTotalLoaded) < totalData

	// MENGGUNAKAN STRUCT BARU
	responseData := response.OptionsResponse{
		Options: data,
		Meta: response.PaginationMeta{
			CurrentPage: page,
			PerPage:     limit,
			Total:       totalData,
			HasMore:     hasMore,
		},
	}

	return utils.SendData(responseData, "Berhasil mengambil opsi responden")
}
