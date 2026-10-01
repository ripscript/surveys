package repository

import (
	"backend/masterapi/models"
	"backend/masterapi/payloads"
	"backend/masterapi/response"
	"backend/masterapi/utils"
	"context"
	"errors"
	"strings"

	"gorm.io/gorm"
)

type DashboardMetricRepository interface {
	GetList(ctx context.Context, req payloads.DatatablePayload) ([]models.DashboardMetric, int64, error)
	Create(db *gorm.DB, ctx context.Context, m *models.DashboardMetric) error
	Update(db *gorm.DB, ctx context.Context, m *models.DashboardMetric) error
	Delete(db *gorm.DB, ctx context.Context, id int64) error
	FindByID(ctx context.Context, id int64) (*models.DashboardMetric, error)
	FindDetailByID(ctx context.Context, id int64) (*models.DashboardMetric, error) // preload Statuses
	FindByMetricKey(ctx context.Context, key string) (*models.DashboardMetric, error)
	List(ctx context.Context, category string) ([]models.DashboardMetric, error)
	CountMappingsByMetricID(ctx context.Context, metricID int64) (int64, error)
	GetCategoryOptions(ctx context.Context, search string, categories []string, page, limit int) ([]string, int64, error)

	GetMetricOptions(ctx context.Context, req payloads.DashboardMetricOptionsPayload) ([]response.OptionItem, int64, error)
	GetPemetaanMetrikSurveyList(req payloads.PemetaanMetrikSurveyDatatablePayload) ([]response.PemetaanMetrikSurveyDatatableResponse, int64, error)
}

type dashboardMetricRepository struct {
	dbMaster *gorm.DB
	dbSlave  *gorm.DB
}

func NewDashboardMetricRepository(dbMaster, dbSlave *gorm.DB) DashboardMetricRepository {
	return &dashboardMetricRepository{dbMaster: dbMaster, dbSlave: dbSlave}
}

