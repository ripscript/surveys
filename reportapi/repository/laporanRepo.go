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
	UpdateReport(laporan models.Report) (*models.Report, error)
	GetReportByID(laporanID int64) (*models.Report, error)
	GetReportCoverByReportId(laporanID int64) (*models.ReportCover, error)
	ReplaceReportCover(laporanID int64, konten models.ReportCover) error

	DeleteReportKontenByReportID(laporanID int64) error
	UpdateReportKonten(konten models.ReportCover) error

	GetListReport(respondentId int64, req request.LaporanDatatablePayload) ([]models.LaporanDatatable, int64, error)

	WithTransaction(ctx context.Context, fn func(txRepo LaporanRepo) error) error
	IsNameExists(ctx context.Context, name string) (bool, error)
	CreateReport(ctx context.Context, laporan models.Report) (*models.Report, error)
	CreateReportCover(ctx context.Context, konten models.ReportCover) (*models.ReportCover, error)
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
		Model(&models.Report{}).
		Where("name = ?", name).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (repository *laporanRepo) CreateReport(ctx context.Context, laporan models.Report) (*models.Report, error) {
	defer utils.GeneralRecover()

	err := repository.dbMaster.WithContext(ctx).Create(&laporan).Error
	if err != nil {
		return nil, err
	}

	return &laporan, nil
}

func (repository *laporanRepo) CreateReportCover(ctx context.Context, konten models.ReportCover) (*models.ReportCover, error) {
	defer utils.GeneralRecover()

	err := repository.dbMaster.WithContext(ctx).Create(&konten).Error
	if err != nil {
		return nil, err
	}

	return &konten, nil
}

func (repository *laporanRepo) UpdateReport(laporan models.Report) (*models.Report, error) {
	defer utils.GeneralRecover()

	err := repository.dbMaster.Save(laporan).Error
	if err != nil {
		return nil, err
	}
	return &laporan, nil
}

func (repository *laporanRepo) GetReportByID(laporanID int64) (*models.Report, error) {
	defer utils.GeneralRecover()

	var laporan *models.Report
	err := repository.dbSlave.Table("reports").
		Where("id = ?", laporanID).
		First(&laporan).Error
	if err != nil {
		return nil, err
	}

	return laporan, nil
}

func (repository *laporanRepo) DeleteReportKontenByReportID(laporanID int64) error {
	defer utils.GeneralRecover()

	err := repository.dbMaster.Table("report_covers").
		Where("report_id = ?", laporanID).
		Delete(&models.ReportCover{}).Error
	if err != nil {
		return err
	}

	return nil
}

func (repository *laporanRepo) GetReportCoverByReportId(laporanID int64) (*models.ReportCover, error) {
	defer utils.GeneralRecover()

	var cover models.ReportCover
	err := repository.dbSlave.Table("report_covers").
		Where("report_id = ?", laporanID).
		First(&cover).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return &cover, nil
}

func (repository *laporanRepo) UpdateReportKonten(konten models.ReportCover) error {
	defer utils.GeneralRecover()

	now := time.Now()

	updates := map[string]interface{}{
		"report_id":      konten.ReportID,
		"text_depan":     konten.TextDepan,
		"img_depan":      konten.ImgDepan,
		"text_belakang":  konten.TextBelakang,
		"img_belakang":   konten.ImgBelakang,
		"kata_pengantar": konten.KataPengantar,
		"updated_at":     now,
	}

	if konten.ID != 0 {
		return repository.dbMaster.Table("report_covers").
			Where("id = ?", konten.ID).
			Updates(updates).Error
	}

	updates["created_at"] = now
	return repository.dbMaster.Table("report_covers").Create(&updates).Error
}

func (repository *laporanRepo) ReplaceReportCover(laporanID int64, konten models.ReportCover) error {
	defer utils.GeneralRecover()

	return repository.dbMaster.Transaction(func(tx *gorm.DB) error {
		if err := tx.Table("report_covers").
			Where("report_id = ?", laporanID).
			Delete(&models.ReportCover{}).Error; err != nil {
			return err
		}

		now := time.Now()
		konten.ID = 0
		konten.CreatedAt = &now
		konten.UpdatedAt = &now

		if err := tx.Table("report_covers").Create(&konten).Error; err != nil {
			return err
		}

		return nil
	})
}

func (repository *laporanRepo) GetListReport(respondentId int64, req request.LaporanDatatablePayload) ([]models.LaporanDatatable, int64, error) {
	defer utils.GeneralRecover()
	var data []models.LaporanDatatable
	var totalData int64

	db := repository.dbSlave.Table("reports")

	if respondentId != 0 {
		db = db.Where("reports.respondent_id = ?", respondentId)
	}

	if req.Search != "" {
		searchTerm := "%" + req.Search + "%"
		searchStr := strings.TrimSpace(req.Search)
		parsedDate, isDate := utils.TryParseIndonesianDate(req.Search)

		if isDate {
			db = db.Where(`
				reports.name ILIKE ? OR
				DATE(reports.created_at) = ? OR
				DATE(reports.updated_at) = ?
			`, searchTerm, parsedDate, parsedDate)
		} else if len(searchStr) == 4 {
			db = db.Where(`
				reports.name ILIKE ? OR
				EXTRACT(YEAR FROM reports.created_at)::TEXT = ? OR
				EXTRACT(YEAR FROM reports.updated_at)::TEXT = ?
			`, searchTerm, searchStr, searchStr)
		} else {
			db = db.Where(`
				reports.name ILIKE ?
			`, searchTerm)
		}
	}

	err := db.Count(&totalData).Error
	if err != nil {
		return nil, 0, err
	}

	db = db.Select(`
		reports.id,
		reports.name,
		reports.updated_at,
		reports.created_at
	`)

	if req.OrderBy != "" {
		finalOrderBy := "reports.id"
		finalOrderDir := "desc"

		allowedOrderCols := map[string]string{
			"id":         "reports.id",
			"name":       "reports.name",
			"created_at": "reports.created_at",
			"updated_at": "reports.updated_at",
		}

		if mappedCol, isAllowed := allowedOrderCols[req.OrderBy]; isAllowed {
			finalOrderBy = mappedCol
		}

		if strings.ToLower(req.OrderDir) == "asc" {
			finalOrderDir = "asc"
		}

		db = db.Order(finalOrderBy + " " + finalOrderDir)
	} else {
		db = db.Order("reports.created_at desc")
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
