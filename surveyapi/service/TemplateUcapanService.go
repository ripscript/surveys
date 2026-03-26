package service

import (
	"backend/siccore/pb"
	"backend/surveyapi/repository"
	"backend/surveyapi/utils"
	"net/http"
)

type TemplateUcapanService interface {
	GetGeneralTemplate(slug map[string]interface{}) (*pb.ProxyResponse, error)
}

type templateUcapanService struct {
	templateUcapanRepo repository.TemplateUcapanRepo
}

func NewTemplateUcapanService(
	templateUcapanRepo repository.TemplateUcapanRepo,
) TemplateUcapanService {
	return &templateUcapanService{
		templateUcapanRepo: templateUcapanRepo,
	}
}

func (service *templateUcapanService) GetGeneralTemplate(slug map[string]interface{}) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()
	StrId := slug["template_ucapan_id"]
	Id, err := utils.ToInt64(StrId)
	if err != nil {
		return utils.SendError(err, http.StatusBadRequest)
	}

	data, err := service.templateUcapanRepo.GetTemplateUcapanById(int(Id))
	if err != nil {
		return utils.SendError(err, http.StatusInternalServerError)
	}
	return utils.SendData(data, "Berhasil mengambil data")
}
