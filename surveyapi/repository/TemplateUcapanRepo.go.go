package repository

import (
	"backend/surveyapi/models"
	"backend/surveyapi/payloads"
	"backend/surveyapi/response"
	"backend/surveyapi/utils"
	"strings"

	"gorm.io/gorm"
)

type TemplateUcapanRepo interface {
	GetTemplateUcapanById(id int) (*models.GeneralTemplateDetail, error)
	GetVariableTemplateUcapanOptions(req payloads.TemplateUcapanOptionsPayload) ([]response.OptionItem, int64, error)
	GetTemplateUcapanByName(name string) (*models.GeneralTemplate, error)
	CreateTemplateUcapan(templateUcapan models.GeneralTemplate) (*models.GeneralTemplate, error)
	UpdateTemplateUcapan(templateUcapan models.GeneralTemplate) (*models.GeneralTemplate, error)
	DeleteTemplateUcapan(id int) error
	GetListTemplateUcapan(req payloads.DatatablePayload) ([]models.GeneralTemplateDatatableResponse, int64, error)
	GetTemplateUcapanOption(req payloads.UcapanOptionsPayload) ([]response.OptionItem, int64, error)
	IsUsedTemplateUcapan(templateUcapanId int64) (bool, error)
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

func (repository *templateUcapanRepo) GetTemplateUcapanById(id int) (*models.GeneralTemplateDetail, error) {
	defer utils.GeneralRecover()
	var data models.GeneralTemplateDetail
	db := repository.dbSlave

	err := db.Select("general_templates.*").
		Where("general_templates.id = ?", id).
		First(&data).Error

	if err != nil {
		return nil, err
	}

	return &data, nil
}

func (repository *templateUcapanRepo) GetVariableTemplateUcapanOptions(req payloads.TemplateUcapanOptionsPayload) ([]response.OptionItem, int64, error) {
	defer utils.GeneralRecover()
	var data []response.OptionItem
	var totalData int64

	db := repository.dbSlave.Table("data_general_templates").
		Select(`
            data_general_templates.id AS id, 
            data_general_templates.variable_name AS label
        `)

	if len(req.IDs) > 0 {
		db = db.Where("data_general_templates.id IN ?", req.IDs)
		err := db.Find(&data).Error
		return data, int64(len(data)), err
	}

	if req.Q != "" {
		searchTerm := "%" + req.Q + "%"

		db = db.Where("data_general_templates.variable_name ILIKE ?", searchTerm)
	}

	err := db.Count(&totalData).Error
	if err != nil {
		return nil, 0, err
	}

	db = db.Order("data_general_templates.id asc")

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

func (repository *templateUcapanRepo) GetTemplateUcapanByName(name string) (*models.GeneralTemplate, error) {
	defer utils.GeneralRecover()
	var data models.GeneralTemplate
	db := repository.dbSlave

	err := db.Select("general_templates.*").
		Where("general_templates.name = ?", name).
		First(&data).Error

	if err != nil {
		return nil, err
	}

	return &data, nil
}

func (repository *templateUcapanRepo) CreateTemplateUcapan(templateUcapan models.GeneralTemplate) (*models.GeneralTemplate, error) {
	defer utils.GeneralRecover()
	err := repository.dbMaster.Create(&templateUcapan).Error
	if err != nil {
		return nil, err
	}
	return &templateUcapan, nil
}

func (repository *templateUcapanRepo) UpdateTemplateUcapan(templateUcapan models.GeneralTemplate) (*models.GeneralTemplate, error) {
	defer utils.GeneralRecover()
	err := repository.dbMaster.Save(&templateUcapan).Error
	if err != nil {
		return nil, err
	}
	return &templateUcapan, nil
}

func (repository *templateUcapanRepo) DeleteTemplateUcapan(id int) error {
	defer utils.GeneralRecover()
	err := repository.dbMaster.Delete(&models.GeneralTemplate{}, id).Error
	return err
}

func (repository *templateUcapanRepo) GetListTemplateUcapan(req payloads.DatatablePayload) ([]models.GeneralTemplateDatatableResponse, int64, error) {
	defer utils.GeneralRecover()
	var data []models.GeneralTemplateDatatableResponse
	var totalData int64

	db := repository.dbSlave.Table("general_templates").
		Select(`
			general_templates.id AS id,
			general_templates.type AS type,
			general_templates.name AS name,
			general_templates.created_at AS created_at,
			general_templates.updated_at AS updated_at,
			general_templates.deleted_at AS deleted_at,
			users.first_name AS created_by_name
		`).
		Joins("LEFT JOIN users ON users.id = general_templates.created_by")

	if req.Search != "" {
		searchTerm := "%" + req.Search + "%"
		searchStr := strings.TrimSpace(req.Search)
		parsedDate, isDate := utils.TryParseIndonesianDate(req.Search)

		if isDate {
			db = db.Where(`
				general_templates.name ILIKE ? OR 
				general_templates.type ILIKE ? OR 
				users.first_name ILIKE ? OR
				DATE(general_templates.created_at) = ? OR
				DATE(general_templates.updated_at) = ?
			`, searchTerm, searchTerm, searchTerm, parsedDate, parsedDate)
		} else if len(searchStr) == 4 {
			db = db.Where(`
				general_templates.name ILIKE ? OR 
				general_templates.type ILIKE ? OR 
				users.first_name ILIKE ? OR
				EXTRACT(YEAR FROM general_templates.created_at)::TEXT = ? OR
				EXTRACT(YEAR FROM general_templates.updated_at)::TEXT = ?
			`, searchTerm, searchTerm, searchTerm, searchStr, searchStr)
		} else {
			db = db.Where(`
			general_templates.name ILIKE ? OR 
			general_templates.type ILIKE ? OR
			users.first_name ILIKE ?
			`, searchTerm, searchTerm, searchTerm)
		}
	}

	err := db.Count(&totalData).Error
	if err != nil {
		return nil, 0, err
	}

	if req.OrderBy != "" {
		finalOrderBy := "general_templates.id"
		finalOrderDir := "desc"

		allowedOrderCols := map[string]string{
			"name":            "general_templates.name",
			"type":            "general_templates.type",
			"created_at":      "general_templates.created_at",
			"updated_at":      "general_templates.updated_at",
			"created_by_name": "users.first_name",
		}

		if mappedCol, isAllowed := allowedOrderCols[req.OrderBy]; isAllowed && req.OrderBy != "" {
			finalOrderBy = mappedCol
		}

		if strings.ToLower(req.OrderDir) == "asc" {
			finalOrderDir = "asc"
		}

		db = db.Order(finalOrderBy + " " + finalOrderDir)
	} else {
		db = db.Order("general_templates.id desc")
	}

	// Fitur Pagination
	offset := (req.Page - 1) * req.Limit
	err = db.Limit(req.Limit).Offset(offset).Find(&data).Error
	if err != nil {
		return nil, 0, err
	}

	return data, totalData, nil
}

func (repository *templateUcapanRepo) GetTemplateUcapanOption(req payloads.UcapanOptionsPayload) ([]response.OptionItem, int64, error) {
	defer utils.GeneralRecover()
	var data []response.OptionItem
	var totalData int64

	db := repository.dbSlave.Table("general_templates").
		Select(`
			general_templates.id AS id, 
			general_templates.name AS label
		`).
		Where("general_templates.deleted_at IS NULL")

	if len(req.IDs) > 0 {
		db = db.Where("general_templates.id IN ?", req.IDs)
		err := db.Find(&data).Error
		return data, int64(len(data)), err
	}

	if req.Q != "" {
		searchTerm := "%" + req.Q + "%"
		db = db.Where("general_templates.name ILIKE ?", searchTerm)
	}

	if req.Type != nil && *req.Type != "" {
		db = db.Where("general_templates.type = ?", *req.Type)
	}

	err := db.Count(&totalData).Error
	if err != nil {
		return nil, 0, err
	}

	db = db.Order("general_templates.id asc")

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

func (repository *templateUcapanRepo) IsUsedTemplateUcapan(templateUcapanId int64) (bool, error) {
	defer utils.GeneralRecover()

	var count int64
	db := repository.dbSlave

	err := db.Model(&models.FlowDetail{}).Where("closing_id = ?", templateUcapanId).Count(&count).Error
	if err != nil {
		return false, err
	}

	return count > 0, nil
}
