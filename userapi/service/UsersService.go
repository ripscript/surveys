package service

import (
	"backend/siccore/pb"
	"backend/userapi/enums"
	"backend/userapi/models"
	"backend/userapi/payloads"
	"backend/userapi/repository"
	"backend/userapi/transaction"
	"backend/userapi/utils"
	"bytes"
	"context"
	"errors"
	"fmt"
	"math"
	"net/http"
	"net/url"
	"os"
	"slices"
	"strconv"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/speps/go-hashids/v2"
	excelize "github.com/xuri/excelize/v2"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type UsersService interface {
	GetProfile(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	UpdateProfileBundle(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims) (*pb.ProxyResponse, error)
	GetUsers(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	GetDetailUsers(slug map[string]interface{}) (*pb.ProxyResponse, error)
	UpdateUsers(slug map[string]interface{}, req map[string]interface{}) (*pb.ProxyResponse, error)
	DeleteUsers(slug map[string]interface{}) (*pb.ProxyResponse, error)
	ResetPassword(slug map[string]interface{}) (*pb.ProxyResponse, error)
	UserExport() (*pb.ProxyResponse, error)
	CreateUsers(req map[string]interface{}, usr models.JwtCustomClaims) (*pb.ProxyResponse, error)
}

type usersService struct {
	usersRepo      repository.UsersRepo
	respondentRepo repository.RespondentRepo
	txManager      transaction.TxManager
	fileRepo       repository.FileRepo
	wilayahRepo    repository.WilayahRepo
}

func NewUsersService(
	usersRepo repository.UsersRepo,
	respondentRepo repository.RespondentRepo,
	txManager transaction.TxManager,
	fileRepo repository.FileRepo,
	wilayahRepo repository.WilayahRepo,

) UsersService {
	return &usersService{
		usersRepo,
		respondentRepo,
		txManager,
		fileRepo,
		wilayahRepo,
	}
}

func (service *usersService) GetProfile(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()

	userId := usr.ID

	user, err := service.usersRepo.GetUserRawById(int(userId))
	if err != nil {
		if err.Error() != gorm.ErrRecordNotFound.Error() {
			return utils.SendError(errors.New("Terjadi kesalahan pada server, coba lagi nanti"), http.StatusNotFound)
		}
	}

	if user == nil {
		return utils.SendError(errors.New("Pengguna tidak ditemukan"), http.StatusNotFound)
	}

	hd := hashids.NewData()
	hd.Salt = os.Getenv("HASHID_SALT")
	hd.MinLength = 24

	h, err := hashids.NewWithData(hd)
	if err != nil {
		return utils.SendError(err, http.StatusInternalServerError)
	}

	var wilayahCodeStr *string

	var kecamatanCodeStr *string
	if user.Respondent != nil && user.Respondent.KecamatanID != nil {
		kecamatanId := []int{int(*user.Respondent.KecamatanID)}

		kecamatanCode, err := h.Encode(kecamatanId)
		if err != nil {
			return utils.SendError(err, http.StatusInternalServerError)
		}

		kecamatanCodeStr = &kecamatanCode

		wc, err := h.Encode([]int{int(user.Respondent.RoleID), int(*user.Respondent.KecamatanID)})
		if err != nil {
			return utils.SendError(err, http.StatusInternalServerError)
		}
		wilayahCodeStr = &wc
	}

	if user.Respondent != nil {
		user.Respondent.KecamatanCode = kecamatanCodeStr
	}

	var kelurahanCodeStr *string
	if user.Respondent != nil && user.Respondent.KelurahanID != nil {
		kelurahanId := []int{int(*user.Respondent.KelurahanID)}

		kelurahanCode, err := h.Encode(kelurahanId)
		if err != nil {
			return utils.SendError(err, http.StatusInternalServerError)
		}

		kelurahanCodeStr = &kelurahanCode

		wckel, err := h.Encode([]int{int(user.Respondent.RoleID), int(*user.Respondent.KelurahanID)})
		if err != nil {
			return utils.SendError(err, http.StatusInternalServerError)
		}
		wilayahCodeStr = &wckel
	}

	if user.Respondent != nil {
		user.Respondent.KelurahanCode = kelurahanCodeStr
	}

	var RwCodeStr *string
	if user.Respondent != nil && user.Respondent.RwID != nil {
		rwId := []int{int(*user.Respondent.RwID)}

		rwCode, err := h.Encode(rwId)
		if err != nil {
			return utils.SendError(err, http.StatusInternalServerError)
		}

		RwCodeStr = &rwCode

		wcrw, err := h.Encode([]int{int(user.Respondent.RoleID), int(*user.Respondent.RwID)})
		if err != nil {
			return utils.SendError(err, http.StatusInternalServerError)
		}
		wilayahCodeStr = &wcrw
	}

	if user.Respondent != nil {
		user.Respondent.RwCode = RwCodeStr
	}

	if user.Respondent != nil && user.Respondent.Avatar != nil {
		user.Respondent.Avatar = utils.StringToPointer(os.Getenv("API_GATEWAY_URL") + "/view-foto-profil/" + *user.Respondent.Avatar)
	}

	var RtCodeStr *string
	if user.Respondent != nil && user.Respondent.RtID != nil {
		rtId := []int{int(*user.Respondent.RtID)}

		rtCode, err := h.Encode(rtId)
		if err != nil {
			return utils.SendError(err, http.StatusInternalServerError)
		}

		RtCodeStr = &rtCode

		wcrt, err := h.Encode([]int{int(user.Respondent.RoleID), int(*user.Respondent.RtID)})
		if err != nil {
			return utils.SendError(err, http.StatusInternalServerError)
		}
		wilayahCodeStr = &wcrt
	}

	if user.Respondent != nil {
		user.Respondent.RtCode = RtCodeStr
		user.Respondent.WilayahCode = wilayahCodeStr
	}

	if user.Respondent == nil {
		return utils.SendError(errors.New("Data responden pengguna tidak ditemukan"), http.StatusNotFound)
	}

	roleID := user.Respondent.RoleID

	filter := payloads.FilterWilayah{}

	if roleID == int64(enums.ROLE_KECAMATAN) && user.Respondent.KecamatanID != nil {
		kecID := int(*user.Respondent.KecamatanID)
		filter.KecamatanID = &kecID

		getGeoName, err := service.wilayahRepo.GetGeoNameByKecamatanId(*user.Respondent.KecamatanID)
		if err != nil {
			return utils.SendError(err, http.StatusInternalServerError)
		}

		user.GeoName = &getGeoName
	} else if roleID == int64(enums.ROLE_KELURAHAN) && user.Respondent.KelurahanID != nil {
		kelID := int(*user.Respondent.KelurahanID)
		filter.KelurahanID = &kelID

		getGeoName, err := service.wilayahRepo.GetGeoNameByKelurahanId(int64(kelID))
		if err != nil {
			return utils.SendError(err, http.StatusInternalServerError)
		}

		user.GeoName = &getGeoName
	} else if roleID == int64(enums.ROLE_RW) && user.Respondent.RwID != nil {
		rwID := int(*user.Respondent.RwID)
		filter.RwID = &rwID
	}

	isLevelKota := roleID == int64(enums.ROLE_WALIKOTA) || roleID == int64(enums.ROLE_PEMERINTAH_KOTA) || roleID == int64(enums.ROLE_ADMIN)
	isLevelKecamatan := isLevelKota || roleID == int64(enums.ROLE_KECAMATAN)
	isLevelKelurahan := isLevelKecamatan || roleID == int64(enums.ROLE_KELURAHAN)
	isLevelRW := isLevelKelurahan || roleID == int64(enums.ROLE_RW)

	if isLevelKota {
		totalKec, err := service.usersRepo.GetTotalKecamatan(filter)
		if err != nil {
			return utils.SendError(err, http.StatusInternalServerError)
		}

		if totalKec == nil {
			user.TotalWilayahKecamatan = 0
		} else {
			user.TotalWilayahKecamatan = *totalKec
		}
	}

	if isLevelKecamatan {
		totalKel, err := service.usersRepo.GetTotalKelurahan(filter)
		if err != nil {
			return utils.SendError(err, http.StatusInternalServerError)
		}
		if totalKel == nil {
			user.TotalWilayahKelurahan = 0
		} else {
			user.TotalWilayahKelurahan = *totalKel
		}
	}

	if isLevelKelurahan {
		totalRW, err := service.usersRepo.GetTotalRW(filter)
		if err != nil {
			return utils.SendError(err, http.StatusInternalServerError)
		}
		if totalRW == nil {
			user.TotalWilayahRW = 0
		} else {
			user.TotalWilayahRW = *totalRW
		}
	}

	if isLevelRW {
		totalRT, err := service.usersRepo.GetTotalRT(filter)
		if err != nil {
			return utils.SendError(err, http.StatusInternalServerError)
		}
		if totalRT == nil {
			user.TotalWilayahRT = 0
		} else {
			user.TotalWilayahRT = *totalRT
		}
	}

	return utils.SendData(user, "Data Profil Pengguna Berhasil Ditemukan")
}

func (service *usersService) GetUsers(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()
	page, limit, offset, _, err := utils.SetPagination(param)
	if err != nil {
		return utils.SendError(fmt.Errorf("terjadi kesalahan saat memproses pagination: %w", err), http.StatusBadRequest)
	}
	users, total, err := service.usersRepo.GetUsers(offset, limit, param)
	if err != nil || users == nil {
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
		"data": users,
		"meta": meta,
	}
	return utils.SendData(response)
}

func (service *usersService) GetDetailUsers(slug map[string]interface{}) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()
	id := slug["id"].(string)
	idInt, err := utils.ToInt64(id)
	if err != nil {
		return utils.SendError(err, http.StatusInternalServerError)
	}
	detailUsers, err := service.usersRepo.GetDetailUsers(int(idInt))
	if err != nil {
		return utils.SendError(fmt.Errorf("data ditemukan: %w", err), http.StatusNotFound)
	}
	return utils.SendData(detailUsers)
}

func (service *usersService) UpdateUsers(slug map[string]interface{}, req map[string]interface{}) (*pb.ProxyResponse, error) {
	tx := service.txManager.Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}

	defer utils.GeneralRecover()
	id := slug["id"].(string)
	idInt, err := utils.ToInt64(id)
	if err != nil {
		return utils.SendError(err, http.StatusInternalServerError)
	}
	_, err = service.usersRepo.GetDetailUsers(int(idInt))
	if err != nil {
		return utils.SendError(fmt.Errorf("data ditemukan: %w", err), http.StatusNotFound)
	}

	payload := payloads.UpdateRespondent{}

	err = utils.DynamicBind(req, &payload)
	if err != nil {
		return utils.SendError(err, http.StatusBadRequest)
	}

	var validate = validator.New()

	err = validate.Struct(payload)
	if err != nil {
		for _, err := range err.(validator.ValidationErrors) {
			customErrorMsg := utils.TranslateError(err)
			return utils.SendError(errors.New(customErrorMsg), http.StatusBadRequest)
		}
	}

	role := enums.RoleID(payload.RoleID)
	if !role.IsValid() {
		return utils.SendError(errors.New("Role ID tidak valid"), 400)
	}

	if role != enums.ROLE_ADMIN && role != enums.ROLE_SURVEYOR {
		return utils.SendError(errors.New("Role ID tidak valid, hanya diperbolehkan admin dan surveyor"), 400)
	}

	first, last := utils.SplitFullName(payload.Name)
	if last == "" {
		last = first
	}

	var firstName, lastName *string
	if first != "" {
		str := strings.ToLower(first)
		firstName = &str
	}
	if last != "" {
		str := strings.ToLower(last)
		lastName = &str
	}

	oldRespondent, err := service.respondentRepo.GetRespondentById(int(idInt))
	if err != nil {
		if err.Error() != gorm.ErrRecordNotFound.Error() {
			return utils.SendError(errors.New("Terjadi kesalahan pada server, Silakan coba lagi nanti"), http.StatusInternalServerError)
		}
	}

	if oldRespondent == nil {
		return utils.SendError(errors.New("Responden tidak ditemukan"), http.StatusNotFound)
	}

	if oldRespondent.PhoneNumber != nil {
		if *oldRespondent.PhoneNumber != payload.PhoneNumber {
			checkPhoneNumberRespondent, err := service.usersRepo.CheckPhoneNumber(payload.PhoneNumber)
			if err != nil {
				if err.Error() != gorm.ErrRecordNotFound.Error() {
					return utils.SendError(err, http.StatusInternalServerError)
				}
			}
			if checkPhoneNumberRespondent > 0 {
				return utils.SendError(errors.New("Nomor telepon sudah digunakan"), http.StatusConflict)
			}
		}
	}

	if oldRespondent.Email != nil {
		if *oldRespondent.Email != payload.Email {
			checkEmailRespondent, checkEmailUsers, err := service.usersRepo.CheckEmail(payload.Email)
			if err != nil {
				return utils.SendError(err, http.StatusInternalServerError)
			}
			if checkEmailRespondent != 0 || checkEmailUsers != 0 {
				return utils.SendError(fmt.Errorf("Email Sudah Digunakan"), http.StatusBadRequest)
			}
		}
	}

	oldRespondent.Name = payload.Name
	oldRespondent.PhoneNumber = &payload.PhoneNumber
	oldRespondent.Email = &payload.Email
	var roleID int64
	if payload.RoleID != 0 {
		roleIDInt64 := int64(payload.RoleID)
		roleID = roleIDInt64
	}
	oldRespondent.RoleID = roleID
	oldRespondent.UpdatedAt = utils.TimeNowPointer()

	userRepoTx := service.usersRepo.WithTx(tx)
	respondentRepoTx := service.respondentRepo.WithTx(tx)

	_, err = respondentRepoTx.UpdateRespondent(oldRespondent)
	if err != nil {
		tx.Rollback()

		return utils.SendError(err, http.StatusInternalServerError)
	}

	oldUser, err := service.usersRepo.GetUserByRespondentId(int(oldRespondent.ID))
	if err != nil {
		if err.Error() != gorm.ErrRecordNotFound.Error() {
			if err.Error() != gorm.ErrRecordNotFound.Error() {
				return utils.SendError(errors.New("Terjadi kesalahan pada server, Silakan coba lagi nanti"), http.StatusInternalServerError)
			}
		}
	}

	if oldUser == nil {
		return utils.SendError(errors.New("User tidak ditemukan"), http.StatusNotFound)
	}

	oldUser.FirstName = firstName
	oldUser.LastName = lastName
	oldUser.Email = &payload.Email
	oldUser.UpdatedAt = utils.TimeNowPointer()

	_, err = userRepoTx.UpdateUser(oldUser)
	if err != nil {
		tx.Rollback()

		return utils.SendError(err, http.StatusInternalServerError)
	}

	err = tx.Commit().Error
	if err != nil {
		return utils.SendError(err, http.StatusInternalServerError)
	}

	return utils.SendData("Data Berhasil Diperbarui")
}

func (service *usersService) DeleteUsers(slug map[string]interface{}) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()
	id := slug["id"].(string)
	idInt, err := utils.ToInt64(id)
	if err != nil {
		return utils.SendError(err, http.StatusInternalServerError)
	}
	_, err = service.usersRepo.GetDetailUsers(int(idInt))
	if err != nil {
		return utils.SendError(fmt.Errorf("data ditemukan: %w", err), http.StatusNotFound)
	}

	updateData := models.DeleteRespondent{}
	updateData.Id = int(idInt)
	updateData.DeletedAt = utils.TimeNow()

	err = service.usersRepo.DeleteUsers(updateData)
	if err != nil {
		return utils.SendError(err, http.StatusInternalServerError)
	}
	return utils.SendData("Data Berhasil Dihapus")
}

