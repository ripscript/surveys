package repository

import (
	"backend/userapi/models"
	"backend/userapi/utils"

	"gorm.io/gorm"
)

type RegionRepo interface {
	KecamatanOptions(search string) ([]models.KecamatanOptions, error)
	KelurahanOptions(search string, id int) ([]models.KelurahanOptions, error)
	RwOptions(search string, id int) ([]models.RwOptions, error)
	RtOptions(search string, id int) ([]models.RtOptions, error)
}

type regionRepo struct {
	dbSlave  *gorm.DB
	dbMaster *gorm.DB
}

func NewRegionRepo(dbSlave, dbMaster *gorm.DB) *regionRepo {
	defer utils.GeneralRecover()
	return &regionRepo{
		dbSlave,
		dbMaster,
	}
}

func (repository *regionRepo) KecamatanOptions(search string) ([]models.KecamatanOptions, error) {
	defer utils.GeneralRecover()
	var data []models.KecamatanOptions
	db := repository.dbSlave

	query := db.Model(&data)
	if search != "" {
		query = query.Where("LOWER(sub_district_name) LIKE LOWER(?)", "%"+search+"%")
	}
	err := query.Where("deleted_at IS NULL").Find(&data).Error
	if err != nil {
		return data, err
	}

	return data, nil
}

func (repository *regionRepo) KelurahanOptions(search string, id int) ([]models.KelurahanOptions, error) {
	defer utils.GeneralRecover()
	var data []models.KelurahanOptions
	db := repository.dbSlave

	query := db.Where("sub_district_id = ?", id).Model(&data)
	if search != "" {
		query = query.Where("LOWER(village_name) LIKE LOWER(?)", "%"+search+"%")
	}
	err := query.Where("deleted_at IS NULL").Find(&data).Error
	if err != nil {
		return data, err
	}

	return data, nil
}

func (repository *regionRepo) RwOptions(search string, id int) ([]models.RwOptions, error) {
	defer utils.GeneralRecover()
	var data []models.RwOptions
	db := repository.dbSlave

	query := db.Where("kelurahan_id = ?", id).Model(&data)
	if search != "" {
		query = query.Where("LOWER(nama_rw) LIKE LOWER(?)", "%"+search+"%")
	}
	err := query.Where("deleted_at IS NULL").Find(&data).Error
	if err != nil {
		return data, err
	}

	return data, nil
}

func (repository *regionRepo) RtOptions(search string, id int) ([]models.RtOptions, error) {
	defer utils.GeneralRecover()
	var data []models.RtOptions
	db := repository.dbSlave

	query := db.Where("rw_id = ?", id).Model(&data)
	if search != "" {
		query = query.Where("LOWER(nama_rt) LIKE LOWER(?)", "%"+search+"%")
	}
	err := query.Where("deleted_at IS NULL").Find(&data).Error
	if err != nil {
		return data, err
	}

	return data, nil
}
