package repository

import (
	"backend/masterapi/models"
	"context"
	"errors"

	"gorm.io/gorm"
)

type DashboardOptionStatusMappingRepository interface {
	Create(ctx context.Context, m *models.DashboardOptionStatusMapping) error
	BulkCreate(ctx context.Context, list []models.DashboardOptionStatusMapping) error
	Update(ctx context.Context, m *models.DashboardOptionStatusMapping) error
	Delete(ctx context.Context, id int64) error
	FindByID(ctx context.Context, id int64) (*models.DashboardOptionStatusMapping, error)
	FindByMappingAndOption(ctx context.Context, mappingID, answerOptionID int64) (*models.DashboardOptionStatusMapping, error)
	// ListByMappingID dipakai endpoint dashboard runtime: dari 1 dashboard_metric_mapping_id,
	// ambil semua status_key opsi yang sudah di-assign, buat resolve badge saat query field_responses
	ListByMappingID(ctx context.Context, mappingID int64) ([]models.DashboardOptionStatusMapping, error)
}

type dashboardOptionStatusMappingRepository struct {
	dbMaster *gorm.DB
	dbSlave  *gorm.DB
}

func NewDashboardOptionStatusMappingRepository(dbMaster, dbSlave *gorm.DB) DashboardOptionStatusMappingRepository {
	return &dashboardOptionStatusMappingRepository{dbMaster: dbMaster, dbSlave: dbSlave}
}

func (r *dashboardOptionStatusMappingRepository) Create(ctx context.Context, m *models.DashboardOptionStatusMapping) error {
	return r.dbMaster.WithContext(ctx).Create(m).Error
}

func (r *dashboardOptionStatusMappingRepository) BulkCreate(ctx context.Context, list []models.DashboardOptionStatusMapping) error {
	if len(list) == 0 {
		return nil
	}
	return r.dbMaster.WithContext(ctx).Create(&list).Error
}

func (r *dashboardOptionStatusMappingRepository) Update(ctx context.Context, m *models.DashboardOptionStatusMapping) error {
	return r.dbMaster.WithContext(ctx).Save(m).Error
}

func (r *dashboardOptionStatusMappingRepository) Delete(ctx context.Context, id int64) error {
	return r.dbMaster.WithContext(ctx).Delete(&models.DashboardOptionStatusMapping{}, id).Error
}

func (r *dashboardOptionStatusMappingRepository) FindByID(ctx context.Context, id int64) (*models.DashboardOptionStatusMapping, error) {
	var m models.DashboardOptionStatusMapping
	err := r.dbSlave.WithContext(ctx).First(&m, id).Error
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *dashboardOptionStatusMappingRepository) FindByMappingAndOption(ctx context.Context, mappingID, answerOptionID int64) (*models.DashboardOptionStatusMapping, error) {
	var m models.DashboardOptionStatusMapping
	err := r.dbSlave.WithContext(ctx).
		Where("dashboard_metric_mapping_id = ? AND answer_option_id = ?", mappingID, answerOptionID).
		First(&m).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *dashboardOptionStatusMappingRepository) ListByMappingID(ctx context.Context, mappingID int64) ([]models.DashboardOptionStatusMapping, error) {
	var list []models.DashboardOptionStatusMapping
	err := r.dbSlave.WithContext(ctx).
		Where("dashboard_metric_mapping_id = ?", mappingID).
		Find(&list).Error
	return list, err
}
