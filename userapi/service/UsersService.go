package service

import (
	"backend/siccore/pb"
	"backend/userapi/models"
	"backend/userapi/payloads"
	"backend/userapi/repository"
	"backend/userapi/utils"
	"bytes"
	"fmt"
	"math"
	"net/http"
	"net/url"
	"strconv"

	excelize "github.com/xuri/excelize/v2"
)

type UsersService interface {
	GetUsers(usr models.JwtCustomClaims, param url.Values) (*pb.ProxyResponse, error)
	GetDetailUsers(slug map[string]interface{}) (*pb.ProxyResponse, error)
	UpdateUsers(slug map[string]interface{}, req map[string]interface{}) (*pb.ProxyResponse, error)
	DeleteUsers(slug map[string]interface{}) (*pb.ProxyResponse, error)
	ResetPassword(slug map[string]interface{}) (*pb.ProxyResponse, error)
	UserExport() (*pb.ProxyResponse, error)
	CreateUsers(req map[string]interface{}, usr models.JwtCustomClaims) (*pb.ProxyResponse, error)
}

type usersService struct {
	usersRepo repository.UsersRepo
}

func NewUsersService(
	usersRepo repository.UsersRepo,

) UsersService {
	return &usersService{
		usersRepo,
	}
}

func (service *usersService) GetUsers(usr models.JwtCustomClaims, param url.Values) (*pb.ProxyResponse, error) {
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

	updateData := models.UpdateRespondent{}
	err = utils.DynamicBind(payload, &updateData)
	if err != nil {
		return utils.SendError(err, http.StatusBadRequest)
	}
	updateData.Id = int(idInt)
	updateData.UpdatedAt = utils.TimeNow()

	err = service.usersRepo.UpdateUsers(int(idInt), updateData)
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

	checkEmailRespondent, checkEmailUsers, err := service.usersRepo.CheckEmail(payload.Email)
	if err != nil {
		return utils.SendError(err, http.StatusInternalServerError)
	}

	checkPhoneNumberRespondent, err := service.usersRepo.CheckPhoneNumber(payload.PhoneNumber)
	if err != nil {
		return utils.SendError(err, http.StatusInternalServerError)
	}

	if checkEmailRespondent != 0 || checkEmailUsers != 0 {
		return utils.SendError(fmt.Errorf("Email Sudah Digunakan"), http.StatusBadRequest)
	}

	if checkPhoneNumberRespondent != 0 {
		return utils.SendError(fmt.Errorf("Nomor Telepon Sudah Digunakan"), http.StatusBadRequest)
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
