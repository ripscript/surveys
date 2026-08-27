package repository

import (
	"backend/masterapi/models"
	"context"

	"gorm.io/gorm"
)

type FormRepository interface {
	FindByCode(ctx context.Context, code string) (*models.Form, error)
}

type formRepository struct {
	dbSlave *gorm.DB
}

func NewFormRepository(dbSlave *gorm.DB) FormRepository {
	return &formRepository{dbSlave: dbSlave}
}

func (repository *formRepository) FindByCode(ctx context.Context, code string) (*models.Form, error) {
	var form models.Form
	err := repository.dbSlave.WithContext(ctx).Where("code = ?", code).First(&form).Error
	if err != nil {
		return nil, err
	}
	return &form, nil
}
