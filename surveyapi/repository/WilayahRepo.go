package repository

import (
	"backend/surveyapi/models"
	"backend/surveyapi/payloads"
	"backend/surveyapi/response"
	"backend/surveyapi/utils"
	"context"
	"encoding/json"
	"errors"
	"net/url"
	"os"
	"strconv"

	"gorm.io/gorm"
)

type WilayahRepo interface {
	GetKecamatanById(ctx context.Context, id int64) (*models.Kecamatan, error)
	GetKelurahanById(ctx context.Context, id int64) (*models.Kelurahan, error)
	GetRWById(ctx context.Context, id int64) (*models.DataRwDetail, error)
	GetRTById(ctx context.Context, id int64) (*models.DataRtDetail, error)

	GetDaftarRT(ctx context.Context, rw_id int64, payload payloads.DatatablePayload) (response.RTDatatableResponse, error)
	GetDaftarRW(ctx context.Context, kelurahan_id int64, payload payloads.DatatablePayload) (response.RWDatatableResponse, error)
	GetDaftarKelurahan(ctx context.Context, kecamatan_id int64, payload payloads.DatatablePayload) (response.KelurahanDatatableResponse, error)
	GetDaftarKecamatan(ctx context.Context, payload payloads.DatatablePayload) (response.KecamatanDatatableResponse, error)
}

type wilayahRepo struct {
	dbSlave  *gorm.DB
	dbMaster *gorm.DB
}

func NewWilayahRepo(dbSlave, dbMaster *gorm.DB) WilayahRepo {
	defer utils.GeneralRecover()
	return &wilayahRepo{
		dbSlave,
		dbMaster,
	}
}

func (repository *wilayahRepo) GetKecamatanById(ctx context.Context, id int64) (*models.Kecamatan, error) {
	host := os.Getenv("MASTERAPI_HOST") + ":" + os.Getenv("MASTERAPI_PORT")

	newSlug := map[string]interface{}{"kecamatan_id": strconv.FormatInt(id, 10)}

	dataBytes, err := utils.HitBackend(ctx, host, "GET", "/manajemen-wilayah/kecamatan/detail/:kecamatan_id", newSlug, nil)
	if err != nil {
		return nil, errors.New("Gagal mendapatkan data kecamatan dari MasterAPI")
	}

	var data models.Kecamatan
	if err := json.Unmarshal(dataBytes, &data); err != nil {
		return nil, errors.New("Gagal memparsing data kecamatan dari MasterAPI")
	}
	return &data, nil
}

func (repository *wilayahRepo) GetKelurahanById(ctx context.Context, id int64) (*models.Kelurahan, error) {
	host := os.Getenv("MASTERAPI_HOST") + ":" + os.Getenv("MASTERAPI_PORT")

	newSlug := map[string]interface{}{"kelurahan_id": strconv.FormatInt(id, 10)}

	dataBytes, err := utils.HitBackend(ctx, host, "GET", "/manajemen-wilayah/kelurahan/detail/:kelurahan_id", newSlug, nil)
	if err != nil {
		return nil, errors.New("Gagal mendapatkan data kelurahan dari MasterAPI")
	}

	var data models.Kelurahan
	if err := json.Unmarshal(dataBytes, &data); err != nil {
		return nil, errors.New("Gagal memparsing data kelurahan dari MasterAPI")
	}
	return &data, nil
}

func (repository *wilayahRepo) GetRWById(ctx context.Context, id int64) (*models.DataRwDetail, error) {
	host := os.Getenv("MASTERAPI_HOST") + ":" + os.Getenv("MASTERAPI_PORT")

	newSlug := map[string]interface{}{"rw_id": strconv.FormatInt(id, 10)}

	dataBytes, err := utils.HitBackend(ctx, host, "GET", "/manajemen-wilayah/rw/detail/:rw_id", newSlug, nil)
	if err != nil {
		return nil, errors.New("Gagal mendapatkan data RW dari MasterAPI")
	}

	var data models.DataRwDetail
	if err := json.Unmarshal(dataBytes, &data); err != nil {
		return nil, errors.New("Gagal memparsing data RW dari MasterAPI")
	}
	return &data, nil
}

func (repository *wilayahRepo) GetRTById(ctx context.Context, id int64) (*models.DataRtDetail, error) {
	host := os.Getenv("MASTERAPI_HOST") + ":" + os.Getenv("MASTERAPI_PORT")

	newSlug := map[string]interface{}{"rt_id": strconv.FormatInt(id, 10)}

	dataBytes, err := utils.HitBackend(ctx, host, "GET", "/manajemen-wilayah/rt/detail/:rt_id", newSlug, nil)
	if err != nil {
		return nil, errors.New("Gagal mendapatkan data RT dari MasterAPI")
	}

	var data models.DataRtDetail
	if err := json.Unmarshal(dataBytes, &data); err != nil {
		return nil, errors.New("Gagal memparsing data RT dari MasterAPI")
	}
	return &data, nil
}

