package repository

import (
	"backend/reportapi/models"
	"backend/reportapi/request"
	"backend/reportapi/utils"
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"gorm.io/gorm"
)

func parseJSONInt64Array(raw []byte) ([]int64, error) {
	if len(raw) == 0 {
		return nil, nil
	}

	var asInts []int64
	if err := json.Unmarshal(raw, &asInts); err == nil {
		return asInts, nil
	}

	var asStrings []string
	if err := json.Unmarshal(raw, &asStrings); err != nil {
		return nil, err
	}

	result := make([]int64, 0, len(asStrings))
	for _, s := range asStrings {
		var v int64
		if _, err := fmt.Sscanf(s, "%d", &v); err == nil {
			result = append(result, v)
		}
	}
	return result, nil
}

type TableDataRow struct {
	TerritoryName string
	Values        map[int]int64 // Key: FormFieldID -> Value: Angka/Jumlah Jawaban
}

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
	GetListSectionReport(laporanId int64, req request.SectionLaporanDatatablePayload) ([]models.LaporanSectionDatatable, int64, error)

	CheckSectionTitleExists(reportID int64, title string) (bool, error)
	CheckSubSectionTitleExists(reportID int64, title string) (bool, error)
	GetMaxSectionSequence(reportID int64) (int, error)
	CreateSectionTx(section *models.ReportSection) error
	GetFormFieldLabels(fieldIDs []int) (map[int]string, error)
	GetFormFieldTypes(fieldIDs []int) (map[int]string, error)

	GetSectionByID(sectionID int64) (*models.ReportSection, error)

	UpdateSectionMeta(tx *gorm.DB, sectionID int64, title string, hasSubSection *bool) error
	DeleteSubSectionsBySectionID(tx *gorm.DB, sectionID int64) error
	DeleteComponentsBySectionID(tx *gorm.DB, sectionID int64) error
	CreateSubSectionsTx(tx *gorm.DB, sectionID int64, subSections []models.ReportSubSection) error
	CreateComponentsForSectionTx(tx *gorm.DB, sectionID int64, components []models.ReportComponent) error

	UpdateSectionMetaTx(sectionID int64, title string, hasSubSection *bool) error
	DeleteSubSectionsBySectionIDTx(sectionID int64) error
	DeleteComponentsBySectionIDTx(sectionID int64) error
	CreateSubSectionsTxWrapped(sectionID int64, subSections []models.ReportSubSection) error
	CreateComponentsForSectionTxWrapped(sectionID int64, components []models.ReportComponent) error

	CheckSectionTitleExistsExcludingID(reportID int64, title string, excludeSectionID int64) (bool, error)
	CheckSubSectionTitleExistsExcludingSection(reportID int64, title string, excludeSectionID int64) (bool, error)

	RunInTransaction(fn func(txRepo LaporanRepo) error) error
	DeleteReportByID(reportID int64) error
	DeleteSectionReportBySectionID(sectionID int64) error

	GetKelurahanIDsByKecamatanID(kecamatanID int64) ([]int64, error)
	GetRWIDsByKelurahanID(kelurahanID int64) ([]int64, error)
	ValidateKecamatanExist(kecamatanIDs []int64) (bool, error)
	ValidateKelurahanExist(kelurahanIDs []int64) (bool, error)
}

type laporanRepo struct {
	dbSlave  *gorm.DB
	dbMaster *gorm.DB
}

func NewLaporanRepo(dbSlave, dbMaster *gorm.DB) LaporanRepo {
	defer utils.GeneralRecover()
	return &laporanRepo{
		dbSlave:  dbSlave,
		dbMaster: dbMaster,
	}
}

func (r *laporanRepo) UpdateReport(laporan models.Report) (*models.Report, error) {
	defer utils.GeneralRecover()
	err := r.dbMaster.Save(&laporan).Error
	if err != nil {
		return nil, err
	}
	return &laporan, nil
}

