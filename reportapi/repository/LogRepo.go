package repository

import (
	"backend/reportapi/models"
	"backend/reportapi/utils"
	"net/url"

	"gorm.io/gorm"
)

type LogRepo interface {
	SaveLogActivities(models.LogActivitySave) error
	GetLogActivities(offset int, limit int, param url.Values) ([]models.LogActivity, int64, error)
}

type logRepo struct {
	dbSlave  *gorm.DB
	dbMaster *gorm.DB
}

func NewLogRepo(dbSlave, dbMaster *gorm.DB) *logRepo {
	defer utils.GeneralRecover()
	return &logRepo{
		dbSlave,
		dbMaster,
	}
}

func (r *logRepo) SaveLogActivities(data models.LogActivitySave) error {
	defer utils.GeneralRecover()

	db := r.dbMaster

	err := db.Create(&data).Error
	if err != nil {
		return err
	}

	return nil
}

func (r *logRepo) GetLogActivities(offset int, limit int, param url.Values) ([]models.LogActivity, int64, error) {
	defer utils.GeneralRecover()

	var data []models.LogActivity
	var total int64
	db := r.dbSlave

	query := db.Model(data)

	if err := query.Model(&models.LogActivity{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if limit != 1 {
		if err := query.Offset(offset).Limit(limit).Find(&data).Error; err != nil {
			return nil, 0, err
		}
	} else {
		if err := query.Offset(offset).Find(&data).Error; err != nil {
			return nil, 0, err
		}
	}

	for i := range data {
		data[i].No = offset + i + 1
	}

	return data, total, nil
}
