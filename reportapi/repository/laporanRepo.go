package repository

import (
	"backend/reportapi/models"
	"backend/reportapi/request"
	"backend/reportapi/utils"
	"context"
	"errors"
	"strings"
	"time"

	"gorm.io/gorm"
)

type LaporanRepo interface {
	UpdateLaporan(laporan models.Laporan) (*models.Laporan, error)
	GetLaporanByID(laporanID int64) (*models.Laporan, error)
	GetLaporanCoverByLaporanId(laporanID int64) (*models.LaporanCover, error)
	ReplaceLaporanCover(laporanID int64, konten models.LaporanCover) error

	DeleteLaporanKontenByLaporanID(laporanID int64) error
	UpdateLaporanKonten(konten models.LaporanCover) error

	GetListLaporan(respondentId int64, req request.LaporanDatatablePayload) ([]models.LaporanDatatable, int64, error)

	WithTransaction(ctx context.Context, fn func(txRepo LaporanRepo) error) error
	IsNameExists(ctx context.Context, name string) (bool, error)
	CreateLaporan(ctx context.Context, laporan models.Laporan) (*models.Laporan, error)
	CreateLaporanCover(ctx context.Context, konten models.LaporanCover) (*models.LaporanCover, error)
}

type laporanRepo struct {
	dbSlave  *gorm.DB
	dbMaster *gorm.DB
}

func NewLaporanRepo(dbSlave, dbMaster *gorm.DB) LaporanRepo {
	defer utils.GeneralRecover()
	return &laporanRepo{dbSlave, dbMaster}
}

func (repository *laporanRepo) WithTransaction(ctx context.Context, fn func(txRepo LaporanRepo) error) error {
	return repository.dbMaster.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txRepo := &laporanRepo{dbMaster: tx}
		return fn(txRepo)
	})
}

