package repository

import (
	"backend/masterapi/models"
	"backend/masterapi/payloads"
	"backend/masterapi/utils"
	"context"
	"strings"

	"gorm.io/gorm"
)

type ArtikelPromoteRepository interface {
	Create(ctx context.Context, promote *models.ArtikelPromote) (*models.ArtikelPromote, error)
	GetByID(ctx context.Context, id int) (*models.ArtikelPromote, error)
	GetByArtikelID(ctx context.Context, artikelID int64) (*models.ArtikelPromote, error)
	GetAll(ctx context.Context, req payloads.ArtikelPromoteOptionsPayload) ([]models.ArtikelPromote, int64, error)
	Update(ctx context.Context, promote *models.ArtikelPromote) (*models.ArtikelPromote, error)
	Delete(ctx context.Context, id int) error
	GetList(req payloads.DatatablePayload) ([]models.ArtikelPromoteDatatable, int64, error)
}

type artikelPromoteRepository struct {
	dbSlave  *gorm.DB
	dbMaster *gorm.DB
}

func NewArtikelPromoteRepository(dbSlave, dbMaster *gorm.DB) ArtikelPromoteRepository {
	return &artikelPromoteRepository{dbSlave: dbSlave, dbMaster: dbMaster}
}

func (repository *artikelPromoteRepository) Create(ctx context.Context, promote *models.ArtikelPromote) (*models.ArtikelPromote, error) {
	defer utils.GeneralRecover()
	err := repository.dbMaster.WithContext(ctx).Create(promote).Error
	if err != nil {
		return nil, err
	}
	return promote, nil
}

func (repository *artikelPromoteRepository) GetByID(ctx context.Context, id int) (*models.ArtikelPromote, error) {
	defer utils.GeneralRecover()
	var promote models.ArtikelPromote
	err := repository.dbSlave.WithContext(ctx).First(&promote, id).Error
	if err != nil {
		return nil, err
	}
	return &promote, nil
}

func (repository *artikelPromoteRepository) GetByArtikelID(ctx context.Context, artikelID int64) (*models.ArtikelPromote, error) {
	defer utils.GeneralRecover()
	var promote models.ArtikelPromote
	err := repository.dbSlave.WithContext(ctx).Where("artikel_id = ?", artikelID).First(&promote).Error
	if err != nil {
		return nil, err
	}
	return &promote, nil
}

func (repository *artikelPromoteRepository) GetAll(ctx context.Context, req payloads.ArtikelPromoteOptionsPayload) ([]models.ArtikelPromote, int64, error) {
	defer utils.GeneralRecover()
	var data []models.ArtikelPromote
	var totalData int64

	db := repository.dbSlave.WithContext(ctx).Model(&models.ArtikelPromote{})

	if req.Status != "" {
		db = db.Where("status = ?", req.Status)
	}

	if len(req.IDs) > 0 {
		db = db.Where("id IN ?", req.IDs)
		err := db.Find(&data).Error
		return data, int64(len(data)), err
	}

	err := db.Count(&totalData).Error
	if err != nil {
		return nil, 0, err
	}

	db = db.Order("id desc")

	limit := req.Limit
	if limit <= 0 {
		limit = 20
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

func (repository *artikelPromoteRepository) Update(ctx context.Context, promote *models.ArtikelPromote) (*models.ArtikelPromote, error) {
	defer utils.GeneralRecover()
	err := repository.dbMaster.WithContext(ctx).Save(promote).Error
	if err != nil {
		return nil, err
	}
	return promote, nil
}

func (repository *artikelPromoteRepository) Delete(ctx context.Context, id int) error {
	defer utils.GeneralRecover()
	result := repository.dbMaster.WithContext(ctx).Delete(&models.ArtikelPromote{}, id)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (repository *artikelPromoteRepository) GetList(req payloads.DatatablePayload) ([]models.ArtikelPromoteDatatable, int64, error) {
	defer utils.GeneralRecover()
	var data []models.ArtikelPromoteDatatable
	var totalData int64

	if req.Page <= 0 {
		req.Page = 1
	}
	if req.Limit <= 0 {
		req.Limit = 5
	}

	baseQuery := func(db *gorm.DB) *gorm.DB {
		return db.Table("artikel_promotes").
			Joins("JOIN artikels ON artikel_promotes.artikel_id = artikels.id").
			Joins("JOIN artikel_categories ON artikels.artikel_category_id = artikel_categories.id").
			Where("artikel_categories.deleted_at IS NULL")
	}

	db := baseQuery(repository.dbSlave).Select(`
		artikel_promotes.id,
		artikels.judul as nama_artikel,
		artikel_categories.name as nama_kategori, 
		artikel_promotes.status,
		artikel_promotes.created_at, 
		artikel_promotes.updated_at
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
				DATE(artikel_promotes.created_at) = ? OR 
				DATE(artikel_promotes.updated_at) = ?
			`
			db = db.Where(condition, searchTerm, searchTerm, parsedDate, parsedDate)
			countDB = countDB.Where(condition, searchTerm, searchTerm, parsedDate, parsedDate)
		} else if len(searchStr) == 4 {
			condition := `
				artikels.judul ILIKE ? OR 
				artikel_categories.name ILIKE ? OR 
				EXTRACT(YEAR FROM artikel_promotes.created_at)::TEXT = ? OR 
				EXTRACT(YEAR FROM artikel_promotes.updated_at)::TEXT = ?
			`
			db = db.Where(condition, searchTerm, searchTerm, searchStr, searchStr)
			countDB = countDB.Where(condition, searchTerm, searchTerm, searchStr, searchStr)
		} else {
			condition := `
				artikels.judul ILIKE ? OR 
				artikel_categories.name ILIKE ?
			`
			db = db.Where(condition, searchTerm, searchTerm)
			countDB = countDB.Where(condition, searchTerm, searchTerm)
		}
	}

	err := countDB.Distinct("artikel_promotes.id").Count(&totalData).Error
	if err != nil {
		return nil, 0, err
	}

	allowedOrderCols := map[string]string{
		"id":            "artikel_promotes.id",
		"nama_kategori": "artikel_categories.name",
		"nama_artikel":  "artikels.judul",
		"status":        "artikel_promotes.status",
		"created_at":    "artikel_promotes.created_at",
		"updated_at":    "artikel_promotes.updated_at",
	}

	finalOrderBy := "artikel_promotes.id"
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