func (service *usersService) ResetPassword(slug map[string]interface{}) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()
	id := slug["id"].(string)
	idInt, err := utils.ToInt64(id)
	if err != nil {
		return utils.SendError(err, http.StatusInternalServerError)
	}
	_, err = service.usersRepo.GetDetailUsers(int(idInt))
	if err != nil {
		return utils.SendError(fmt.Errorf("data ditemukan: %w", err), http.StatusNotFound)
	}

	err = service.usersRepo.ResetPasswordUsers(int(idInt))
	if err != nil {
		return utils.SendError(err, http.StatusInternalServerError)
	}
	return utils.SendData("Password Berhasil Direset")
}

func (service *usersService) UserExport() (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()

	users, err := service.usersRepo.UserExport()
	if err != nil {
		return utils.SendError(err, http.StatusInternalServerError)
	}

	f := excelize.NewFile()
	sheet := "Sheet1"
	f.SetSheetName("Sheet1", sheet)

	f.MergeCell(sheet, "A1", "C1")
	f.SetCellValue(sheet, "A1", "List Data Tiket Data Surveyor")

	titleStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold: true,
			Size: 14,
		},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
		},
	})

	f.SetCellStyle(sheet, "A1", "C1", titleStyle)
	f.SetRowHeight(sheet, 1, 28)

	headers := []string{"No", "Nama", "Email"}

	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold:  true,
			Color: "FFFFFF",
			Size:  11,
		},
		Fill: excelize.Fill{
			Type:    "solid",
			Color:   []string{"1F3A5F"},
			Pattern: 1,
		},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
		},
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
		},
	})

	for i, header := range headers {
		col := string('A' + i)
		cell := col + "2"
		f.SetCellValue(sheet, cell, header)
		f.SetCellStyle(sheet, cell, cell, headerStyle)
	}

	f.SetRowHeight(sheet, 2, 24)

	dataStyleWhite, _ := f.NewStyle(&excelize.Style{
		Fill: excelize.Fill{
			Type:    "solid",
			Color:   []string{"FFFFFF"},
			Pattern: 1,
		},
		Alignment: &excelize.Alignment{
			Horizontal: "left",
			Vertical:   "center",
		},
		Border: []excelize.Border{
			{Type: "left", Color: "D3D3D3", Style: 1},
			{Type: "top", Color: "D3D3D3", Style: 1},
			{Type: "bottom", Color: "D3D3D3", Style: 1},
			{Type: "right", Color: "D3D3D3", Style: 1},
		},
	})

	dataStyleGray, _ := f.NewStyle(&excelize.Style{
		Fill: excelize.Fill{
			Type:    "solid",
			Color:   []string{"D9E8F5"},
			Pattern: 1,
		},
		Alignment: &excelize.Alignment{
			Horizontal: "left",
			Vertical:   "center",
		},
		Border: []excelize.Border{
			{Type: "left", Color: "D3D3D3", Style: 1},
			{Type: "top", Color: "D3D3D3", Style: 1},
			{Type: "bottom", Color: "D3D3D3", Style: 1},
			{Type: "right", Color: "D3D3D3", Style: 1},
		},
	})

	maxName := len("Nama")
	maxEmail := len("Email")

	for i, user := range users {
		row := strconv.Itoa(i + 5)

		f.SetCellValue(sheet, "A"+row, i+1)
		f.SetCellValue(sheet, "B"+row, user.Name)
		f.SetCellValue(sheet, "C"+row, user.Email)

		f.SetRowHeight(sheet, i+5, 20)

		if len(user.Name) > maxName {
			maxName = len(user.Name)
		}
		if len(user.Email) > maxEmail {
			maxEmail = len(user.Email)
		}

		if i%2 == 0 {
			f.SetCellStyle(sheet, "A"+row, "C"+row, dataStyleWhite)
		} else {
			f.SetCellStyle(sheet, "A"+row, "C"+row, dataStyleGray)
		}
	}

	f.SetColWidth(sheet, "A", "A", 8)
	f.SetColWidth(sheet, "B", "B", float64(maxName+5))
	f.SetColWidth(sheet, "C", "C", float64(maxEmail+5))

	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		return utils.SendError(err, http.StatusInternalServerError)
	}

	fileBytes := buf.Bytes()

	return utils.SetResponseData(fileBytes, true, "Data File.xlsx", http.StatusOK, nil, ""), nil
}

