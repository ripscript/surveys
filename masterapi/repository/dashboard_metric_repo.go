package repository

import (
	"backend/masterapi/models"
	"context"
	"errors"

	"gorm.io/gorm"
)

type DashboardMetricRepository interface {
	Create(ctx context.Context, m *models.DashboardMetric) error
	Update(ctx context.Context, m *models.DashboardMetric) error
	Delete(ctx context.Context, id int64) error
	FindByID(ctx context.Context, id int64) (*models.DashboardMetric, error)
	FindByMetricKey(ctx context.Context, key string) (*models.DashboardMetric, error)
	List(ctx context.Context, category string) ([]models.DashboardMetric, error)
}

type dashboardMetricRepository struct {
	dbMaster *gorm.DB
	dbSlave  *gorm.DB
}

func NewDashboardMetricRepository(dbMaster, dbSlave *gorm.DB) DashboardMetricRepository {
	return &dashboardMetricRepository{dbMaster: dbMaster, dbSlave: dbSlave}
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
