package repository

import (
	"backend/masterapi/models"
	"backend/masterapi/payloads"
	"backend/masterapi/response"
	"backend/masterapi/utils"
	"context"
	"strings"

	"gorm.io/gorm"
)

type ArtikelCategoryRepository interface {
	Create(ctx context.Context, category *models.ArtikelCategory) (*models.ArtikelCategory, error)
	GetByID(ctx context.Context, id int) (*models.ArtikelCategory, error)
	GetAll(ctx context.Context) ([]models.ArtikelCategory, error)
	Update(ctx context.Context, category *models.ArtikelCategory) (*models.ArtikelCategory, error)
	Delete(ctx context.Context, id int) error

	GetByName(ctx context.Context, name string) (*models.ArtikelCategory, error)
	GetOptions(req payloads.ArtikelCategoryOptionsPayload) ([]response.OptionItem, int64, error)
	GetList(req payloads.DatatablePayload) ([]models.ArtikelCategoryDatatable, int64, error)
	GetPublicList() ([]models.ArtikelCategoryPublic, error)
}

type artikelCategoryRepository struct {
	dbSlave  *gorm.DB
	dbMaster *gorm.DB
}

func NewArtikelCategoryRepository(dbSlave, dbMaster *gorm.DB) ArtikelCategoryRepository {
	return &artikelCategoryRepository{dbSlave: dbSlave, dbMaster: dbMaster}
}

func (repository *artikelCategoryRepository) Create(ctx context.Context, category *models.ArtikelCategory) (*models.ArtikelCategory, error) {
	defer utils.GeneralRecover()
	err := repository.dbMaster.WithContext(ctx).Create(category).Error
	if err != nil {
		return nil, err
	}
	return category, nil
}

func (repository *artikelCategoryRepository) GetByID(ctx context.Context, id int) (*models.ArtikelCategory, error) {
	defer utils.GeneralRecover()
	var category models.ArtikelCategory
	err := repository.dbSlave.WithContext(ctx).First(&category, id).Error
	if err != nil {
		return nil, err
	}
	return &category, nil
}

func (repository *artikelCategoryRepository) GetAll(ctx context.Context) ([]models.ArtikelCategory, error) {
	defer utils.GeneralRecover()
	var categories []models.ArtikelCategory
	err := repository.dbSlave.WithContext(ctx).Find(&categories).Error
	if err != nil {
		return nil, err
	}
	return categories, nil
}

func (repository *artikelCategoryRepository) Update(ctx context.Context, category *models.ArtikelCategory) (*models.ArtikelCategory, error) {
	defer utils.GeneralRecover()

	err := repository.dbMaster.WithContext(ctx).Save(category).Error
	if err != nil {
		return nil, err
	}
	return category, nil
}

