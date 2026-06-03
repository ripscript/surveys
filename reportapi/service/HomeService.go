package service

import (
	"backend/reportapi/repository"
	"backend/reportapi/utils"
	"backend/siccore/pb"
	"net/http"
)

type HomeService interface {
	CountKecamatan() (*pb.ProxyResponse, error)
	CountKelurahan() (*pb.ProxyResponse, error)
	CountRw() (*pb.ProxyResponse, error)
	CountRt() (*pb.ProxyResponse, error)
	CountSurveyOngoing() (*pb.ProxyResponse, error)
	CountSurveyUpcoming() (*pb.ProxyResponse, error)
	CountSurveyFinished() (*pb.ProxyResponse, error)
}

type homeService struct {
	homeRepo repository.HomeRepo
}

func NewHomeService(
	homeRepo repository.HomeRepo,

) HomeService {
	return &homeService{
		homeRepo,
	}
}

func (service *homeService) CountKecamatan() (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()

	CountKecamatan, err := service.homeRepo.CountKecamatan()
	if err != nil {
		return utils.SendError(err, http.StatusInternalServerError)
	}

	return utils.SendData(CountKecamatan)
}

func (service *homeService) CountKelurahan() (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()

	CountKecamatan, err := service.homeRepo.CountKelurahan()
	if err != nil {
		return utils.SendError(err, http.StatusInternalServerError)
	}

	return utils.SendData(CountKecamatan)
}

func (service *homeService) CountRw() (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()

	CountKecamatan, err := service.homeRepo.CountRw()
	if err != nil {
		return utils.SendError(err, http.StatusInternalServerError)
	}

	return utils.SendData(CountKecamatan)
}

func (service *homeService) CountRt() (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()

	CountKecamatan, err := service.homeRepo.CountRt()
	if err != nil {
		return utils.SendError(err, http.StatusInternalServerError)
	}

	return utils.SendData(CountKecamatan)
}

func (service *homeService) CountSurveyOngoing() (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()

	CountKecamatan, err := service.homeRepo.CountSurveyOngoing()
	if err != nil {
		return utils.SendError(err, http.StatusInternalServerError)
	}

	return utils.SendData(CountKecamatan)
}

func (service *homeService) CountSurveyUpcoming() (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()

	CountKecamatan, err := service.homeRepo.CountSurveyUpcoming()
	if err != nil {
		return utils.SendError(err, http.StatusInternalServerError)
	}

	return utils.SendData(CountKecamatan)
}

func (service *homeService) CountSurveyFinished() (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()

	CountKecamatan, err := service.homeRepo.CountSurveyFinished()
	if err != nil {
		return utils.SendError(err, http.StatusInternalServerError)
	}

	return utils.SendData(CountKecamatan)
}
