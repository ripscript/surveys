package repository

import (
	"backend/surveyapi/utils"

	"gorm.io/gorm"
)

type ManajemenAlurRepo interface {
}

type manajemenAlurRepo struct {
	dbSlave  *gorm.DB
	dbMaster *gorm.DB
}

func NewManajemenAlurRepo(dbSlave, dbMaster *gorm.DB) *manajemenAlurRepo {
	defer utils.GeneralRecover()
	return &manajemenAlurRepo{
		dbSlave,
		dbMaster,
	}
}