func (repository *laporanRepo) IsNameExists(ctx context.Context, name string) (bool, error) {
	defer utils.GeneralRecover()

	var count int64
	err := repository.dbMaster.WithContext(ctx).
		Model(&models.Laporan{}).
		Where("name = ?", name).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (repository *laporanRepo) CreateLaporan(ctx context.Context, laporan models.Laporan) (*models.Laporan, error) {
	defer utils.GeneralRecover()

	err := repository.dbMaster.WithContext(ctx).Create(&laporan).Error
	if err != nil {
		return nil, err
	}

	return &laporan, nil
}

func (repository *laporanRepo) CreateLaporanCover(ctx context.Context, konten models.LaporanCover) (*models.LaporanCover, error) {
	defer utils.GeneralRecover()

	err := repository.dbMaster.WithContext(ctx).Create(&konten).Error
	if err != nil {
		return nil, err
	}

	return &konten, nil
}

func (repository *laporanRepo) UpdateLaporan(laporan models.Laporan) (*models.Laporan, error) {
	defer utils.GeneralRecover()

	err := repository.dbMaster.Save(laporan).Error
	if err != nil {
		return nil, err
	}
	return &laporan, nil
}

func (repository *laporanRepo) GetLaporanByID(laporanID int64) (*models.Laporan, error) {
	defer utils.GeneralRecover()

	var laporan *models.Laporan
	err := repository.dbSlave.Table("laporans").
		Where("id = ?", laporanID).
		First(&laporan).Error
	if err != nil {
		return nil, err
	}

	return laporan, nil
}

func (repository *laporanRepo) DeleteLaporanKontenByLaporanID(laporanID int64) error {
	defer utils.GeneralRecover()

	err := repository.dbMaster.Table("laporan_kontens").
		Where("laporan_id = ?", laporanID).
		Delete(&models.LaporanCover{}).Error
	if err != nil {
		return err
	}

	return nil
}

func (repository *laporanRepo) GetLaporanCoverByLaporanId(laporanID int64) (*models.LaporanCover, error) {
	defer utils.GeneralRecover()

	var cover models.LaporanCover
	err := repository.dbSlave.Table("laporan_covers").
		Where("laporan_id = ?", laporanID).
		First(&cover).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return &cover, nil
}

func (repository *laporanRepo) UpdateLaporanKonten(konten models.LaporanCover) error {
	defer utils.GeneralRecover()

	now := time.Now()

	updates := map[string]interface{}{
		"laporan_id":     konten.LaporanID,
		"text_depan":     konten.TextDepan,
		"img_depan":      konten.ImgDepan,
		"text_belakang":  konten.TextBelakang,
		"img_belakang":   konten.ImgBelakang,
		"kata_pengantar": konten.KataPengantar,
		"updated_at":     now,
	}

	if konten.ID != 0 {
		return repository.dbMaster.Table("laporan_kontens").
			Where("id = ?", konten.ID).
			Updates(updates).Error
	}

	updates["created_at"] = now
	return repository.dbMaster.Table("laporan_kontens").Create(&updates).Error
}

func (repository *laporanRepo) ReplaceLaporanCover(laporanID int64, konten models.LaporanCover) error {
	defer utils.GeneralRecover()

	return repository.dbMaster.Transaction(func(tx *gorm.DB) error {
		if err := tx.Table("laporan_covers").
			Where("laporan_id = ?", laporanID).
			Delete(&models.LaporanCover{}).Error; err != nil {
			return err
		}

		now := time.Now()
		konten.ID = 0
		konten.CreatedAt = &now
		konten.UpdatedAt = &now

		if err := tx.Table("laporan_covers").Create(&konten).Error; err != nil {
			return err
		}

		return nil
	})
}

func (repository *laporanRepo) GetListLaporan(respondentId int64, req request.LaporanDatatablePayload) ([]models.LaporanDatatable, int64, error) {
	defer utils.GeneralRecover()
	var data []models.LaporanDatatable
	var totalData int64

	db := repository.dbSlave.Table("laporans")

	if respondentId != 0 {
		db = db.Where("laporans.respondent_id = ?", respondentId)
	}

	if req.Search != "" {
		searchTerm := "%" + req.Search + "%"
		searchStr := strings.TrimSpace(req.Search)
		parsedDate, isDate := utils.TryParseIndonesianDate(req.Search)

		if isDate {
			db = db.Where(`
				laporans.name ILIKE ? OR
				DATE(laporans.created_at) = ? OR
				DATE(laporans.updated_at) = ?
			`, searchTerm, parsedDate, parsedDate)
		} else if len(searchStr) == 4 {
			db = db.Where(`
				laporans.name ILIKE ? OR
				EXTRACT(YEAR FROM laporans.created_at)::TEXT = ? OR
				EXTRACT(YEAR FROM laporans.updated_at)::TEXT = ?
			`, searchTerm, searchStr, searchStr)
		} else {
			db = db.Where(`
				laporans.name ILIKE ?
			`, searchTerm)
		}
	}

	err := db.Count(&totalData).Error
	if err != nil {
		return nil, 0, err
	}

	db = db.Select(`
		laporans.id,
		laporans.name,
		laporans.updated_at,
		laporans.created_at
	`)

	if req.OrderBy != "" {
		finalOrderBy := "laporans.id"
		finalOrderDir := "desc"

		allowedOrderCols := map[string]string{
			"id":         "laporans.id",
			"name":       "laporans.name",
			"created_at": "laporans.created_at",
			"updated_at": "laporans.updated_at",
		}

		if mappedCol, isAllowed := allowedOrderCols[req.OrderBy]; isAllowed {
			finalOrderBy = mappedCol
		}

		if strings.ToLower(req.OrderDir) == "asc" {
			finalOrderDir = "asc"
		}

		db = db.Order(finalOrderBy + " " + finalOrderDir)
	} else {
		db = db.Order("laporans.created_at desc")
	}

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