func (service *usersService) CreateUsers(req map[string]interface{}, usr models.JwtCustomClaims) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()

	payload := payloads.UpdateRespondent{}

	err := utils.DynamicBind(req, &payload)
	if err != nil {
		return utils.SendError(fmt.Errorf("Form Tidak Sesuai"), http.StatusBadRequest)
	}

	var validate = validator.New()

	err = validate.Struct(payload)
	if err != nil {
		for _, err := range err.(validator.ValidationErrors) {
			customErrorMsg := utils.TranslateError(err)
			return utils.SendError(errors.New(customErrorMsg), http.StatusBadRequest)
		}
	}

	role := enums.RoleID(payload.RoleID)
	if !role.IsValid() {
		return utils.SendError(errors.New("Role ID tidak valid"), 400)
	}

	if role != enums.ROLE_ADMIN && role != enums.ROLE_SURVEYOR {
		return utils.SendError(errors.New("Role ID tidak valid, hanya diperbolehkan admin dan surveyor"), 400)
	}

	checkPhoneNumberRespondent, err := service.usersRepo.CheckPhoneNumber(payload.PhoneNumber)
	if err != nil {
		return utils.SendError(err, http.StatusInternalServerError)
	}

	if checkPhoneNumberRespondent != 0 {
		return utils.SendError(fmt.Errorf("Nomor Telepon Sudah Digunakan"), http.StatusBadRequest)
	}

	checkEmailRespondent, checkEmailUsers, err := service.usersRepo.CheckEmail(payload.Email)
	if err != nil {
		return utils.SendError(err, http.StatusInternalServerError)
	}

	if checkEmailRespondent != 0 || checkEmailUsers != 0 {
		return utils.SendError(fmt.Errorf("Email Sudah Digunakan"), http.StatusBadRequest)
	}

	dataUser := models.CreateRespondent{}
	err = utils.DynamicBind(payload, &dataUser)
	if err != nil {
		return utils.SendError(fmt.Errorf("Form Tidak Sesuai"), http.StatusBadRequest)
	}
	dataUser.BlkId = 1
	dataUser.CreatedAt = utils.TimeNow()
	dataUser.UpdatedAt = utils.TimeNow()

	err = service.usersRepo.StoreUsers(dataUser)
	if err != nil {
		return utils.SendError(err, http.StatusInternalServerError)
	}

	return utils.SendData(nil, "Data Pengguna Berhasil Dibuat")
}

