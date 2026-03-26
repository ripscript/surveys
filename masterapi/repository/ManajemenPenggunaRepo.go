package repository

import (
	"backend/masterapi/models"
	"backend/masterapi/payloads"
	"backend/masterapi/response"
	"backend/masterapi/utils"

	"gorm.io/gorm"
)

type ManajemenPenggunaRepo interface {
	GetRoleOptions(req payloads.RtOptionsPayload) ([]response.OptionItem, int64, error)
	GetRoleByID(id int64) (*models.Role, error)
	CreateResponden(rw models.Respondent) (*models.Respondent, error)
}

type manajemenPenggunaRepo struct {
	dbSlave  *gorm.DB
	dbMaster *gorm.DB
}

func NewManajemenPenggunaRepo(dbSlave, dbMaster *gorm.DB) *manajemenPenggunaRepo {
	defer utils.GeneralRecover()
	return &manajemenPenggunaRepo{
		dbSlave,
		dbMaster,
	}
}

func (repository *manajemenPenggunaRepo) GetRoleOptions(req payloads.RtOptionsPayload) ([]response.OptionItem, int64, error) {
	defer utils.GeneralRecover()
	var data []response.OptionItem
	var totalData int64

	db := repository.dbSlave.Table("role").
		Select(`
            role.id AS id, 
            role.name AS label
        `)

	if len(req.IDs) > 0 {
		db = db.Where("role.id IN ?", req.IDs)
		err := db.Find(&data).Error
		return data, int64(len(data)), err
	}

	if req.Q != "" {
		searchTerm := "%" + req.Q + "%"

		db = db.Where("role.name ILIKE ?", searchTerm)
	}

	err := db.Count(&totalData).Error
	if err != nil {
		return nil, 0, err
	}

	db = db.Order("role.name ASC")

	limit := req.Limit
	if limit <= 0 {
		limit = 1000
	}

	page := req.Page
	if page <= 0 {
		page = 1
	}

	offset := (page - 1) * limit
	err = db.Limit(limit).Offset(offset).Find(&data).Error
	if err != nil {
		return nil, 0, err
	}

	return data, totalData, nil
}

func (repository *manajemenPenggunaRepo) GetRoleByID(id int64) (*models.Role, error) {
	defer utils.GeneralRecover()
	var data models.Role
	db := repository.dbSlave

	err := db.Select("role.*").
		Where("role.id = ?", id).
		First(&data).Error

	if err != nil {
		return nil, err
	}

	return &data, nil
}

func (repository *manajemenPenggunaRepo) CreateResponden(rw models.Respondent) (*models.Respondent, error) {
	defer utils.GeneralRecover()

	db := repository.dbMaster

	err := db.Create(&rw).Error
	if err != nil {
		return nil, err
	}

	return &rw, nil
}