func (r *laporanRepo) GetReportByID(laporanID int64) (*models.Report, error) {
	defer utils.GeneralRecover()
	var report models.Report
	err := r.dbSlave.
		Preload("Cover").
		Preload("Sections", func(db *gorm.DB) *gorm.DB {
			return db.Order("report_sections.sequence ASC")
		}).
		Preload("Sections.SubSections", func(db *gorm.DB) *gorm.DB {
			return db.Order("report_subsections.sequence ASC")
		}).
		Preload("Sections.SubSections.Components", func(db *gorm.DB) *gorm.DB {
			return db.Order("report_components.sequence ASC")
		}).
		Preload("Sections.Components", func(db *gorm.DB) *gorm.DB {
			return db.Order("report_components.sequence ASC")
		}).
		Where("id = ?", laporanID).
		First(&report).Error

	if err != nil {
		return nil, err
	}

	return &report, nil
}

func (r *laporanRepo) GetReportCoverByReportId(laporanID int64) (*models.ReportCover, error) {
	defer utils.GeneralRecover()
	var cover models.ReportCover
	err := r.dbSlave.Where("report_id = ?", laporanID).First(&cover).Error
	if err != nil {
		return nil, err
	}
	return &cover, nil
}

func (r *laporanRepo) ReplaceReportCover(laporanID int64, konten models.ReportCover) error {
	defer utils.GeneralRecover()
	return r.dbMaster.Where("report_id = ?", laporanID).Delete(&models.ReportCover{}).Create(&konten).Error
}

func (r *laporanRepo) DeleteReportKontenByReportID(laporanID int64) error {
	defer utils.GeneralRecover()
	return r.dbMaster.Where("report_id = ?", laporanID).Delete(&models.ReportCover{}).Error
}

func (r *laporanRepo) UpdateReportKonten(konten models.ReportCover) error {
	defer utils.GeneralRecover()
	return r.dbMaster.Save(&konten).Error
}

func (r *laporanRepo) GetListReport(respondentId int64, req request.LaporanDatatablePayload) ([]models.LaporanDatatable, int64, error) {
	defer utils.GeneralRecover()
	var data []models.LaporanDatatable
	var totalData int64

	db := r.dbSlave.Model(&models.Report{}).Where("respondent_id = ?", respondentId)

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

		db = db.Order(fmt.Sprintf("%s %s", finalOrderBy, finalOrderDir))
	} else {
		db = db.Order("reports.id desc")
	}

	limit := req.Limit
	if limit <= 0 {
		limit = 10
	}
	page := req.Page
	if page <= 0 {
		page = 1
	}
	offset := (page - 1) * limit

	err = db.Limit(limit).Offset(offset).Scan(&data).Error
	if err != nil {
		return nil, 0, err
	}

	for i := range data {
		data[i].No = int64(offset + i + 1)
	}

	return data, totalData, nil
}

func (r *laporanRepo) WithTransaction(ctx context.Context, fn func(txRepo LaporanRepo) error) error {
	defer utils.GeneralRecover()
	return r.dbMaster.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txRepo := &laporanRepo{
			dbSlave:  tx,
			dbMaster: tx,
		}
		return fn(txRepo)
	})
}

func (r *laporanRepo) IsNameExists(ctx context.Context, name string) (bool, error) {
	defer utils.GeneralRecover()
	var count int64
	err := r.dbSlave.WithContext(ctx).Model(&models.Report{}).
		Where("LOWER(name) = LOWER(?)", name).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *laporanRepo) CreateReport(ctx context.Context, laporan models.Report) (*models.Report, error) {
	defer utils.GeneralRecover()
	err := r.dbMaster.WithContext(ctx).Create(&laporan).Error
	if err != nil {
		return nil, err
	}
	return &laporan, nil
}

func (r *laporanRepo) CreateReportCover(ctx context.Context, konten models.ReportCover) (*models.ReportCover, error) {
	defer utils.GeneralRecover()
	err := r.dbMaster.WithContext(ctx).Create(&konten).Error
	if err != nil {
		return nil, err
	}
	return &konten, nil
}

