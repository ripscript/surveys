package service

import (
	"backend/siccore/pb"
	"backend/userapi/models"
	"backend/userapi/payloads"
	"backend/userapi/repository"
	"backend/userapi/utils"
	"fmt"
	"math"
	"net/http"
	"net/url"
	"time"
)

type UsersService interface {
	GetUsers(usr models.JwtCustomClaims, param url.Values) (*pb.ProxyResponse, error)
	GetDetailUsers(slug map[string]interface{}) (*pb.ProxyResponse, error)
	UpdateUsers(slug map[string]interface{}, req map[string]interface{}) (*pb.ProxyResponse, error)
	DeleteUsers(slug map[string]interface{}) (*pb.ProxyResponse, error)
	ResetPassword(slug map[string]interface{}) (*pb.ProxyResponse, error)
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
	updateData.UpdatedAt = time.Now()

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
	updateData.DeletedAt = time.Now()

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
