package repository

import (
	"backend/masterapi/models"
	"context"

	"gorm.io/gorm"
)

type DashboardMetricStatusRepository interface {
	Create(db *gorm.DB, ctx context.Context, m *models.DashboardMetricStatus) error
	BulkCreate(db *gorm.DB, ctx context.Context, list []models.DashboardMetricStatus) error
	Update(db *gorm.DB, ctx context.Context, m *models.DashboardMetricStatus) error
	Delete(db *gorm.DB, ctx context.Context, id int64) error
	Deactivate(db *gorm.DB, ctx context.Context, id int64) error

	FindByID(ctx context.Context, id int64) (*models.DashboardMetricStatus, error)
	FindByMetricID(ctx context.Context, metricID int64) ([]models.DashboardMetricStatus, error)
	CountUsage(ctx context.Context, statusID int64) (int64, error)
}

type dashboardMetricStatusRepository struct {
	dbSlave *gorm.DB
}

func NewDashboardMetricStatusRepository(dbSlave *gorm.DB) DashboardMetricStatusRepository {
	return &dashboardMetricStatusRepository{dbSlave: dbSlave}
}

func (r *dashboardMetricStatusRepository) Create(db *gorm.DB, ctx context.Context, m *models.DashboardMetricStatus) error {
	return db.WithContext(ctx).Create(m).Error
}

func (r *dashboardMetricStatusRepository) BulkCreate(db *gorm.DB, ctx context.Context, list []models.DashboardMetricStatus) error {
	if len(list) == 0 {
		return nil
	}
	return db.WithContext(ctx).Create(&list).Error
}

func (r *dashboardMetricStatusRepository) Update(db *gorm.DB, ctx context.Context, m *models.DashboardMetricStatus) error {
	return db.WithContext(ctx).Save(m).Error
}

func (r *dashboardMetricStatusRepository) Delete(db *gorm.DB, ctx context.Context, id int64) error {
	return db.WithContext(ctx).Delete(&models.DashboardMetricStatus{}, id).Error
}

func (r *dashboardMetricStatusRepository) Deactivate(db *gorm.DB, ctx context.Context, id int64) error {
	return db.WithContext(ctx).
		Model(&models.DashboardMetricStatus{}).
		Where("id = ?", id).
		Update("is_active", false).Error
}

func (r *dashboardMetricStatusRepository) FindByID(ctx context.Context, id int64) (*models.DashboardMetricStatus, error) {
	var m models.DashboardMetricStatus
	err := r.dbSlave.WithContext(ctx).First(&m, id).Error
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *dashboardMetricStatusRepository) FindByMetricID(ctx context.Context, metricID int64) ([]models.DashboardMetricStatus, error) {
	var list []models.DashboardMetricStatus
	err := r.dbSlave.WithContext(ctx).
		Where("dashboard_metric_id = ?", metricID).
		Order("sequence asc, id asc").
		Find(&list).Error
	return list, err
}

func (r *dashboardMetricStatusRepository) CountUsage(ctx context.Context, statusID int64) (int64, error) {
	var count int64
	err := r.dbSlave.WithContext(ctx).
		Table("dashboard_option_status_mappings").
		Where("dashboard_metric_status_id = ?", statusID).
		Count(&count).Error
	return count, err
}
