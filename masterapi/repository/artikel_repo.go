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

type ArtikelRepository interface {
	Create(ctx context.Context, artikel *models.Artikel) (*models.Artikel, error)
	GetByName(ctx context.Context, name string) (*models.Artikel, error)
	Update(ctx context.Context, artikel *models.Artikel) (*models.Artikel, error)
	GetById(ctx context.Context, id int) (*models.Artikel, error)
	Delete(ctx context.Context, id int) error
	GetOptions(req payloads.ArtikelOptionsPayload, userId int64) ([]response.OptionItem, int64, error)
	GetList(req payloads.DatatablePayload, userId int64) ([]models.ArtikelDatatable, int64, error)
	PublicList(categoryId *int64) ([]models.ArtikelListPublic, error)
	GetByIdPublic(ctx context.Context, id int) (*models.Artikel, error)
}

type artikelRepository struct {
	dbSlave  *gorm.DB
	dbMaster *gorm.DB
}

func NewArtikelRepository(dbSlave, dbMaster *gorm.DB) ArtikelRepository {
	return &artikelRepository{dbSlave: dbSlave, dbMaster: dbMaster}
}

func (repository *artikelRepository) Create(ctx context.Context, artikel *models.Artikel) (*models.Artikel, error) {
	defer utils.GeneralRecover()
	err := repository.dbMaster.WithContext(ctx).Create(artikel).Error
	if err != nil {
		return nil, err
	}
	return artikel, nil
}

func (repository *artikelRepository) GetByName(ctx context.Context, name string) (*models.Artikel, error) {
	defer utils.GeneralRecover()
	var category models.Artikel
	err := repository.dbSlave.WithContext(ctx).Where("LOWER(judul) = ?", strings.ToLower(name)).First(&category).Error
	if err != nil {
		return nil, err
	}
	return &category, nil
}

func (repository *artikelRepository) Update(ctx context.Context, artikel *models.Artikel) (*models.Artikel, error) {
	defer utils.GeneralRecover()
	err := repository.dbMaster.WithContext(ctx).Save(artikel).Error
	if err != nil {
		return nil, err
	}
	return artikel, nil
}

func (repository *artikelRepository) GetById(ctx context.Context, id int) (*models.Artikel, error) {
	defer utils.GeneralRecover()
	var artikel models.Artikel
	err := repository.dbSlave.WithContext(ctx).First(&artikel, id).Error
	if err != nil {
		return nil, err
	}
	return &artikel, nil
}

func (repository *artikelRepository) GetByIdPublic(ctx context.Context, id int) (*models.Artikel, error) {
	defer utils.GeneralRecover()
	var artikel models.Artikel
	err := repository.dbSlave.WithContext(ctx).
		Joins("JOIN artikel_promotes ON artikels.id = artikel_promotes.artikel_id").
		Where("artikel_promotes.status = 'active'").
		First(&artikel, id).Error
	if err != nil {
		return nil, err
	}
	return &artikel, nil
}

