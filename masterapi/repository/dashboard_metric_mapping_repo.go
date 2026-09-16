package repository

import (
	"backend/masterapi/models"
	"backend/masterapi/payloads"
	"backend/masterapi/utils"
	"context"
	"errors"
	"strings"
	"time"

	"gorm.io/gorm"
)

type DashboardMetricMappingRepository interface {
	Create(ctx context.Context, m *models.DashboardMetricMapping) error
	Update(ctx context.Context, m *models.DashboardMetricMapping) error
	Delete(ctx context.Context, id int64) error
	FindByID(ctx context.Context, id int64) (*models.DashboardMetricMapping, error)
	// FindByMetricAndForm dipakai untuk cek constraint UNIQUE(dashboard_metric_id, form_id) sebelum insert
	FindByMetricAndForm(ctx context.Context, metricID, formID int64) (*models.DashboardMetricMapping, error)
	// ListByFormAndCategory adalah query utama yang dipakai endpoint dashboard runtime nanti
	ListByFormAndCategory(ctx context.Context, formID int64, category string) ([]models.DashboardMetricMapping, error)

	GetList(ctx context.Context, req payloads.DatatablePayload, formCode string, dashboardMetricID int64) ([]MappingListRow, int64, error)
	FindDetailByID(ctx context.Context, id int64) (*models.DashboardMetricMapping, error)
	CountOptionStatusByMappingID(ctx context.Context, mappingID int64) (int64, error)
	CountByMetricID(ctx context.Context, metricID int64) (int64, error)
}

type dashboardMetricMappingRepository struct {
	dbMaster *gorm.DB
	dbSlave  *gorm.DB
}

func NewDashboardMetricMappingRepository(dbMaster, dbSlave *gorm.DB) DashboardMetricMappingRepository {
	return &dashboardMetricMappingRepository{dbMaster: dbMaster, dbSlave: dbSlave}
}

type MappingListRow struct {
	ID                int64
	DashboardMetricID int64
	MetricKey         string
	MetricLabel       string
	ExpectedTemplate  string
	Category          string
	FormID            int64
	FormTitle         string
	FormCode          string
	FormFieldID       int64
	FormFieldQuestion string
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

func (r *dashboardMetricMappingRepository) Create(ctx context.Context, m *models.DashboardMetricMapping) error {
	return r.dbMaster.WithContext(ctx).Create(m).Error
}

func (r *dashboardMetricMappingRepository) Update(ctx context.Context, m *models.DashboardMetricMapping) error {
	return r.dbMaster.WithContext(ctx).Save(m).Error
}

func (r *dashboardMetricMappingRepository) Delete(ctx context.Context, id int64) error {
	return r.dbMaster.WithContext(ctx).Delete(&models.DashboardMetricMapping{}, id).Error
}

func (r *dashboardMetricMappingRepository) FindByID(ctx context.Context, id int64) (*models.DashboardMetricMapping, error) {
	var m models.DashboardMetricMapping
	err := r.dbSlave.WithContext(ctx).Preload("DashboardMetric").First(&m, id).Error
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *dashboardMetricMappingRepository) FindByMetricAndForm(ctx context.Context, metricID, formID int64) (*models.DashboardMetricMapping, error) {
	var m models.DashboardMetricMapping
	err := r.dbSlave.WithContext(ctx).
		Where("dashboard_metric_id = ? AND form_id = ?", metricID, formID).
		First(&m).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *dashboardMetricMappingRepository) ListByFormAndCategory(ctx context.Context, formID int64, category string) ([]models.DashboardMetricMapping, error) {
	var list []models.DashboardMetricMapping
	err := r.dbSlave.WithContext(ctx).
		Joins("JOIN dashboard_metrics dm ON dm.id = dashboard_metric_mappings.dashboard_metric_id").
		Where("dashboard_metric_mappings.form_id = ? AND dm.category = ?", formID, category).
		Preload("DashboardMetric").
		Find(&list).Error
	return list, err
}

func (r *dashboardMetricMappingRepository) GetList(ctx context.Context, req payloads.DatatablePayload, formCode string, dashboardMetricID int64) ([]MappingListRow, int64, error) {
	defer utils.GeneralRecover()

	var data []MappingListRow
	var totalData int64

	db := r.dbSlave.WithContext(ctx).
		Table("dashboard_metric_mappings dmm").
		Select(`
			dmm.id,
			dmm.dashboard_metric_id,
			dm.metric_key,
			dm.label as metric_label,
			dm.expected_template,
			dm.category,
			dmm.form_id,
			f.title as form_title,
			f.code as form_code,
			dmm.form_field_id,
			ff.question as form_field_question,
			dmm.created_at,
			dmm.updated_at
		`).
		Joins("JOIN dashboard_metrics dm ON dm.id = dmm.dashboard_metric_id").
		Joins("JOIN forms f ON f.id = dmm.form_id").
		Joins("JOIN form_fields ff ON ff.id = dmm.form_field_id")

	if formCode != "" {
		db = db.Where("f.code = ?", formCode)
	}
	if dashboardMetricID > 0 {
		db = db.Where("dmm.dashboard_metric_id = ?", dashboardMetricID)
	}

	if req.Search != "" {
		searchTerm := "%" + req.Search + "%"
		db = db.Where(`
			dm.metric_key ILIKE ? OR
			dm.label ILIKE ? OR
			f.title ILIKE ? OR
			ff.question ILIKE ?
		`, searchTerm, searchTerm, searchTerm, searchTerm)
	}

	if err := db.Count(&totalData).Error; err != nil {
		return nil, 0, err
	}

	finalOrderBy := "dmm.id"
	finalOrderDir := "desc"
	allowedOrderCols := map[string]string{
		"id":         "dmm.id",
		"metric_key": "dm.metric_key",
		"form_title": "f.title",
		"created_at": "dmm.created_at",
		"updated_at": "dmm.updated_at",
	}
	if mappedCol, ok := allowedOrderCols[req.OrderBy]; ok {
		finalOrderBy = mappedCol
	}
	if strings.ToLower(req.OrderDir) == "asc" {
		finalOrderDir = "asc"
	}
	db = db.Order(finalOrderBy + " " + finalOrderDir)

	if req.Page <= 0 {
		req.Page = 1
	}
	if req.Limit <= 0 {
		req.Limit = 25
	}
	offset := (req.Page - 1) * req.Limit

	if err := db.Limit(req.Limit).Offset(offset).Scan(&data).Error; err != nil {
		return nil, 0, err
	}

	return data, totalData, nil
}

func (r *dashboardMetricMappingRepository) FindDetailByID(ctx context.Context, id int64) (*models.DashboardMetricMapping, error) {
	var m models.DashboardMetricMapping
	err := r.dbSlave.WithContext(ctx).
		Preload("DashboardMetric").
		Preload("Form").
		Preload("FormField").
		First(&m, id).Error
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *dashboardMetricMappingRepository) CountOptionStatusByMappingID(ctx context.Context, mappingID int64) (int64, error) {
	var count int64
	err := r.dbSlave.WithContext(ctx).
		Table("dashboard_option_status_mappings").
		Where("dashboard_metric_mapping_id = ?", mappingID).
		Count(&count).Error
	return count, err
}

func (r *dashboardMetricMappingRepository) CountByMetricID(ctx context.Context, metricID int64) (int64, error) {
	var count int64
	err := r.dbSlave.WithContext(ctx).
		Model(&models.DashboardMetricMapping{}).
		Where("dashboard_metric_id = ?", metricID).
		Count(&count).Error
	return count, err
}