func (repository *artikelCategoryRepository) Delete(ctx context.Context, id int) error {
	defer utils.GeneralRecover()
	result := repository.dbMaster.WithContext(ctx).Delete(&models.ArtikelCategory{}, id)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (repository *artikelCategoryRepository) GetByName(ctx context.Context, name string) (*models.ArtikelCategory, error) {
	defer utils.GeneralRecover()
	var category models.ArtikelCategory
	err := repository.dbSlave.WithContext(ctx).Where("LOWER(name) = ?", strings.ToLower(name)).First(&category).Error
	if err != nil {
		return nil, err
	}
	return &category, nil
}

func (repository *artikelCategoryRepository) GetOptions(req payloads.ArtikelCategoryOptionsPayload) ([]response.OptionItem, int64, error) {
	defer utils.GeneralRecover()
	var data []response.OptionItem
	var totalData int64

	db := repository.dbSlave.Table("artikel_categories").
		Select(`
			artikel_categories.id AS id, 
			artikel_categories.name AS label
		`).
		Where("artikel_categories.deleted_at IS NULL")

	if req.Q != "" {
		searchTerm := "%" + req.Q + "%"
		db = db.Where("artikel_categories.name ILIKE ?", searchTerm)
	}

	if len(req.IDs) > 0 {
		db = db.Where("artikel_categories.id IN ?", req.IDs)
		err := db.Find(&data).Error
		return data, int64(len(data)), err
	}

	err := db.Count(&totalData).Error
	if err != nil {
		return nil, 0, err
	}

	db = db.Order("artikel_categories.id asc")

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

func (repository *artikelCategoryRepository) GetList(req payloads.DatatablePayload) ([]models.ArtikelCategoryDatatable, int64, error) {
	defer utils.GeneralRecover()
	var data []models.ArtikelCategoryDatatable
	var totalData int64

	if req.Page <= 0 {
		req.Page = 1
	}
	if req.Limit <= 0 {
		req.Limit = 5
	}

	db := repository.dbSlave.Table("artikel_categories").
		Select(`
			artikel_categories.id, 
			artikel_categories.name, 
			artikel_categories.created_at, 
			artikel_categories.updated_at,
			(SELECT COUNT(id) FROM artikels WHERE artikels.artikel_category_id = artikel_categories.id) AS total_artikel
		`).
		Where("artikel_categories.deleted_at IS NULL")

	countDB := repository.dbSlave.Table("artikel_categories").
		Where("artikel_categories.deleted_at IS NULL")

	if req.Search != "" {
		searchTerm := "%" + req.Search + "%"
		searchStr := strings.TrimSpace(req.Search)
		parsedDate, isDate := utils.TryParseIndonesianDate(req.Search)

		if isDate {
			condition := `
				artikel_categories.name ILIKE ? OR 
				DATE(artikel_categories.created_at) = ? OR 
				DATE(artikel_categories.updated_at) = ?
			`
			db = db.Where(condition, searchTerm, parsedDate, parsedDate)
			countDB = countDB.Where(condition, searchTerm, parsedDate, parsedDate)
		} else if len(searchStr) == 4 {
			condition := `
				artikel_categories.name ILIKE ? OR 
				EXTRACT(YEAR FROM artikel_categories.created_at)::TEXT = ? OR 
				EXTRACT(YEAR FROM artikel_categories.updated_at)::TEXT = ?
			`
			db = db.Where(condition, searchTerm, searchStr, searchStr)
			countDB = countDB.Where(condition, searchTerm, searchStr, searchStr)
		} else {
			condition := `
				artikel_categories.name ILIKE ?
			`
			db = db.Where(condition, searchTerm)
			countDB = countDB.Where(condition, searchTerm)
		}
	}

	err := countDB.Distinct("artikel_categories.id").Count(&totalData).Error
	if err != nil {
		return nil, 0, err
	}

	allowedOrderCols := map[string]string{
		"id":         "artikel_categories.id",
		"name":       "artikel_categories.name",
		"created_at": "artikel_categories.created_at",
		"updated_at": "artikel_categories.updated_at",
	}

	finalOrderBy := "artikel_categories.id"
	finalOrderDir := "desc"

	if mappedCol, isAllowed := allowedOrderCols[req.OrderBy]; isAllowed && req.OrderBy != "" {
		finalOrderBy = mappedCol
	}

	if strings.ToLower(req.OrderDir) == "asc" {
		finalOrderDir = "asc"
	}

	db = db.Order(finalOrderBy + " " + finalOrderDir)

	offset := (req.Page - 1) * req.Limit
	err = db.Limit(req.Limit).Offset(offset).Find(&data).Error
	if err != nil {
		return nil, 0, err
	}

	for i := range data {
		data[i].No = int64(offset + i + 1)
	}

	return data, totalData, nil
}

func (repository *artikelCategoryRepository) GetPublicList() ([]models.ArtikelCategoryPublic, error) {
	defer utils.GeneralRecover()

	var data []models.ArtikelCategoryPublic
	err := repository.dbSlave.Table("artikel_categories").
		Select(`
			artikel_categories.id, 
			artikel_categories.name
		`).
		Where("artikel_categories.deleted_at IS NULL").
		Order("artikel_categories.id asc").
		Find(&data).Error

	if err != nil {
		return nil, err
	}

	return data, nil
}