func (repository *artikelRepository) Delete(ctx context.Context, id int) error {
	defer utils.GeneralRecover()
	result := repository.dbMaster.WithContext(ctx).Delete(&models.Artikel{}, id)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (repository *artikelRepository) GetOptions(req payloads.ArtikelOptionsPayload, userId int64) ([]response.OptionItem, int64, error) {
	defer utils.GeneralRecover()
	var data []response.OptionItem
	var totalData int64

	db := repository.dbSlave.Table("artikels").
		Select(`
			artikels.id AS id, 
			artikels.judul AS label
		`).Where("artikels.created_by = ?", userId)

	if req.Q != "" {
		searchTerm := "%" + req.Q + "%"
		db = db.Where("artikels.judul ILIKE ?", searchTerm)
	}

	if len(req.IDs) > 0 {
		db = db.Where("artikels.id IN ?", req.IDs)
		err := db.Find(&data).Error
		return data, int64(len(data)), err
	}

	err := db.Count(&totalData).Error
	if err != nil {
		return nil, 0, err
	}

	db = db.Order("artikels.id asc")

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

func (repository *artikelRepository) GetList(req payloads.DatatablePayload, userId int64) ([]models.ArtikelDatatable, int64, error) {
	defer utils.GeneralRecover()
	var data []models.ArtikelDatatable
	var totalData int64

	if req.Page <= 0 {
		req.Page = 1
	}
	if req.Limit <= 0 {
		req.Limit = 5
	}

	baseQuery := func(db *gorm.DB) *gorm.DB {
		return db.Table("artikels").
			Joins("JOIN artikel_categories ON artikels.artikel_category_id = artikel_categories.id").
			Joins("JOIN users ON artikels.created_by = users.id").
			Where("artikel_categories.deleted_at IS NULL AND artikels.created_by = ?", userId)
	}

	db := baseQuery(repository.dbSlave).Select(`
		artikels.id, 
		artikel_categories.name as nama_kategori,
		artikels.artikel_category_id,
		artikels.judul, 
		artikels.created_by,
		users.first_name as created_by_name, 
		artikels.created_at, 
		artikels.updated_at
	`)

	countDB := baseQuery(repository.dbSlave)

	if req.Search != "" {
		searchTerm := "%" + req.Search + "%"
		searchStr := strings.TrimSpace(req.Search)
		parsedDate, isDate := utils.TryParseIndonesianDate(req.Search)

		if isDate {
			condition := `
				artikels.judul ILIKE ? OR 
				artikel_categories.name ILIKE ? OR 
				users.first_name ILIKE ? OR
				DATE(artikels.created_at) = ? OR 
				DATE(artikels.updated_at) = ?
			`
			db = db.Where(condition, searchTerm, searchTerm, searchTerm, parsedDate, parsedDate)
			countDB = countDB.Where(condition, searchTerm, searchTerm, searchTerm, parsedDate, parsedDate)
		} else if len(searchStr) == 4 {
			condition := `
				artikels.judul ILIKE ? OR 
				artikel_categories.name ILIKE ? OR 
				users.first_name ILIKE ? OR
				EXTRACT(YEAR FROM artikels.created_at)::TEXT = ? OR 
				EXTRACT(YEAR FROM artikels.updated_at)::TEXT = ?
			`
			db = db.Where(condition, searchTerm, searchTerm, searchTerm, searchStr, searchStr)
			countDB = countDB.Where(condition, searchTerm, searchTerm, searchTerm, searchStr, searchStr)
		} else {
			condition := `
				artikels.judul ILIKE ? OR 
				artikel_categories.name ILIKE ? OR 
				users.first_name ILIKE ?
			`
			db = db.Where(condition, searchTerm, searchTerm, searchTerm)
			countDB = countDB.Where(condition, searchTerm, searchTerm, searchTerm)
		}
	}

	err := countDB.Distinct("artikels.id").Count(&totalData).Error
	if err != nil {
		return nil, 0, err
	}

	allowedOrderCols := map[string]string{
		"id":              "artikels.id",
		"nama_kategori":   "artikel_categories.name",
		"judul":           "artikels.judul",
		"created_by_name": "users.first_name",
		"created_at":      "artikels.created_at",
		"updated_at":      "artikels.updated_at",
	}

	finalOrderBy := "artikels.id"
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

func (repository *artikelRepository) PublicList(categoryId *int64) ([]models.ArtikelListPublic, error) {
	defer utils.GeneralRecover()
	var data []models.ArtikelListPublic
	db := repository.dbSlave.Table("artikels").
		Select(`
			artikels.id, 
			artikels.judul, 
			artikels.thumbnail,
			artikels.created_at
		`).
		Joins("JOIN artikel_promotes ON artikels.id = artikel_promotes.artikel_id").
		Where("artikel_promotes.status = 'active'")

	if categoryId != nil {
		db = db.Where("artikels.artikel_category_id = ?", *categoryId)
	}

	err := db.Order("artikels.id desc").Find(&data).Error
	if err != nil {
		return nil, err
	}
	return data, nil
}
