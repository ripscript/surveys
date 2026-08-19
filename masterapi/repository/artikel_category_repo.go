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
