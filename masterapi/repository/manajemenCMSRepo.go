package repository

import (
	"backend/masterapi/models"
	"backend/masterapi/payloads"
	"backend/masterapi/utils"
	"errors"
	"strconv"
	"strings"

	"gorm.io/gorm"
)

type ManajemenCMSRepo interface {
	Create(section *models.ManajemenCMS) error
	CreateSection(section *models.CMSSection) (*models.CMSSection, error)
	CreateCMSContent(content *models.CMSContent) (*models.CMSContent, error)
	GetLastSection() (int, error)
	IsNameSectionExist(nameSection string) (bool, error)
	IsSlugSectionExist(slugSection string) (bool, error)
	GetListSection(req payloads.DatatablePayload) ([]models.CMSSectionDatatable, int64, error)
	GetSectionById(id int) (*models.CMSSection, error)
	UpdateSection(section *models.CMSSection) (*models.CMSSection, error)
	DeleteSection(section *models.CMSSection) error
	GetSectionBySlug(slug string) (*models.CMSSection, error)
	UpsertContents(sectionID int, contents []models.CMSContent) error
	UpsertItems(sectionID int, items []models.CMSItem) error
	UpsertMedia(sectionID int, media []models.CMSMedia) error
	GetLastSectionOrder() (int, error)
	GetAdjacentSectionForReorder(currentOrder int, direction string) (*models.CMSSection, error)
	SwapSectionOrder(sectionA, sectionB *models.CMSSection) error
	SyncMedia(sectionID int, media []models.CMSMedia) (removedImagePaths []string, err error)
}

type manajemenCMSRepo struct {
	dbSlave  *gorm.DB
	dbMaster *gorm.DB
}

func NewManajemenCMSRepo(dbSlave, dbMaster *gorm.DB) ManajemenCMSRepo {
	defer utils.GeneralRecover()
	return &manajemenCMSRepo{
		dbSlave,
		dbMaster,
	}
}

func (repo *manajemenCMSRepo) Create(section *models.ManajemenCMS) error {
	return repo.dbMaster.Create(section).Error
}

func (repo *manajemenCMSRepo) CreateSection(section *models.CMSSection) (*models.CMSSection, error) {
	err := repo.dbMaster.Create(section).Error
	if err != nil {
		return nil, err
	}
	return section, nil
}

func (repo *manajemenCMSRepo) CreateCMSContent(content *models.CMSContent) (*models.CMSContent, error) {
	err := repo.dbMaster.Create(content).Error
	if err != nil {
		return nil, err
	}
	return content, nil
}

func (repo *manajemenCMSRepo) GetLastSection() (int, error) {
	var lastSection string

	err := repo.dbSlave.
		Model(&models.ManajemenCMS{}).
		Select("section").
		Order("CAST(section AS INTEGER) DESC").
		Limit(1).
		Scan(&lastSection).Error

	if err != nil {
		return 0, err
	}

	if lastSection == "" {
		return 0, nil
	}

	num, err := strconv.Atoi(lastSection)
	if err != nil {
		return 0, err
	}

	return num, nil
}

