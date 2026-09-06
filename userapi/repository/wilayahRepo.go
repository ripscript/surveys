package repository

import (
	"backend/userapi/models"
	"backend/userapi/utils"

	"gorm.io/gorm"
)

type WilayahRepo interface {
	GetGeoNameByKecamatanId(kecamatanId int64) (string, error)
	GetGeoNameByKelurahanId(kecamatanId int64) (string, error)
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

func (r *wilayahRepo) GetGeoNameByKecamatanId(kecamatanId int64) (string, error) {
	var kecamatan models.KecamatanModel
	err := r.dbSlave.Where("id = ?", kecamatanId).First(&kecamatan).Error
	if err != nil {
		return "", err
	}

	if kecamatan.GeoName != nil {
		return *kecamatan.GeoName, nil
	}

	return "", nil
}

func (r *wilayahRepo) GetGeoNameByKelurahanId(kelurahanId int64) (string, error) {
	var kelurahan models.KelurahanModel
	err := r.dbSlave.Where("id = ?", kelurahanId).First(&kelurahan).Error
	if err != nil {
		return "", err
	}

	if kelurahan.GeoName != nil {
		return *kelurahan.GeoName, nil
	}

	return "", nil

}
