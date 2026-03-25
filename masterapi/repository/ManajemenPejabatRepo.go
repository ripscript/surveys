package repository

import (
	"backend/masterapi/models"
	"backend/masterapi/utils"

	"gorm.io/gorm"
)

type ManajemenPejabatRepo interface {
	GetFirstPejabatanWilayahByWilayahIdDanTipeWilayah(id int, tipeWilayah int) (*models.PejabatWilayah, error)
}

type manajemenPejabatRepo struct {
	dbSlave  *gorm.DB
	dbMaster *gorm.DB
}

func NewManajemenPejabatRepo(dbSlave, dbMaster *gorm.DB) *manajemenPejabatRepo {
	defer utils.GeneralRecover()
	return &manajemenPejabatRepo{
		dbSlave,
		dbMaster,
	}
}

func (repository *manajemenPejabatRepo) GetFirstPejabatanWilayahByWilayahIdDanTipeWilayah(id int, tipeWilayah int) (*models.PejabatWilayah, error) {
	defer utils.GeneralRecover()
	var data models.PejabatWilayah
	db := repository.dbSlave

	err := db.Preload("Respondent").Where("tipe_wilayah = ? AND id_wilayah = ?", tipeWilayah, id).Order("id DESC").First(&data).Error
	if err != nil {
		return nil, err
	}

	return &data, nil
}
