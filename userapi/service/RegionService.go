package service

import (
	"backend/siccore/pb"
	"backend/userapi/repository"
	"backend/userapi/utils"
	"fmt"
	"net/http"
	"net/url"
)

type RegionService interface {
	KecamatanOptions(param url.Values) (*pb.ProxyResponse, error)
	KelurahansOptions(param url.Values) (*pb.ProxyResponse, error)
	RwOptions(param url.Values) (*pb.ProxyResponse, error)
	RtOptions(param url.Values) (*pb.ProxyResponse, error)
}

type regionService struct {
	regionRepo repository.RegionRepo
}

func NewRegionService(
	regionRepo repository.RegionRepo,

) RegionService {
	return &regionService{
		regionRepo,
	}
}

func (service *regionService) KecamatanOptions(param url.Values) (*pb.ProxyResponse, error) {
	_, _, _, search, err := utils.SetPagination(param)
	if err != nil {
		return utils.SendError(fmt.Errorf("terjadi kesalahan saat memproses data parameter: %w", err), http.StatusBadRequest)
	}

	dataOptions, err := service.regionRepo.KecamatanOptions(search)
	if err != nil {
		return utils.SendError(fmt.Errorf("Gagal Mendapatkan Data Kecamatan"), http.StatusInternalServerError)
	}

	return utils.SendData(dataOptions)
}

func (service *regionService) KelurahansOptions(param url.Values) (*pb.ProxyResponse, error) {
	_, _, _, search, err := utils.SetPagination(param)
	if err != nil {
		return utils.SendError(fmt.Errorf("terjadi kesalahan saat memproses data parameter: %w", err), http.StatusBadRequest)
	}
	id := "0"
	if param.Get("idKecamatan") != "" {
		id = param.Get("idKecamatan")
	} else {
		return utils.SendError(fmt.Errorf("ID Tidak Ditemukan"), http.StatusBadRequest)
	}

	idInt, err := utils.ToInt64(id)
	if err != nil {
		return utils.SendError(fmt.Errorf("Id Kecamatan Tidak Valid"), http.StatusBadRequest)
	}

	dataOptions, err := service.regionRepo.KelurahanOptions(search, int(idInt))
	if err != nil {
		return utils.SendError(fmt.Errorf("Gagal Mendapatkan Data Kelurahan"), http.StatusInternalServerError)
	}

	return utils.SendData(dataOptions)
}

func (service *regionService) RwOptions(param url.Values) (*pb.ProxyResponse, error) {
	_, _, _, search, err := utils.SetPagination(param)
	if err != nil {
		return utils.SendError(fmt.Errorf("terjadi kesalahan saat memproses data parameter: %w", err), http.StatusBadRequest)
	}
	id := "0"
	if param.Get("idKelurahan") != "" {
		id = param.Get("idKelurahan")
	} else {
		return utils.SendError(fmt.Errorf("ID Tidak Ditemukan"), http.StatusBadRequest)
	}

	idInt, err := utils.ToInt64(id)
	if err != nil {
		return utils.SendError(fmt.Errorf("Id Kelurahan Tidak Valid"), http.StatusBadRequest)
	}

	dataOptions, err := service.regionRepo.RwOptions(search, int(idInt))
	if err != nil {
		return utils.SendError(fmt.Errorf("Gagal Mendapatkan Data RW"), http.StatusInternalServerError)
	}

	return utils.SendData(dataOptions)
}

func (service *regionService) RtOptions(param url.Values) (*pb.ProxyResponse, error) {
	_, _, _, search, err := utils.SetPagination(param)
	if err != nil {
		return utils.SendError(fmt.Errorf("terjadi kesalahan saat memproses data parameter: %w", err), http.StatusBadRequest)
	}
	id := "0"
	if param.Get("idRw") != "" {
		id = param.Get("idRw")
	} else {
		return utils.SendError(fmt.Errorf("ID Tidak Ditemukan"), http.StatusBadRequest)
	}

	idInt, err := utils.ToInt64(id)
	if err != nil {
		return utils.SendError(fmt.Errorf("Id Rw Tidak Valid"), http.StatusBadRequest)
	}

	dataOptions, err := service.regionRepo.RtOptions(search, int(idInt))
	if err != nil {
		return utils.SendError(fmt.Errorf("Gagal Mendapatkan Data RT"), http.StatusInternalServerError)
	}

	return utils.SendData(dataOptions)
}
