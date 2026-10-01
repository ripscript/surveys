package repository

import (
	"backend/masterapi/models"
	"backend/masterapi/utils"

	"gorm.io/gorm"
)

type SurveyRepo interface {
	GetSurveyByID(id int) (*models.Survey, error)
}

type surveyRepo struct {
	dbSlave  *gorm.DB
	dbMaster *gorm.DB
}

func NewSurveyRepo(dbSlave, dbMaster *gorm.DB) *surveyRepo {
	defer utils.GeneralRecover()
	return &surveyRepo{
		dbSlave,
		dbMaster,
	}
}

func (repository *surveyRepo) GetSurveyByID(id int) (*models.Survey, error) {
	defer utils.GeneralRecover()
	var survey *models.Survey

	err := repository.dbSlave.Table("surveys").
		Where("id = ?", id).
		First(&survey).Error

	if err != nil {
		return nil, err
	}

	return survey, nil
}
