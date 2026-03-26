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

type RespondentService interface {
	GetRespondent(usr models.JwtCustomClaims, param url.Values) (*pb.ProxyResponse, error)
	GetDetailRespondent(slug map[string]interface{}) (*pb.ProxyResponse, error)
	DeleteRespondent(slug map[string]interface{}) (*pb.ProxyResponse, error)
	UpdateRespondent(slug map[string]interface{}, req map[string]interface{}) (*pb.ProxyResponse, error)
}

type respondentService struct {
	respondentRepo repository.RespondentRepo
}

func NewRespondentService(
	respondentRepo repository.RespondentRepo,

) RespondentService {
	return &respondentService{
		respondentRepo,
	}
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
	return utils.SendData(updateData)
	err = service.respondentRepo.UpdateUsers(int(idInt), updateData)
	if err != nil {
		return utils.SendError(err, http.StatusInternalServerError)
	}
	return utils.SendData("Data Berhasil Diperbarui")
}