func (r *dashboardMetricRepository) GetList(ctx context.Context, req payloads.DatatablePayload) ([]models.DashboardMetric, int64, error) {
	defer utils.GeneralRecover()
	var data []models.DashboardMetric
	var totalData int64

	db := r.dbSlave.WithContext(ctx).Model(&models.DashboardMetric{})

	if req.Search != "" {
		searchTerm := "%" + req.Search + "%"
		parsedDate, isDate := utils.TryParseIndonesianDate(req.Search)
		searchStr := strings.TrimSpace(req.Search)

		switch {
		case isDate:
			db = db.Where(`
				metric_key ILIKE ? OR label ILIKE ? OR expected_template ILIKE ? OR category ILIKE ? OR
				DATE(created_at) = ? OR DATE(updated_at) = ?
			`, searchTerm, searchTerm, searchTerm, searchTerm, parsedDate, parsedDate)
		case len(searchStr) == 4:
			db = db.Where(`
				metric_key ILIKE ? OR label ILIKE ? OR expected_template ILIKE ? OR category ILIKE ? OR
				EXTRACT(YEAR FROM created_at)::TEXT = ? OR EXTRACT(YEAR FROM updated_at)::TEXT = ?
			`, searchTerm, searchTerm, searchTerm, searchTerm, searchStr, searchStr)
		default:
			db = db.Where(`
				metric_key ILIKE ? OR label ILIKE ? OR expected_template ILIKE ? OR category ILIKE ?
			`, searchTerm, searchTerm, searchTerm, searchTerm)
		}
	}

	if err := db.Count(&totalData).Error; err != nil {
		return nil, 0, err
	}

	finalOrderBy := "id"
	finalOrderDir := "desc"
	allowedOrderCols := map[string]string{
		"id": "id", "metric_key": "metric_key", "label": "label",
		"expected_template": "expected_template", "category": "category",
		"created_at": "created_at", "updated_at": "updated_at",
	}
	if col, ok := allowedOrderCols[req.OrderBy]; ok {
		finalOrderBy = col
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

	if err := db.Limit(req.Limit).Offset(offset).Find(&data).Error; err != nil {
		return nil, 0, err
	}
	return data, totalData, nil
}

func (r *dashboardMetricRepository) Create(db *gorm.DB, ctx context.Context, m *models.DashboardMetric) error {
	return db.WithContext(ctx).Create(m).Error
}

func (r *dashboardMetricRepository) Update(db *gorm.DB, ctx context.Context, m *models.DashboardMetric) error {
	return db.WithContext(ctx).Save(m).Error
}

func (r *dashboardMetricRepository) Delete(db *gorm.DB, ctx context.Context, id int64) error {
	return db.WithContext(ctx).Delete(&models.DashboardMetric{}, id).Error
}

func (r *dashboardMetricRepository) FindByID(ctx context.Context, id int64) (*models.DashboardMetric, error) {
	var m models.DashboardMetric
	err := r.dbSlave.WithContext(ctx).First(&m, id).Error
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *dashboardMetricRepository) FindDetailByID(ctx context.Context, id int64) (*models.DashboardMetric, error) {
	var m models.DashboardMetric
	err := r.dbSlave.WithContext(ctx).
		Preload("Statuses", func(db *gorm.DB) *gorm.DB {
			return db.Order("sequence asc, id asc")
		}).
		First(&m, id).Error
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

func (r *dashboardMetricRepository) CountMappingsByMetricID(ctx context.Context, metricID int64) (int64, error) {
	var count int64
	err := r.dbSlave.WithContext(ctx).
		Table("dashboard_metric_mappings").
		Where("dashboard_metric_id = ?", metricID).
		Count(&count).Error
	return count, err
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

	var totalData int64
	if err := r.dbSlave.WithContext(ctx).Table("(?) as sub", base).Count(&totalData).Error; err != nil {
		return nil, 0, err
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
	return results, totalData, err
}

func (r *dashboardMetricRepository) GetMetricOptions(ctx context.Context, req payloads.DashboardMetricOptionsPayload) ([]response.OptionItem, int64, error) {
	defer utils.GeneralRecover()

	var data []response.OptionItem
	var totalData int64

	applyTemplateFilter := req.ExpectedTemplate != ""
	searchTerm := "%" + req.Q + "%"

	buildFilter := func(tx *gorm.DB) *gorm.DB {
		if applyTemplateFilter {
			tx = tx.Where("expected_template = ?", req.ExpectedTemplate)
		}
		if req.Q != "" {
			tx = tx.Where("metric_key ILIKE ? OR label ILIKE ?", searchTerm, searchTerm)
		}
		return tx
	}

	db := r.dbSlave.WithContext(ctx).
		Table("dashboard_metrics").
		Select(`
			dashboard_metrics.id AS id,
			CONCAT(dashboard_metrics.label, ' (', dashboard_metrics.metric_key, ')') AS label
		`)

	if len(req.IDs) > 0 {
		db = db.Where(
			buildFilter(r.dbSlave.Session(&gorm.Session{NewDB: true})),
		).Or("dashboard_metrics.id IN ?", req.IDs)
	} else {
		db = buildFilter(db)
	}

	if err := db.Count(&totalData).Error; err != nil {
		return nil, 0, err
	}

	db = db.Order("dashboard_metrics.label ASC")

	limit := req.Limit
	if limit <= 0 {
		limit = 1000
	}
	page := req.Page
	if page <= 0 {
		page = 1
	}
	offset := (page - 1) * limit

	if err := db.Limit(limit).Offset(offset).Find(&data).Error; err != nil {
		return nil, 0, err
	}

	return data, totalData, nil
}

func (repository *dashboardMetricRepository) GetPemetaanMetrikSurveyList(req payloads.PemetaanMetrikSurveyDatatablePayload) ([]response.PemetaanMetrikSurveyDatatableResponse, int64, error) {
	defer utils.GeneralRecover()
	var data []response.PemetaanMetrikSurveyDatatableResponse
	var totalData int64

	totalFieldsSub := repository.dbSlave.
		Table("flow_fields ff").
		Select("ff.flow_detail_id, COUNT(DISTINCT ff.form_field_id) AS total_fields").
		Group("ff.flow_detail_id")

	mappedFieldsSub := repository.dbSlave.
		Table("flow_fields ff").
		Select("ff.flow_detail_id, COUNT(DISTINCT ff.form_field_id) AS mapped_fields").
		Joins("JOIN dashboard_metric_mappings dmm ON dmm.form_field_id = ff.form_field_id").
		Group("ff.flow_detail_id")

	db := repository.dbSlave.Table("surveys").
		Joins("LEFT JOIN flow_details ON flow_details.id = surveys.flow_detail_id").
		Joins("LEFT JOIN (?) AS ft ON ft.flow_detail_id = flow_details.id", totalFieldsSub).
		Joins("LEFT JOIN (?) AS fm ON fm.flow_detail_id = flow_details.id", mappedFieldsSub)

	if req.Status != "" {
		db = db.Where("surveys.status = ?", req.Status)
	}

	if req.Search != "" {
		searchTerm := "%" + req.Search + "%"
		db = db.Where(`
			surveys.name ILIKE ? OR
			flow_details.name ILIKE ? OR
			flow_details.version::text ILIKE ?
		`, searchTerm, searchTerm, searchTerm)
	}

	err := db.Count(&totalData).Error
	if err != nil {
		return nil, 0, err
	}

	db = db.Select(`
		surveys.id,
		surveys.name AS survey_name,
		flow_details.name AS flow_name,
		flow_details.version AS flow_version,
		surveys.status,
		COALESCE(ft.total_fields, 0) > 0
			AND COALESCE(ft.total_fields, 0) = COALESCE(fm.mapped_fields, 0) AS mapping_status
	`)

	if req.OrderBy != "" {
		finalOrderBy := "surveys.id"
		finalOrderDir := "desc"

		allowedOrderCols := map[string]string{
			"id":             "surveys.id",
			"survey_name":    "surveys.name",
			"flow_name":      "flow_details.name",
			"flow_version":   "flow_details.version",
			"status":         "surveys.status",
			"mapping_status": "mapping_status",
		}

		if mappedCol, isAllowed := allowedOrderCols[req.OrderBy]; isAllowed {
			finalOrderBy = mappedCol
		}

		if strings.ToLower(req.OrderDir) == "asc" {
			finalOrderDir = "asc"
		}

		db = db.Order(finalOrderBy + " " + finalOrderDir)
	} else {
		db = db.Order("surveys.id desc")
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
