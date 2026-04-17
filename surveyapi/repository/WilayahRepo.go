package repository

import (
	"backend/surveyapi/models"
	"backend/surveyapi/utils"
	"context"
	"encoding/json"
	"errors"
	"os"
	"strconv"

	"github.com/davecgh/go-spew/spew"
	"gorm.io/gorm"
)

type WilayahRepo interface {
	GetKecamatanById(ctx context.Context, id int64) (*models.Kecamatan, error)
	GetKelurahanById(ctx context.Context, id int64) (*models.Kelurahan, error)
	GetRWById(ctx context.Context, id int64) (*models.DataRw, error)
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
		spew.Dump(err)
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
		spew.Dump(err)
		return nil, errors.New("Gagal mendapatkan data kelurahan dari MasterAPI")
	}

	var data models.Kelurahan
	if err := json.Unmarshal(dataBytes, &data); err != nil {
		return nil, errors.New("Gagal memparsing data kelurahan dari MasterAPI")
	}
	return &data, nil
}

func (repository *wilayahRepo) GetRWById(ctx context.Context, id int64) (*models.DataRw, error) {
	host := os.Getenv("MASTERAPI_HOST") + ":" + os.Getenv("MASTERAPI_PORT")

	newSlug := map[string]interface{}{"rw_id": strconv.FormatInt(id, 10)}

	dataBytes, err := utils.HitBackend(ctx, host, "GET", "/manajemen-wilayah/rw/detail/:rw_id", newSlug, nil)
	if err != nil {
		spew.Dump(err)
		return nil, errors.New("Gagal mendapatkan data RW dari MasterAPI")
	}

	var data models.DataRw
	if err := json.Unmarshal(dataBytes, &data); err != nil {
		return nil, errors.New("Gagal memparsing data RW dari MasterAPI")
	}
	return &data, nil
}
