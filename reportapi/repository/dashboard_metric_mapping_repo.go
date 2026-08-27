package repository

import (
	"backend/reportapi/models"
	"context"
	"errors"

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
}

type dashboardMetricMappingRepository struct {
	dbMaster *gorm.DB
	dbSlave  *gorm.DB
}

func NewDashboardMetricMappingRepository(dbMaster, dbSlave *gorm.DB) DashboardMetricMappingRepository {
	return &dashboardMetricMappingRepository{dbMaster: dbMaster, dbSlave: dbSlave}
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