func (r *laporanRepo) GetListSectionReport(laporanId int64, req request.SectionLaporanDatatablePayload) ([]models.LaporanSectionDatatable, int64, error) {
	defer utils.GeneralRecover()
	var data []models.LaporanSectionDatatable
	var totalData int64

	db := r.dbSlave.Model(&models.ReportSection{}).Where("report_id = ?", laporanId)

	if req.Search != "" {
		searchTerm := "%" + req.Search + "%"
		searchStr := strings.TrimSpace(req.Search)
		parsedDate, isDate := utils.TryParseIndonesianDate(req.Search)

		if isDate {
			db = db.Where(`
				report_sections.title ILIKE ? OR
				DATE(report_sections.created_at) = ? OR
				DATE(report_sections.updated_at) = ?
			`, searchTerm, parsedDate, parsedDate)
		} else if len(searchStr) == 4 {
			db = db.Where(`
				report_sections.title ILIKE ? OR
				EXTRACT(YEAR FROM report_sections.created_at)::TEXT = ? OR
				EXTRACT(YEAR FROM report_sections.updated_at)::TEXT = ?
			`, searchTerm, searchStr, searchStr)
		} else {
			db = db.Where(`
				report_sections.title ILIKE ?
			`, searchTerm)
		}
	}

	err := db.Count(&totalData).Error
	if err != nil {
		return nil, 0, err
	}

	db = db.Select(`
		report_sections.id,
		report_sections.report_id,
		report_sections.sequence,
		report_sections.title,
		report_sections.updated_at,
		report_sections.created_at,
		CASE 
			WHEN report_sections.has_sub_section = false THEN 1 
			ELSE (
				SELECT COUNT(id) 
				FROM report_subsections 
				WHERE report_subsections.report_section_id = report_sections.id
			) 
		END AS total_konten
	`)

	if req.OrderBy != "" {
		finalOrderBy := "report_sections.id"
		finalOrderDir := "desc"

		allowedOrderCols := map[string]string{
			"id":         "report_sections.id",
			"name":       "report_sections.title",
			"created_at": "report_sections.created_at",
			"updated_at": "report_sections.updated_at",
		}

		if mappedCol, isAllowed := allowedOrderCols[req.OrderBy]; isAllowed {
			finalOrderBy = mappedCol
		}

		if strings.ToLower(req.OrderDir) == "asc" {
			finalOrderDir = "asc"
		}

		db = db.Order(fmt.Sprintf("%s %s", finalOrderBy, finalOrderDir))
	} else {
		db = db.Order("report_sections.id desc")
	}

	limit := req.Limit
	if limit <= 0 {
		limit = 10
	}
	page := req.Page
	if page <= 0 {
		page = 1
	}
	offset := (page - 1) * limit

	err = db.Limit(limit).Offset(offset).Scan(&data).Error
	if err != nil {
		return nil, 0, err
	}

	for i := range data {
		data[i].No = int64(offset + i + 1)
	}

	return data, totalData, nil
}

func (r *laporanRepo) CheckSectionTitleExists(reportID int64, title string) (bool, error) {
	defer utils.GeneralRecover()
	var count int64
	err := r.dbSlave.Model(&models.ReportSection{}).
		Where("report_id = ? AND LOWER(title) = LOWER(?)", reportID, title).
		Count(&count).Error

	return count > 0, err
}

func (r *laporanRepo) CheckSubSectionTitleExists(reportID int64, title string) (bool, error) {
	defer utils.GeneralRecover()
	var count int64
	err := r.dbSlave.Table("report_subsections").
		Joins("JOIN report_sections ON report_sections.id = report_subsections.report_section_id").
		Where("report_sections.report_id = ? AND LOWER(report_subsections.title) = LOWER(?)", reportID, title).
		Count(&count).Error

	return count > 0, err
}

func (r *laporanRepo) GetMaxSectionSequence(reportID int64) (int, error) {
	defer utils.GeneralRecover()
	var maxSeq int
	err := r.dbSlave.Model(&models.ReportSection{}).
		Where("report_id = ?", reportID).
		Select("COALESCE(MAX(sequence), 0)").
		Scan(&maxSeq).Error

	if err != nil {
		return 0, err
	}

	return maxSeq, nil
}

func (r *laporanRepo) CreateSectionTx(section *models.ReportSection) error {
	defer utils.GeneralRecover()
	return r.dbMaster.Create(section).Error
}

func (r *laporanRepo) GetFormFieldLabels(fieldIDs []int) (map[int]string, error) {
	defer utils.GeneralRecover()
	result := make(map[int]string)
	if len(fieldIDs) == 0 {
		return result, nil
	}

	type FieldLabel struct {
		ID       int    `gorm:"column:id"`
		Question string `gorm:"column:question"` // Ubah dari Title menjadi Question
	}
	var fields []FieldLabel
	err := r.dbSlave.Table("form_fields").
		Select("id, question"). // Ambil kolom question
		Where("id IN ?", fieldIDs).
		Scan(&fields).Error

	if err != nil {
		return result, nil
	}

	for _, f := range fields {
		result[f.ID] = f.Question
	}

	return result, nil
}

