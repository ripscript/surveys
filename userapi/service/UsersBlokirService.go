package service

import (
	"backend/siccore/pb"
	"backend/userapi/repository"
	"backend/userapi/utils"
	"fmt"
	"math"
	"net/http"
	"net/url"
)

type UsersBlokirService interface {
	GetListdata(param url.Values) (*pb.ProxyResponse, error)
	OpenBlokir(slug map[string]interface{}) (*pb.ProxyResponse, error)
}

type usersBlokirService struct {
	usersBlokirRepo repository.UsersBlokirRepo
}

func NewUsersBlokirService(
	usersBlokirRepo repository.UsersBlokirRepo,

) UsersBlokirService {
	return &usersBlokirService{
		usersBlokirRepo,
	}
}

func (service *usersBlokirService) GetListdata(param url.Values) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()
	page, limit, offset, _, err := utils.SetPagination(param)
	if err != nil {
		return utils.SendError(fmt.Errorf("terjadi kesalahan saat memproses pagination: %w", err), http.StatusBadRequest)
	}
	users, total, err := service.usersBlokirRepo.GetListData(offset, limit, param)
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

func (service *usersBlokirService) OpenBlokir(slug map[string]interface{}) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()
	id := slug["id"].(string)
	idInt, err := utils.ToInt64(id)
	if err != nil {
		return utils.SendError(err, http.StatusInternalServerError)
	}

	checkUser, err := service.usersBlokirRepo.CheckUser(int(idInt))
	if err != nil {
		return utils.SendError(fmt.Errorf("Gagal Melakukan Pengecekan Data Pengguna"), http.StatusInternalServerError)
	}

	if checkUser.IsBlocked != "true" {
		return utils.SendError(fmt.Errorf("Pengguna Tidak Terblokir"), http.StatusInternalServerError)
	}

	err = service.usersBlokirRepo.OpenBlokir(int(idInt))
	if err != nil {
		return utils.SendError(fmt.Errorf("Gagal Membuka Blokir"), http.StatusInternalServerError)
	}

	return utils.SendData(nil, "Blokir Berhasil Dibuka")
}