func (service *usersService) UpdateProfileBundle(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()

	detachedCtx := context.WithoutCancel(ctx)

	availableMime := []string{"image/png", "image/jpg", "image/jpeg", "image/webp"}
	availablesExt := []string{".png", ".jpg", ".jpeg", ".webp", ".jfif"}
	maxSizeInKB := float64(5120)

	var uploadedFiles []string
	rollbackUploadedFiles := func() {
		if len(uploadedFiles) > 0 {
			service.fileRepo.DeleteFotoProfilBulk(detachedCtx, uploadedFiles)
		}
	}
	defer func() {
		if r := recover(); r != nil {
			rollbackUploadedFiles()
			panic(r)
		}
	}()

	var payload payloads.UpdateProfileBundlePayload

	err := utils.DynamicBind(req, &payload)
	if err != nil {
		return utils.SendError(err, http.StatusBadRequest)
	}

	var validate = validator.New()
	validate.RegisterValidation("password_rule", utils.PasswordRuleValidation)

	err = validate.Struct(payload)
	if err != nil {
		for _, valErr := range err.(validator.ValidationErrors) {
			customErrorMsg := utils.TranslateError(valErr)
			return utils.SendError(errors.New(customErrorMsg), http.StatusBadRequest)
		}
	}

	user, err := service.usersRepo.GetUserRawById(int(usr.ID))
	if err != nil {
		return utils.SendError(errors.New("Terjadi kesalahan pada server, coba lagi nanti"), http.StatusInternalServerError)
	}
	if user == nil {
		return utils.SendError(errors.New("Pengguna tidak ditemukan"), http.StatusNotFound)
	}

	if payload.Avatar != nil && *payload.Avatar != "" {
		var oldAvatar string
		if user.Respondent != nil && user.Respondent.Avatar != nil {
			oldAvatar = *user.Respondent.Avatar
		}

		if *payload.Avatar != oldAvatar {
			parsedURL, errParse := url.ParseRequestURI(*payload.Avatar)
			isURL := errParse == nil && parsedURL.Scheme != "" && parsedURL.Host != ""

			if isURL {
				payload.Avatar = &oldAvatar
			} else {
				if !strings.HasPrefix(*payload.Avatar, "data:") {
					rollbackUploadedFiles()
					return utils.SendError(errors.New("Format gambar baru tidak valid. Pastikan menggunakan format Data URI Base64"), http.StatusBadRequest)
				}

				if !strings.HasPrefix(*payload.Avatar, "data:image") {
					rollbackUploadedFiles()
					return utils.SendError(errors.New("Format file tidak didukung. Hanya menerima file gambar (PNG, JPG, WEBP)"), http.StatusBadRequest)
				}

				base64Data, err := utils.ExtractBase64Info(*payload.Avatar)
				if err != nil || base64Data == nil {
					rollbackUploadedFiles()
					return utils.SendError(errors.New("Gagal memproses gambar: Data Base64 tidak valid atau korup"), http.StatusBadRequest)
				}

				if !slices.Contains(availablesExt, base64Data.Extension) || !slices.Contains(availableMime, base64Data.MimeType) {
					rollbackUploadedFiles()
					return utils.SendError(errors.New("Format file gambar tidak didukung"), http.StatusBadRequest)
				}
				if base64Data.SizeInKB > maxSizeInKB {
					rollbackUploadedFiles()
					return utils.SendError(errors.New("Ukuran gambar tidak boleh melebihi 5MB"), http.StatusBadRequest)
				}

				path, err := service.fileRepo.UploadFotoProfil(ctx, payload.Avatar)
				if err != nil || path == nil {
					rollbackUploadedFiles()
					return utils.SendError(errors.New("Gagal mengunggah gambar"), http.StatusInternalServerError)
				}

				uploadedFiles = append(uploadedFiles, *path)
				payload.Avatar = path

				if oldAvatar != "" {
					_, _ = service.fileRepo.DeleteFotoProfilBulk(detachedCtx, []string{oldAvatar})
				}
			}
		}
	} else {
		if user.Respondent != nil {
			payload.Avatar = user.Respondent.Avatar
		}
	}

	if user.MustChangePassword != nil {
		if *user.MustChangePassword == false {
			if payload.NewPassword != "" && payload.CurrentPassword == "" {
				return utils.SendError(errors.New("Password saat ini harus diisi"), http.StatusBadRequest)
			}
		}
	}

	var respondentID int
	if user.RespondentID != nil {
		respondentID = int(*user.RespondentID)
	}

	err = service.usersRepo.CheckDuplicateProfileData(
		payload.Email,
		payload.NIK,
		payload.PhoneNumber,
		int(user.ID),
		respondentID,
	)
	if err != nil {
		return utils.SendError(err, http.StatusConflict)
	}

	var hashedPassword string
	if user.MustChangePassword != nil {
		if *user.MustChangePassword == true {
			hashedBytes, _ := bcrypt.GenerateFromPassword([]byte(payload.NewPassword), bcrypt.DefaultCost)
			hashedPassword = string(hashedBytes)
		} else {
			if payload.NewPassword != "" {
				var oldPassword string
				if user.Password != nil {
					oldPassword = *user.Password
				}

				errCompare := bcrypt.CompareHashAndPassword([]byte(oldPassword), []byte(payload.CurrentPassword))
				if errCompare != nil {
					return utils.SendError(errors.New("Password saat ini yang Anda masukkan salah"), http.StatusBadRequest)
				}
				hashedBytes, _ := bcrypt.GenerateFromPassword([]byte(payload.NewPassword), bcrypt.DefaultCost)
				hashedPassword = string(hashedBytes)
			}
		}
	}

	isPejabat := user.PejabatWilayah != nil

	err = service.usersRepo.UpdateProfileBundleTx(
		int(user.ID),
		respondentID,
		isPejabat,
		payload,
		hashedPassword,
	)
	if err != nil {
		return utils.SendError(err, http.StatusInternalServerError)
	}

	return utils.SendData(nil, "Profil berhasil diperbarui")
}