func (repo *manajemenCMSRepo) IsNameSectionExist(nameSection string) (bool, error) {
	var count int64

	err := repo.dbSlave.
		Model(&models.CMSSection{}).
		Where("LOWER(name) = ?", strings.ToLower(nameSection)).
		Count(&count).Error

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func (repo *manajemenCMSRepo) IsSlugSectionExist(slugSection string) (bool, error) {
	var count int64

	err := repo.dbSlave.
		Model(&models.CMSSection{}).
		Where("slug = ?", slugSection).
		Count(&count).Error

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func (repository *manajemenCMSRepo) GetListSection(req payloads.DatatablePayload) ([]models.CMSSectionDatatable, int64, error) {
	defer utils.GeneralRecover()
	var data []models.CMSSectionDatatable
	var totalData int64

	if req.Page <= 0 {
		req.Page = 1
	}
	if req.Limit <= 0 {
		req.Limit = 5
	}

	db := repository.dbSlave.Table("cms_sections").
		Select(`
			cms_sections.id,
			cms_sections.slug,
			cms_sections.name,
			cms_sections.type,
			cms_sections.section_order,
			cms_sections.is_repeatable,
			cms_sections.is_system,
			cms_sections.status,
			cms_sections.created_at,
			cms_sections.updated_at
		`)

	countDB := repository.dbSlave.Table("cms_sections").
		Select(`
			cms_sections.id,
			cms_sections.slug,
			cms_sections.name,
			cms_sections.type,
			cms_sections.section_order,
			cms_sections.is_repeatable,
			cms_sections.is_system,
			cms_sections.status,
			cms_sections.created_at,
			cms_sections.updated_at
		`)

	if req.Search != "" {
		searchTerm := "%" + req.Search + "%"
		searchStr := strings.TrimSpace(req.Search)
		parsedDate, isDate := utils.TryParseIndonesianDate(req.Search)

		if isDate {
			condition := `
				cms_sections.name ILIKE ? OR 
				cms_sections.type ILIKE ? OR 
				DATE(cms_sections.created_at) = ? OR 
				DATE(cms_sections.updated_at) = ?
			`
			db = db.Where(condition, searchTerm, searchTerm, parsedDate, parsedDate)
			countDB = countDB.Where(condition, searchTerm, searchTerm, parsedDate, parsedDate)
		} else if len(searchStr) == 4 {
			condition := `
				cms_sections.name ILIKE ? OR 
				cms_sections.type ILIKE ? OR 
				EXTRACT(YEAR FROM cms_sections.created_at)::TEXT = ? OR 
				EXTRACT(YEAR FROM cms_sections.updated_at)::TEXT = ?
			`
			db = db.Where(condition, searchTerm, searchTerm, searchStr, searchStr)
			countDB = countDB.Where(condition, searchTerm, searchTerm, searchStr, searchStr)
		} else {
			condition := `
				cms_sections.name ILIKE ? OR 
				cms_sections.type ILIKE ?
			`
			db = db.Where(condition, searchTerm, searchTerm)
			countDB = countDB.Where(condition, searchTerm, searchTerm)
		}
	}

	err := countDB.Distinct("cms_sections.id").Count(&totalData).Error
	if err != nil {
		return nil, 0, err
	}

	allowedOrderCols := map[string]string{
		"id":            "cms_sections.id",
		"slug":          "cms_sections.slug",
		"name":          "cms_sections.name",
		"type":          "cms_sections.type",
		"section_order": "cms_sections.section_order",
		"is_repeatable": "cms_sections.is_repeatable",
		"is_system":     "cms_sections.is_system",
		"status":        "cms_sections.status",
		"created_at":    "cms_sections.created_at",
		"updated_at":    "cms_sections.updated_at",
	}

	finalOrderBy := "cms_sections.id"
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

	return data, totalData, nil
}

func (repository *manajemenCMSRepo) GetSectionById(id int) (*models.CMSSection, error) {
	var section models.CMSSection
	err := repository.dbSlave.First(&section, id).Error
	if err != nil {
		return nil, err
	}

	return &section, nil
}

func (repository *manajemenCMSRepo) UpdateSection(section *models.CMSSection) (*models.CMSSection, error) {
	err := repository.dbMaster.Save(section).Error
	if err != nil {
		return nil, err
	}
	return section, nil
}

func (repository *manajemenCMSRepo) DeleteSection(section *models.CMSSection) error {
	err := repository.dbMaster.Delete(section).Error
	if err != nil {
		return err
	}
	return nil
}

func (repository *manajemenCMSRepo) GetSectionBySlug(slug string) (*models.CMSSection, error) {
	var section models.CMSSection
	err := repository.dbSlave.Preload("Contents").Preload("Items").Preload("Media").Where("slug = ?", slug).First(&section).Error
	if err != nil {
		return nil, err
	}

	return &section, nil
}

func (repository *manajemenCMSRepo) UpsertContents(sectionID int, contents []models.CMSContent) error {
	return repository.dbMaster.Transaction(func(tx *gorm.DB) error {
		for _, c := range contents {
			var existing models.CMSContent
			err := tx.Where("section_id = ? AND key = ?", sectionID, c.Key).First(&existing).Error

			if errors.Is(err, gorm.ErrRecordNotFound) {
				if err := tx.Create(&c).Error; err != nil {
					return err
				}
				continue
			}
			if err != nil {
				return err
			}

			existing.ValueText = c.ValueText
			existing.Status = c.Status
			if err := tx.Save(&existing).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

func (repository *manajemenCMSRepo) UpsertItems(sectionID int, items []models.CMSItem) error {
	return repository.dbMaster.Transaction(func(tx *gorm.DB) error {
		for _, item := range items {
			var existing models.CMSItem
			err := tx.Where("section_id = ? AND item_order = ?", sectionID, item.ItemOrder).First(&existing).Error

			if errors.Is(err, gorm.ErrRecordNotFound) {
				if err := tx.Create(&item).Error; err != nil {
					return err
				}
				continue
			}
			if err != nil {
				return err
			}

			existing.Title = item.Title
			existing.Description = item.Description
			existing.Category = item.Category
			existing.Image = item.Image
			existing.Status = item.Status
			if err := tx.Save(&existing).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

func (repository *manajemenCMSRepo) UpsertMedia(sectionID int, media []models.CMSMedia) error {
	return repository.dbMaster.Transaction(func(tx *gorm.DB) error {
		for _, m := range media {
			var existing models.CMSMedia
			err := tx.Where("section_id = ? AND item_order = ?", sectionID, m.ItemOrder).First(&existing).Error

			if errors.Is(err, gorm.ErrRecordNotFound) {
				if err := tx.Create(&m).Error; err != nil {
					return err
				}
				continue
			}
			if err != nil {
				return err
			}

			existing.ImageURL = m.ImageURL
			existing.Caption = m.Caption
			existing.Status = m.Status
			if err := tx.Save(&existing).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

func (repository *manajemenCMSRepo) DeleteItem(sectionID int, itemOrder int) error {
	return repository.dbMaster.
		Where("section_id = ? AND item_order = ?", sectionID, itemOrder).
		Delete(&models.CMSItem{}).Error
}

func (repository *manajemenCMSRepo) DeleteMedia(sectionID int, itemOrder int) error {
	return repository.dbMaster.
		Where("section_id = ? AND item_order = ?", sectionID, itemOrder).
		Delete(&models.CMSMedia{}).Error
}

func (repository *manajemenCMSRepo) SyncContents(sectionID int, contents []models.CMSContent) error {
	return repository.dbMaster.Transaction(func(tx *gorm.DB) error {
		keepKeys := make([]string, 0, len(contents))

		for _, c := range contents {
			keepKeys = append(keepKeys, c.Key)

			var existing models.CMSContent
			err := tx.Where("section_id = ? AND key = ?", sectionID, c.Key).First(&existing).Error

			if errors.Is(err, gorm.ErrRecordNotFound) {
				if err := tx.Create(&c).Error; err != nil {
					return err
				}
				continue
			}
			if err != nil {
				return err
			}

			existing.ValueText = c.ValueText
			existing.Status = c.Status
			if err := tx.Save(&existing).Error; err != nil {
				return err
			}
		}

		return tx.Where("section_id = ? AND key NOT IN ?", sectionID, keepKeys).
			Delete(&models.CMSContent{}).Error
	})
}

func (repository *manajemenCMSRepo) SyncItems(sectionID int, items []models.CMSItem) error {
	return repository.dbMaster.Transaction(func(tx *gorm.DB) error {
		keepOrders := make([]int, 0, len(items))

		for _, item := range items {
			keepOrders = append(keepOrders, item.ItemOrder)

			var existing models.CMSItem
			err := tx.Where("section_id = ? AND item_order = ?", sectionID, item.ItemOrder).First(&existing).Error

			if errors.Is(err, gorm.ErrRecordNotFound) {
				if err := tx.Create(&item).Error; err != nil {
					return err
				}
				continue
			}
			if err != nil {
				return err
			}

			existing.Title = item.Title
			existing.Description = item.Description
			existing.Category = item.Category
			existing.Image = item.Image
			existing.Status = item.Status
			if err := tx.Save(&existing).Error; err != nil {
				return err
			}
		}

		return tx.Where("section_id = ? AND item_order NOT IN ?", sectionID, keepOrders).
			Delete(&models.CMSItem{}).Error
	})
}

func (repo *manajemenCMSRepo) GetLastSectionOrder() (int, error) {
	defer utils.GeneralRecover()

	var maxOrder int
	// KUNCI UTAMA: Tambahkan WHERE slug != 'footer' agar angka 999999 milik footer tidak terbawa
	err := repo.dbSlave.Model(&models.CMSSection{}).
		Where("slug != ?", "footer").
		Select("COALESCE(MAX(section_order), 0)").
		Scan(&maxOrder).Error

	if err != nil {
		return 0, err
	}

	// Jika belum ada data sama sekali (fallback safety), set ke 500 (starter terakhir)
	if maxOrder == 0 {
		maxOrder = 500
	}

	return maxOrder, nil
}

func (repo *manajemenCMSRepo) GetAdjacentSectionForReorder(currentOrder int, direction string) (*models.CMSSection, error) {
	defer utils.GeneralRecover()

	var adjacent models.CMSSection
	db := repo.dbSlave.Model(&models.CMSSection{}).Where("is_system = ?", false)

	if direction == "up" {
		err := db.Where("section_order < ?", currentOrder).
			Order("section_order DESC").
			First(&adjacent).Error
		if err != nil {
			return nil, err
		}
	} else if direction == "down" {
		// Cari section dinamis dengan order teratas di bawah currentOrder
		err := db.Where("section_order > ?", currentOrder).
			Order("section_order ASC").
			First(&adjacent).Error
		if err != nil {
			return nil, err
		}
	} else {
		return nil, errors.New("direction tidak valid")
	}

	return &adjacent, nil
}

func (repo *manajemenCMSRepo) SwapSectionOrder(sectionA, sectionB *models.CMSSection) error {
	defer utils.GeneralRecover()

	return repo.dbMaster.Transaction(func(tx *gorm.DB) error {
		tempOrder := sectionA.SectionOrder
		sectionA.SectionOrder = sectionB.SectionOrder
		sectionB.SectionOrder = tempOrder

		// Update Order Section A
		if err := tx.Model(&models.CMSSection{}).
			Where("id = ?", sectionA.ID).
			Update("section_order", sectionA.SectionOrder).Error; err != nil {
			return err
		}

		// Update Order Section B
		if err := tx.Model(&models.CMSSection{}).
			Where("id = ?", sectionB.ID).
			Update("section_order", sectionB.SectionOrder).Error; err != nil {
			return err
		}

		return nil
	})
}

func (repository *manajemenCMSRepo) SyncMedia(sectionID int, media []models.CMSMedia) (removedImagePaths []string, err error) {
	err = repository.dbMaster.Transaction(func(tx *gorm.DB) error {
		keepOrders := make([]int, 0, len(media))

		for _, m := range media {
			keepOrders = append(keepOrders, m.ItemOrder)

			var existing models.CMSMedia
			err := tx.Where("section_id = ? AND item_order = ?", sectionID, m.ItemOrder).First(&existing).Error

			if errors.Is(err, gorm.ErrRecordNotFound) {
				if err := tx.Create(&m).Error; err != nil {
					return err
				}
				continue
			}
			if err != nil {
				return err
			}

			// gambar item ini diganti dengan yang baru -> path lama perlu dihapus juga
			if existing.ImageURL != m.ImageURL {
				removedImagePaths = append(removedImagePaths, existing.ImageURL)
			}

			existing.ImageURL = m.ImageURL
			existing.Caption = m.Caption
			existing.Status = m.Status
			if err := tx.Save(&existing).Error; err != nil {
				return err
			}
		}

		// ambil row yang akan dihapus dulu supaya path-nya bisa dikembalikan ke service
		var toDelete []models.CMSMedia
		if err := tx.Where("section_id = ? AND item_order NOT IN ?", sectionID, keepOrders).
			Find(&toDelete).Error; err != nil {
			return err
		}
		for _, d := range toDelete {
			if d.ImageURL != "" {
				removedImagePaths = append(removedImagePaths, d.ImageURL)
			}
		}

		return tx.Where("section_id = ? AND item_order NOT IN ?", sectionID, keepOrders).
			Delete(&models.CMSMedia{}).Error
	})

	return removedImagePaths, err
}

// func (repository *manajemen)
