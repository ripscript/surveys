package repository

import (
	"backend/surveyapi/models"
	"backend/surveyapi/utils"

	"gorm.io/gorm"
)

type TemplateUcapanRepo interface {
	GetTemplateUcapanById(id int) (*models.GeneralTemplate, error)
}

type templateUcapanRepo struct {
	dbSlave  *gorm.DB
	dbMaster *gorm.DB
}

func NewTemplateUcapanRepo(dbSlave, dbMaster *gorm.DB) *templateUcapanRepo {
	defer utils.GeneralRecover()
	return &templateUcapanRepo{
		dbSlave,
		dbMaster,
	}
}

func (repository *templateUcapanRepo) GetTemplateUcapanById(id int) (*models.GeneralTemplate, error) {
	defer utils.GeneralRecover()
	var data models.GeneralTemplate
	db := repository.dbSlave

	err := db.Select("general_templates.*").
		Where("general_templates.id = ?", id).
		First(&data).Error

	if err != nil {
		return nil, err
	}

	return &data, nil
}