func (r *laporanRepo) GetFormFieldTypes(fieldIDs []int) (map[int]string, error) {
	defer utils.GeneralRecover()
	result := make(map[int]string)
	if len(fieldIDs) == 0 {
		return result, nil
	}

	type FieldType struct {
		ID       int    `gorm:"column:id"`
		Template string `gorm:"column:template"`
	}
	var fields []FieldType
	err := r.dbSlave.Table("form_fields").
		Select("id, template").
		Where("id IN ?", fieldIDs).
		Scan(&fields).Error

	if err != nil {
		return result, err
	}

	for _, f := range fields {
		result[f.ID] = f.Template
	}

	return result, nil
}

func (r *laporanRepo) GetSectionByID(sectionID int64) (*models.ReportSection, error) {
	defer utils.GeneralRecover()
	var section models.ReportSection

	err := r.dbSlave.
		Preload("SubSections", func(db *gorm.DB) *gorm.DB {
			return db.Order("report_subsections.sequence ASC")
		}).
		Preload("SubSections.Components", func(db *gorm.DB) *gorm.DB {
			return db.Order("report_components.sequence ASC")
		}).
		Preload("Components", func(db *gorm.DB) *gorm.DB {
			return db.Order("report_components.sequence ASC")
		}).
		Where("id = ?", sectionID).
		First(&section).Error

	if err != nil {
		return nil, err
	}
	return &section, nil
}

func (r *laporanRepo) UpdateSectionMeta(tx *gorm.DB, sectionID int64, title string, hasSubSection *bool) error {
	defer utils.GeneralRecover()
	return tx.Model(&models.ReportSection{}).
		Where("id = ?", sectionID).
		Updates(map[string]interface{}{
			"title":           title,
			"has_sub_section": hasSubSection,
		}).Error
}

func (r *laporanRepo) DeleteSubSectionsBySectionID(tx *gorm.DB, sectionID int64) error {
	defer utils.GeneralRecover()
	return tx.Where("report_section_id = ?", sectionID).Delete(&models.ReportSubSection{}).Error
}

func (r *laporanRepo) DeleteComponentsBySectionID(tx *gorm.DB, sectionID int64) error {
	defer utils.GeneralRecover()
	return tx.Where("report_section_id = ?", sectionID).Delete(&models.ReportComponent{}).Error
}

func (r *laporanRepo) CreateSubSectionsTx(tx *gorm.DB, sectionID int64, subSections []models.ReportSubSection) error {
	defer utils.GeneralRecover()
	for i := range subSections {
		subSections[i].ReportSectionID = sectionID
	}
	if len(subSections) == 0 {
		return nil
	}
	return tx.Create(&subSections).Error
}

func (r *laporanRepo) CreateComponentsForSectionTx(tx *gorm.DB, sectionID int64, components []models.ReportComponent) error {
	defer utils.GeneralRecover()
	for i := range components {
		components[i].ReportSectionID = &sectionID
	}
	if len(components) == 0 {
		return nil
	}
	return tx.Create(&components).Error
}

func (r *laporanRepo) UpdateSectionMetaTx(sectionID int64, title string, hasSubSection *bool) error {
	return r.UpdateSectionMeta(r.dbMaster, sectionID, title, hasSubSection)
}
func (r *laporanRepo) DeleteSubSectionsBySectionIDTx(sectionID int64) error {
	return r.DeleteSubSectionsBySectionID(r.dbMaster, sectionID)
}
func (r *laporanRepo) DeleteComponentsBySectionIDTx(sectionID int64) error {
	return r.DeleteComponentsBySectionID(r.dbMaster, sectionID)
}
func (r *laporanRepo) CreateSubSectionsTxWrapped(sectionID int64, subSections []models.ReportSubSection) error {
	return r.CreateSubSectionsTx(r.dbMaster, sectionID, subSections)
}
func (r *laporanRepo) CreateComponentsForSectionTxWrapped(sectionID int64, components []models.ReportComponent) error {
	return r.CreateComponentsForSectionTx(r.dbMaster, sectionID, components)
}

