package repository

import (
	"backend/masterapi/models"
	"backend/masterapi/payloads"
	"backend/masterapi/utils"
	"context"
	"errors"
	"strings"

	"gorm.io/gorm"
)

type DashboardMetricRepository interface {
	GetList(req payloads.DatatablePayload) ([]models.DashboardMetric, int64, error)
	Create(ctx context.Context, m *models.DashboardMetric) error
	Update(ctx context.Context, m *models.DashboardMetric) error
	Delete(ctx context.Context, id int64) error
	FindByID(ctx context.Context, id int64) (*models.DashboardMetric, error)
	FindByMetricKey(ctx context.Context, key string) (*models.DashboardMetric, error)
	List(ctx context.Context, category string) ([]models.DashboardMetric, error)

	GetCategoryOptions(ctx context.Context, search string, categories []string, page, limit int) ([]string, int64, error)
}

type dashboardMetricRepository struct {
	dbMaster *gorm.DB
	dbSlave  *gorm.DB
}

func NewDashboardMetricRepository(dbMaster, dbSlave *gorm.DB) DashboardMetricRepository {
	return &dashboardMetricRepository{dbMaster: dbMaster, dbSlave: dbSlave}
}

func (repository *dashboardMetricRepository) GetList(req payloads.DatatablePayload) ([]models.DashboardMetric, int64, error) {
	defer utils.GeneralRecover()
	var data []models.DashboardMetric
	var totalData int64

	db := repository.dbSlave.Table("dashboard_metrics").
		Select(`
			dashboard_metrics.id,
			dashboard_metrics.metric_key,
			dashboard_metrics.label,
			dashboard_metrics.expected_template,
			dashboard_metrics.category,
			dashboard_metrics.created_at,
			dashboard_metrics.updated_at,
			dashboard_metrics.is_dashboard
		`)

	if req.Search != "" {
		searchTerm := "%" + req.Search + "%"
		searchStr := strings.TrimSpace(req.Search)
		parsedDate, isDate := utils.TryParseIndonesianDate(req.Search)

		if isDate {
			db = db.Where(`
				dashboard_metrics.metric_key ILIKE ? OR 
				dashboard_metrics.label ILIKE ? OR 
				dashboard_metrics.expected_template ILIKE ? OR 
				dashboard_metrics.category ILIKE ? OR 
				DATE(dashboard_metrics.created_at) = ? OR 
				DATE(dashboard_metrics.updated_at) = ?
			`, searchTerm, searchTerm, searchTerm, searchTerm, parsedDate, parsedDate)
		} else if len(searchStr) == 4 {
			db = db.Where(`
				dashboard_metrics.metric_key ILIKE ? OR 
				dashboard_metrics.label ILIKE ? OR 
				dashboard_metrics.expected_template ILIKE ? OR 
				dashboard_metrics.category ILIKE ? OR 
				EXTRACT(YEAR FROM artikel_categories.created_at)::TEXT = ? OR 
				EXTRACT(YEAR FROM artikel_categories.updated_at)::TEXT = ?
			`, searchTerm, searchTerm, searchTerm, searchTerm, searchStr, searchStr)
		} else {
			db = db.Where(`
				dashboard_metrics.metric_key ILIKE ? OR 
				dashboard_metrics.label ILIKE ? OR 
				dashboard_metrics.expected_template ILIKE ? OR 
				dashboard_metrics.category ILIKE ?
			`, searchTerm, searchTerm, searchTerm, searchTerm)
		}
	}

	err := db.Count(&totalData).Error
	if err != nil {
		return nil, 0, err
	}

	if req.OrderBy != "" {
		finalOrderBy := "dashboard_metrics.id"
		finalOrderDir := "desc"

		allowedOrderCols := map[string]string{
			"id":                "dashboard_metrics.id",
			"metric_key":        "dashboard_metrics.metric_key",
			"label":             "dashboard_metrics.label",
			"expected_template": "dashboard_metrics.expected_template",
			"category":          "dashboard_metrics.category",
			"created_at":        "dashboard_metrics.created_at",
			"updated_at":        "dashboard_metrics.updated_at",
		}

		if mappedCol, isAllowed := allowedOrderCols[req.OrderBy]; isAllowed && req.OrderBy != "" {
			finalOrderBy = mappedCol
		}

		if strings.ToLower(req.OrderDir) == "asc" {
			finalOrderDir = "asc"
		}

		db = db.Order(finalOrderBy + " " + finalOrderDir)
	} else {
		db = db.Order("dashboard_metrics.id desc")
	}

	// Fitur Pagination
	offset := (req.Page - 1) * req.Limit
	err = db.Limit(req.Limit).Offset(offset).Find(&data).Error
	if err != nil {
		return nil, 0, err
	}

	return data, totalData, nil
}

func (r *dashboardMetricRepository) Create(ctx context.Context, m *models.DashboardMetric) error {
	return r.dbMaster.WithContext(ctx).Create(m).Error
}

func (r *dashboardMetricRepository) Update(ctx context.Context, m *models.DashboardMetric) error {
	return r.dbMaster.WithContext(ctx).Save(m).Error
}

func (r *dashboardMetricRepository) Delete(ctx context.Context, id int64) error {
	return r.dbMaster.WithContext(ctx).Delete(&models.DashboardMetric{}, id).Error
}

func (r *dashboardMetricRepository) FindByID(ctx context.Context, id int64) (*models.DashboardMetric, error) {
	var m models.DashboardMetric
	err := r.dbSlave.WithContext(ctx).First(&m, id).Error
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *dashboardMetricRepository) FindByMetricKey(ctx context.Context, key string) (*models.DashboardMetric, error) {
	var m models.DashboardMetric
	err := r.dbSlave.WithContext(ctx).Where("metric_key = ?", key).First(&m).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *dashboardMetricRepository) List(ctx context.Context, category string) ([]models.DashboardMetric, error) {
	var list []models.DashboardMetric
	q := r.dbSlave.WithContext(ctx)
	if category != "" {
		q = q.Where("category = ?", category)
	}
	err := q.Order("category, id").Find(&list).Error
	return list, err
}

func (r *dashboardMetricRepository) GetCategoryOptions(ctx context.Context, search string, categories []string, page, limit int) ([]string, int64, error) {
	defer utils.GeneralRecover()

	base := r.dbSlave.WithContext(ctx).
		Table("dashboard_metrics").
		Select("DISTINCT category").
		Where("category IS NOT NULL AND category != ''")

	if search != "" {
		base = base.Where("category ILIKE ?", "%"+search+"%")
	}
	if len(categories) > 0 {
		base = base.Where("category IN ?", categories)
	}

	// hitung total distinct category dulu (pakai subquery supaya COUNT tidak kena pengaruh DISTINCT+pagination)
	var totalData int64
	countQuery := r.dbSlave.WithContext(ctx).
		Table("(?) as sub", base).
		Count(&totalData)
	if countQuery.Error != nil {
		return nil, 0, countQuery.Error
	}

	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 25
	}
	offset := (page - 1) * limit

	var results []string
	err := base.Order("category asc").Limit(limit).Offset(offset).Pluck("category", &results).Error
	if err != nil {
		return nil, 0, err
	}

	return results, totalData, nil
}
