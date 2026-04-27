package service

import (
	"backend/siccore/pb"
	"backend/userapi/models"
	"backend/userapi/payloads"
	"backend/userapi/repository"
	"backend/userapi/utils"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"net/url"
	"os"
	"time"

	excelize "github.com/xuri/excelize/v2"
	"gorm.io/gorm"
)

type RespondentService interface {
	CreateRespondent(req map[string]interface{}, usr models.JwtCustomClaims) (*pb.ProxyResponse, error)
	GetRespondent(usr models.JwtCustomClaims, param url.Values) (*pb.ProxyResponse, error)
	GetDetailRespondent(slug map[string]interface{}) (*pb.ProxyResponse, error)
	DeleteRespondent(slug map[string]interface{}) (*pb.ProxyResponse, error)
	UpdateRespondent(slug map[string]interface{}, req map[string]interface{}) (*pb.ProxyResponse, error)
	GetExampleImport() (*pb.ProxyResponse, error)
	ImportRespondent(req map[string]interface{}) (*pb.ProxyResponse, error)
	GetRawDetailRespondent(slug map[string]interface{}) (*pb.ProxyResponse, error)
	SurveyorOption() (*pb.ProxyResponse, error)
	BlockRespondent(usr models.JwtCustomClaims, param url.Values) (*pb.ProxyResponse, error)
}

type respondentService struct {
	respondentRepo repository.RespondentRepo
	usersRepo      repository.UsersRepo
}

func NewRespondentService(
	respondentRepo repository.RespondentRepo,
	usersRepo repository.UsersRepo,

) RespondentService {
	return &respondentService{
		respondentRepo,
		usersRepo,
	}
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
		if dataPengguna.Kecamatan != "" {
			int64Kecamatan, err := utils.ToInt64(dataPengguna.Kecamatan)
			if err != nil {
				tx.Rollback()
				return utils.SendError(fmt.Errorf("Gagal Mendapatkan Data Kecamatan"), http.StatusInternalServerError)
			}
			intKecamatan := int(int64Kecamatan)
			dataRespondent.Kecamatan = &intKecamatan
		}
		if dataPengguna.Kelurahan != "" {
			int64Kelurahan, err := utils.ToInt64(dataPengguna.Kelurahan)
			if err != nil {
				tx.Rollback()
				return utils.SendError(fmt.Errorf("Gagal Mendapatkan Data Kelurahan"), http.StatusInternalServerError)
			}
			intKelurahan := int(int64Kelurahan)
			dataRespondent.Kelurahan = &intKelurahan
		}
		if dataPengguna.RW != "" {
			int64RW, err := utils.ToInt64(dataPengguna.RW)
			if err != nil {
				tx.Rollback()
				return utils.SendError(fmt.Errorf("Gagal Mendapatkan Data RW"), http.StatusInternalServerError)
			}
			intRW := int(int64RW)
			dataRespondent.RW = &intRW
		}
		if dataPengguna.RT != "" {
			int64RT, err := utils.ToInt64(dataPengguna.RT)
			if err != nil {
				tx.Rollback()
				return utils.SendError(fmt.Errorf("Gagal Mendapatkan Data RT"), http.StatusInternalServerError)
			}
			intRT := int(int64RT)
			dataRespondent.RT = &intRT
		}

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

func (service *respondentService) DeleteRespondent(slug map[string]interface{}) (*pb.ProxyResponse, error) {
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
	return utils.SendData("Data Berhasil Dihapus")
}

func (service *respondentService) UpdateRespondent(slug map[string]interface{}, req map[string]interface{}) (*pb.ProxyResponse, error) {
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

func (service *respondentService) ImportRespondent(req map[string]interface{}) (*pb.ProxyResponse, error) {
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

func (service *respondentService) SurveyorOption() (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()

	data, err := service.respondentRepo.SurveyorOption()
	if err != nil {
		return utils.SendError(fmt.Errorf("Gagal Mendapatkan Data Surveyor"), http.StatusInternalServerError)
	}

	return utils.SendData(data)
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

	data.ID = int(intRespondentId)
	data.IsBlocked = blockStatus
	data.UpdatedAt = utils.TimeNow()

	err = service.respondentRepo.RespondentBlock(data)
	if err != nil {
		return utils.SendError(fmt.Errorf("Terjadi Kesalahan Saat Melakukan Block Respondent"), http.StatusInternalServerError)
	}

	return utils.SendData(nil)
}
