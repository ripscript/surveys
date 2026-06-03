package repository

import (
	"backend/reportapi/models"
	"backend/reportapi/utils"

	"gorm.io/gorm"
)

type HomeRepo interface {
	CountKecamatan() (models.ResponseCounting, error)
	CountKelurahan() (models.ResponseCounting, error)
	CountRw() (models.ResponseCounting, error)
	CountRt() (models.ResponseCounting, error)
	CountSurveyOngoing() (models.ResponseCounting, error)
	CountSurveyUpcoming() (models.ResponseCounting, error)
	CountSurveyFinished() (models.ResponseCounting, error)
}

type homeRepo struct {
	dbSlave  *gorm.DB
	dbMaster *gorm.DB
}

func NewHomeRepo(dbSlave, dbMaster *gorm.DB) *homeRepo {
	defer utils.GeneralRecover()
	return &homeRepo{
		dbSlave,
		dbMaster,
	}
}

func (repo *homeRepo) CountKecamatan() (models.ResponseCounting, error) {
	defer utils.GeneralRecover()
	var kecamatan models.CountKecamatan
	var count int64
	var ResponseCounting models.ResponseCounting
	db := repo.dbSlave

	err := db.Model(kecamatan).Where("deleted_at IS NULL").Count(&count).Error
	if err != nil {
		return ResponseCounting, err
	}

	ResponseCounting.Total = int(count)

	return ResponseCounting, nil
}

func (repo *homeRepo) CountKelurahan() (models.ResponseCounting, error) {
	defer utils.GeneralRecover()
	var kelurahan models.CountKelurahan
	var count int64
	var ResponseCounting models.ResponseCounting
	db := repo.dbSlave

	err := db.Model(kelurahan).Where("deleted_at IS NULL").Count(&count).Error
	if err != nil {
		return ResponseCounting, err
	}

	ResponseCounting.Total = int(count)

	return ResponseCounting, nil
}

func (repo *homeRepo) CountRw() (models.ResponseCounting, error) {
	defer utils.GeneralRecover()
	var rw models.CountRw
	var count int64
	var ResponseCounting models.ResponseCounting
	db := repo.dbSlave

	err := db.Model(rw).Where("deleted_at IS NULL").Count(&count).Error
	if err != nil {
		return ResponseCounting, err
	}

	ResponseCounting.Total = int(count)

	return ResponseCounting, nil
}

func (repo *homeRepo) CountRt() (models.ResponseCounting, error) {
	defer utils.GeneralRecover()
	var rt models.CountRt
	var count int64
	var ResponseCounting models.ResponseCounting
	db := repo.dbSlave

	err := db.Model(rt).Where("deleted_at IS NULL").Count(&count).Error
	if err != nil {
		return ResponseCounting, err
	}

	ResponseCounting.Total = int(count)

	return ResponseCounting, nil
}

func (repo *homeRepo) CountSurveyOngoing() (models.ResponseCounting, error) {
	defer utils.GeneralRecover()
	var rt models.Surveys
	var count int64
	var ResponseCounting models.ResponseCounting
	db := repo.dbSlave

	err := db.Model(rt).Where("status = ?", "ongoing").Count(&count).Error
	if err != nil {
		return ResponseCounting, err
	}

	ResponseCounting.Total = int(count)

	return ResponseCounting, nil
}

func (repo *homeRepo) CountSurveyUpcoming() (models.ResponseCounting, error) {
	defer utils.GeneralRecover()
	var rt models.Surveys
	var count int64
	var ResponseCounting models.ResponseCounting
	db := repo.dbSlave

	err := db.Model(rt).Where("status = ?", "upcoming").Count(&count).Error
	if err != nil {
		return ResponseCounting, err
	}

	ResponseCounting.Total = int(count)

	return ResponseCounting, nil
}

func (repo *homeRepo) CountSurveyFinished() (models.ResponseCounting, error) {
	defer utils.GeneralRecover()
	var rt models.Surveys
	var count int64
	var ResponseCounting models.ResponseCounting
	db := repo.dbSlave

	err := db.Model(rt).Where("status = ?", "finished").Count(&count).Error
	if err != nil {
		return ResponseCounting, err
	}

	ResponseCounting.Total = int(count)

	return ResponseCounting, nil
}
