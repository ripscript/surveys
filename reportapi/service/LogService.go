package service

import (
	"backend/reportapi/models"
	"backend/reportapi/payloads"
	"backend/reportapi/repository"
	"backend/reportapi/utils"
	"backend/siccore/pb"
	"fmt"
	"math"
	"net/http"
	"net/url"
)

type LogService interface {
	SaveLogActivities(req map[string]interface{}) (*pb.ProxyResponse, error)
	GetLogActivities(param url.Values) (*pb.ProxyResponse, error)
}

type logService struct {
	logRepo repository.LogRepo
}

func NewLogService(
	logRepo repository.LogRepo,

) LogService {
	return &logService{
		logRepo,
	}
}

func (service *logService) SaveLogActivities(req map[string]interface{}) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()

	payload := payloads.LogPayload{}

	err := utils.DynamicBind(req, &payload)
	if err != nil {
		return utils.SendError(err, http.StatusInternalServerError)
	}

	model := models.LogActivitySave{}

	err = utils.DynamicBind(payload, &model)
	if err != nil {
		return utils.SendError(err, http.StatusInternalServerError)
	}

	model.CreatedAt = utils.TimeNow()

	err = service.logRepo.SaveLogActivities(model)
	if err != nil {
		return utils.SendError(err, http.StatusInternalServerError)
	}

	return utils.SendData("Log berhasil Disimpan")
}

func (service *logService) GetLogActivities(param url.Values) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()

	page, limit, offset, _, err := utils.SetPagination(param)
	if err != nil {
		return utils.SendError(fmt.Errorf("terjadi kesalahan saat memproses pagination: %w", err), http.StatusBadRequest)
	}

	data, total, err := service.logRepo.GetLogActivities(offset, limit, param)
	if err != nil {
		return utils.SendError(err, http.StatusInternalServerError)
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
		"data": data,
		"meta": meta,
	}
	return utils.SendData(response)
}
