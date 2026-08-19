package repository

import (
	"backend/masterapi/models"
	"backend/masterapi/utils"
	"context"
	"strings"

	"gorm.io/gorm"
)

type ArtikelRepository interface {
	Create(ctx context.Context, artikel *models.Artikel) (*models.Artikel, error)
	GetByName(ctx context.Context, name string) (*models.Artikel, error)
}

type artikelRepository struct {
	dbSlave  *gorm.DB
	dbMaster *gorm.DB
}

func NewArtikelRepository(dbSlave, dbMaster *gorm.DB) ArtikelRepository {
	return &artikelRepository{dbSlave: dbSlave, dbMaster: dbMaster}
}

func (repository *artikelRepository) Create(ctx context.Context, artikel *models.Artikel) (*models.Artikel, error) {
	defer utils.GeneralRecover()
	err := repository.dbMaster.WithContext(ctx).Create(artikel).Error
	if err != nil {
		return nil, err
	}
	return artikel, nil
}

func (repository *artikelRepository) GetByName(ctx context.Context, name string) (*models.Artikel, error) {
	defer utils.GeneralRecover()
	var category models.Artikel
	err := repository.dbSlave.WithContext(ctx).Where("LOWER(judul) = ?", strings.ToLower(name)).First(&category).Error
	if err != nil {
		return nil, err
	}
	return &category, nil
}