func (r *laporanRepo) CheckSectionTitleExistsExcludingID(reportID int64, title string, excludeSectionID int64) (bool, error) {
	defer utils.GeneralRecover()
	var count int64
	err := r.dbSlave.Model(&models.ReportSection{}).
		Where("report_id = ? AND LOWER(title) = LOWER(?) AND id != ?", reportID, title, excludeSectionID).
		Count(&count).Error
	return count > 0, err
}

func (r *laporanRepo) CheckSubSectionTitleExistsExcludingSection(reportID int64, title string, excludeSectionID int64) (bool, error) {
	defer utils.GeneralRecover()
	var count int64
	err := r.dbSlave.Table("report_subsections").
		Joins("JOIN report_sections ON report_sections.id = report_subsections.report_section_id").
		Where("report_sections.report_id = ? AND LOWER(report_subsections.title) = LOWER(?) AND report_subsections.report_section_id != ?", reportID, title, excludeSectionID).
		Count(&count).Error
	return count > 0, err
}

func (repository *laporanRepo) RunInTransaction(fn func(txRepo LaporanRepo) error) error {
	tx := repository.dbMaster.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	txRepo := &laporanRepo{
		dbSlave:  tx,
		dbMaster: tx,
	}

	err := fn(txRepo)
	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

func (repository *laporanRepo) DeleteReportByID(reportID int64) error {
	defer utils.GeneralRecover()
	tx := repository.dbMaster.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	// Hapus komponen terkait
	if err := tx.Where("report_section_id IN (?)", tx.Model(&models.ReportSection{}).Select("id").Where("report_id = ?", reportID)).Delete(&models.ReportComponent{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Hapus sub-sections terkait
	if err := tx.Where("report_section_id IN (?)", tx.Model(&models.ReportSection{}).Select("id").Where("report_id = ?", reportID)).Delete(&models.ReportSubSection{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Hapus sections terkait
	if err := tx.Where("report_id = ?", reportID).Delete(&models.ReportSection{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Hapus cover terkait
	if err := tx.Where("report_id = ?", reportID).Delete(&models.ReportCover{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Where("id = ?", reportID).Delete(&models.Report{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

func (repository *laporanRepo) DeleteSectionReportBySectionID(sectionID int64) error {
	defer utils.GeneralRecover()
	tx := repository.dbMaster.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	// Hapus komponen terkait
	if err := tx.Where("report_section_id = ?", sectionID).Delete(&models.ReportComponent{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Hapus sub-sections terkait
	if err := tx.Where("report_section_id = ?", sectionID).Delete(&models.ReportSubSection{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Hapus section terkait
	if err := tx.Where("id = ?", sectionID).Delete(&models.ReportSection{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

func (r *laporanRepo) GetRWIDsByKelurahanID(kelurahanID int64) ([]int64, error) {
	defer utils.GeneralRecover()
	var rwIDs []int64
	err := r.dbSlave.Table("data__rws").
		Where("kelurahan_id = ? AND deleted_at IS NULL", kelurahanID).
		Pluck("id", &rwIDs).Error
	return rwIDs, err
}

func (r *laporanRepo) ValidateKecamatanExist(kecamatanIDs []int64) (bool, error) {
	defer utils.GeneralRecover()
	var count int64
	err := r.dbSlave.Table("kecamatans").
		Where("id IN ? AND deleted_at IS NULL", kecamatanIDs).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count == int64(len(kecamatanIDs)), nil
}

func (r *laporanRepo) ValidateKelurahanExist(kelurahanIDs []int64) (bool, error) {
	defer utils.GeneralRecover()
	var count int64
	err := r.dbSlave.Table("kelurahans").
		Where("id IN ? AND deleted_at IS NULL", kelurahanIDs).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count == int64(len(kelurahanIDs)), nil
}

func (r *laporanRepo) GetKelurahanIDsByKecamatanID(kecamatanID int64) ([]int64, error) {
	defer utils.GeneralRecover()
	var kelurahanIDs []int64
	err := r.dbSlave.Table("kelurahans").
		Where("sub_district_id = ? AND deleted_at IS NULL", kecamatanID).
		Pluck("id", &kelurahanIDs).Error
	return kelurahanIDs, err
}
