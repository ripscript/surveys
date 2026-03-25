package repository

import (
	"backend/masterapi/models"
	"backend/masterapi/payloads"
	"backend/masterapi/utils"
	"strings"

	"gorm.io/gorm"
)

type ManajemenArtikelRepo interface {
	CreateKategoriArtikel(artikelKategori models.ArtikelKategori) (*models.ArtikelKategori, error)
	UpdateKategoriArtikel(artikelKategori models.ArtikelKategori) (*models.ArtikelKategori, error)
	GetKategoriArtikelById(id int64) (*models.ArtikelKategori, error)
	GetKategoriArtikelByName(name string) (*models.ArtikelKategori, error)
	DeleteKategoriArtikelById(id int64) error
	GetListKategoriArtikel(req payloads.DatatablePayload) ([]models.ArtikelKategori, int64, error)
}

type manajemenArtikelRepo struct {
	dbSlave  *gorm.DB
	dbMaster *gorm.DB
}

func NewManajemenArtikelRepo(dbSlave, dbMaster *gorm.DB) *manajemenArtikelRepo {
	defer utils.GeneralRecover()
	return &manajemenArtikelRepo{
		dbSlave,
		dbMaster,
	}
}

func (repository *manajemenArtikelRepo) CreateKategoriArtikel(artikelKategori models.ArtikelKategori) (*models.ArtikelKategori, error) {
	defer utils.GeneralRecover()

	db := repository.dbMaster

	err := db.Create(&artikelKategori).Error
	if err != nil {
		return nil, err
	}

	return &artikelKategori, nil
}

func (repository *manajemenArtikelRepo) GetKategoriArtikelById(id int64) (*models.ArtikelKategori, error) {
	defer utils.GeneralRecover()

	db := repository.dbSlave

	var artikelKategori models.ArtikelKategori
	err := db.Where("id = ?", id).First(&artikelKategori).Error
	if err != nil {
		return nil, err
	}

	return &artikelKategori, nil
}

func (repository *manajemenArtikelRepo) GetKategoriArtikelByName(name string) (*models.ArtikelKategori, error) {
	defer utils.GeneralRecover()

	db := repository.dbSlave

	var artikelKategori models.ArtikelKategori
	err := db.Where("name = ?", name).First(&artikelKategori).Error
	if err != nil {
		return nil, err
	}

	return &artikelKategori, nil
}

func (repository *manajemenArtikelRepo) UpdateKategoriArtikel(artikelKategori models.ArtikelKategori) (*models.ArtikelKategori, error) {
	defer utils.GeneralRecover()

	db := repository.dbMaster

	err := db.Save(&artikelKategori).Error
	if err != nil {
		return nil, err
	}

	return &artikelKategori, nil
}

func (repository *manajemenArtikelRepo) DeleteKategoriArtikelById(id int64) error {
	defer utils.GeneralRecover()

	db := repository.dbMaster

	err := db.Delete(&models.ArtikelKategori{}, "id = ?", id).Error
	if err != nil {
		return err
	}

	return nil
}

func (repository *manajemenArtikelRepo) GetListKategoriArtikel(req payloads.DatatablePayload) ([]models.ArtikelKategori, int64, error) {
	defer utils.GeneralRecover()
	var data []models.ArtikelKategori
	var totalData int64

	db := repository.dbSlave.Table("artikel_categories").
		Select(`
			artikel_categories.id,
			artikel_categories.name,
			artikel_categories.created_at,
			artikel_categories.updated_at,
			artikel_categories.deleted_at
		`)

	if req.Search != "" {
		searchTerm := "%" + req.Search + "%"
		searchStr := strings.TrimSpace(req.Search)
		parsedDate, isDate := utils.TryParseIndonesianDate(req.Search)

		if isDate {
			db = db.Where(`
				artikel_categories.name ILIKE ? OR 
				DATE(artikel_categories.created_at) = ? OR 
				DATE(artikel_categories.updated_at) = ?
			`, searchTerm, parsedDate, parsedDate)
		} else if len(searchStr) == 4 {
			db = db.Where(`
				artikel_categories.name ILIKE ? OR  
				EXTRACT(YEAR FROM artikel_categories.created_at)::TEXT = ? OR 
				EXTRACT(YEAR FROM artikel_categories.updated_at)::TEXT = ?
			`, searchTerm, searchStr, searchStr)
		} else {
			db = db.Where(`artikel_categories.name ILIKE ?`, searchTerm)
		}
	}

	err := db.Count(&totalData).Error
	if err != nil {
		return nil, 0, err
	}

	if req.OrderBy != "" {
		finalOrderBy := "artikel_categories.id"
		finalOrderDir := "desc"

		allowedOrderCols := map[string]string{
			"id":         "artikel_categories.id",
			"name":       "artikel_categories.name",
			"created_at": "artikel_categories.created_at",
			"updated_at": "artikel_categories.updated_at",
		}

		if mappedCol, isAllowed := allowedOrderCols[req.OrderBy]; isAllowed && req.OrderBy != "" {
			finalOrderBy = mappedCol
		}

		if strings.ToLower(req.OrderDir) == "asc" {
			finalOrderDir = "asc"
		}

		db = db.Order(finalOrderBy + " " + finalOrderDir)
	} else {
		db = db.Order("artikel_categories.id desc")
	}

	// Fitur Pagination
	offset := (req.Page - 1) * req.Limit
	err = db.Limit(req.Limit).Offset(offset).Find(&data).Error
	if err != nil {
		return nil, 0, err
	}

	return data, totalData, nil
}