func (repository *wilayahRepo) GetDaftarRT(ctx context.Context, rw_id int64, payload payloads.DatatablePayload) (response.RTDatatableResponse, error) {
	host := os.Getenv("MASTERAPI_HOST") + ":" + os.Getenv("MASTERAPI_PORT")

	newSlug := map[string]interface{}{
		"rw_id": rw_id,
	}

	params := url.Values{}
	params.Add("search", payload.Search)
	params.Add("page", strconv.Itoa(payload.Page))
	params.Add("limit", strconv.Itoa(payload.Limit))
	params.Add("order_by", payload.OrderBy)
	params.Add("order_dir", payload.OrderDir)

	dataBytes, err := utils.HitBackendGRPC(ctx, host, "GET", "/manajemen-wilayah/rt/list/:rw_id", params, newSlug, nil)
	if err != nil {
		return response.RTDatatableResponse{}, errors.New("Gagal mendapatkan daftar RT dari MasterAPI")
	}

	var data response.RTDatatableResponse
	if err := json.Unmarshal(dataBytes, &data); err != nil {
		return response.RTDatatableResponse{}, errors.New("Gagal memparsing data daftar RT dari MasterAPI")
	}

	return data, nil
}

func (repository *wilayahRepo) GetDaftarRW(ctx context.Context, kelurahan_id int64, payload payloads.DatatablePayload) (response.RWDatatableResponse, error) {
	host := os.Getenv("MASTERAPI_HOST") + ":" + os.Getenv("MASTERAPI_PORT")

	newSlug := map[string]interface{}{
		"kelurahan_id": kelurahan_id,
	}

	params := url.Values{}
	params.Add("search", payload.Search)
	params.Add("page", strconv.Itoa(payload.Page))
	params.Add("limit", strconv.Itoa(payload.Limit))
	params.Add("order_by", payload.OrderBy)
	params.Add("order_dir", payload.OrderDir)

	dataBytes, err := utils.HitBackendGRPC(ctx, host, "GET", "/manajemen-wilayah/rw/list/:kelurahan_id", params, newSlug, nil)
	if err != nil {
		return response.RWDatatableResponse{}, errors.New("Gagal mendapatkan daftar RW dari MasterAPI")
	}

	var data response.RWDatatableResponse
	if err := json.Unmarshal(dataBytes, &data); err != nil {
		return response.RWDatatableResponse{}, errors.New("Gagal memparsing data daftar RW dari MasterAPI")
	}

	return data, nil
}

func (repository *wilayahRepo) GetDaftarKelurahan(ctx context.Context, kecamatan_id int64, payload payloads.DatatablePayload) (response.KelurahanDatatableResponse, error) {
	host := os.Getenv("MASTERAPI_HOST") + ":" + os.Getenv("MASTERAPI_PORT")

	params := url.Values{}
	params.Add("search", payload.Search)
	params.Add("page", strconv.Itoa(payload.Page))
	params.Add("limit", strconv.Itoa(payload.Limit))
	params.Add("order_by", payload.OrderBy)
	params.Add("order_dir", payload.OrderDir)

	newSlug := map[string]interface{}{
		"kecamatan_id": kecamatan_id,
	}

	dataBytes, err := utils.HitBackendGRPC(ctx, host, "GET", "/manajemen-wilayah/kelurahan/list/:kecamatan_id", params, newSlug, nil)
	if err != nil {
		return response.KelurahanDatatableResponse{}, errors.New("Gagal mendapatkan daftar Kelurahan dari MasterAPI")
	}

	var data response.KelurahanDatatableResponse
	if err := json.Unmarshal(dataBytes, &data); err != nil {
		return response.KelurahanDatatableResponse{}, errors.New("Gagal memparsing data daftar Kelurahan dari MasterAPI")
	}

	return data, nil
}

func (repository *wilayahRepo) GetDaftarKecamatan(ctx context.Context, payload payloads.DatatablePayload) (response.KecamatanDatatableResponse, error) {
	host := os.Getenv("MASTERAPI_HOST") + ":" + os.Getenv("MASTERAPI_PORT")

	params := url.Values{}
	params.Add("search", payload.Search)
	params.Add("page", strconv.Itoa(payload.Page))
	params.Add("limit", strconv.Itoa(payload.Limit))
	params.Add("order_by", payload.OrderBy)
	params.Add("order_dir", payload.OrderDir)

	dataBytes, err := utils.HitBackendGRPC(ctx, host, "GET", "/manajemen-wilayah/kecamatan/list", params, nil, nil)
	if err != nil {
		return response.KecamatanDatatableResponse{}, errors.New("Gagal mendapatkan daftar Kecamatan dari MasterAPI")
	}

	var data response.KecamatanDatatableResponse
	if err := json.Unmarshal(dataBytes, &data); err != nil {
		return response.KecamatanDatatableResponse{}, errors.New("Gagal memparsing data daftar Kecamatan dari MasterAPI")
	}

	return data, nil
}
