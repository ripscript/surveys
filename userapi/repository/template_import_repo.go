package repository

import (
	"context"

	"gorm.io/gorm"
)

type TemplateRoleRow struct {
	ID   int64  `gorm:"column:id"`
	Name string `gorm:"column:name"`
}

type TemplateKecamatanRow struct {
	ID   int64  `gorm:"column:id"`
	Name string `gorm:"column:name"`
}

type TemplateKelurahanRow struct {
	ID          int64  `gorm:"column:id"`
	Name        string `gorm:"column:name"`
	KecamatanID int64  `gorm:"column:kecamatan_id"`
}

type TemplateRwRow struct {
	ID          int64  `gorm:"column:id"`
	Name        string `gorm:"column:name"`
	KelurahanID int64  `gorm:"column:kelurahan_id"`
}

type TemplateRtRow struct {
	ID   int64  `gorm:"column:id"`
	Name string `gorm:"column:name"`
	RwID int64  `gorm:"column:rw_id"`
}

type TemplateImportRepo interface {
	GetRolesForTemplate(ctx context.Context) ([]TemplateRoleRow, error)
	GetKecamatansForTemplate(ctx context.Context) ([]TemplateKecamatanRow, error)
	GetKelurahansForTemplate(ctx context.Context) ([]TemplateKelurahanRow, error)
	GetRwsForTemplate(ctx context.Context) ([]TemplateRwRow, error)
	GetRtsForTemplate(ctx context.Context) ([]TemplateRtRow, error)
}

type templateImportRepo struct {
	dbSlave *gorm.DB
}

func NewTemplateImportRepo(dbSlave *gorm.DB) TemplateImportRepo {
	return &templateImportRepo{dbSlave: dbSlave}
}

func (r *templateImportRepo) GetRolesForTemplate(ctx context.Context) ([]TemplateRoleRow, error) {
	var rows []TemplateRoleRow
	err := r.dbSlave.WithContext(ctx).
		Table("role").
		Select("id, name").
		Order("id ASC").
		Scan(&rows).Error
	return rows, err
}

func (r *templateImportRepo) GetKecamatansForTemplate(ctx context.Context) ([]TemplateKecamatanRow, error) {
	var rows []TemplateKecamatanRow
	err := r.dbSlave.WithContext(ctx).
		Table("kecamatans").
		Select("id, sub_district_name AS name").
		Where("deleted_at IS NULL").
		Order("sub_district_name ASC").
		Scan(&rows).Error
	return rows, err
}

func (r *templateImportRepo) GetKelurahansForTemplate(ctx context.Context) ([]TemplateKelurahanRow, error) {
	var rows []TemplateKelurahanRow
	err := r.dbSlave.WithContext(ctx).
		Table("kelurahans").
		Select("id, village_name AS name, sub_district_id AS kecamatan_id").
		Where("deleted_at IS NULL").
		Order("village_name ASC").
		Scan(&rows).Error
	return rows, err
}

func (r *templateImportRepo) GetRwsForTemplate(ctx context.Context) ([]TemplateRwRow, error) {
	var rows []TemplateRwRow
	err := r.dbSlave.WithContext(ctx).
		Table("data__rws").
		Select("id, nama_rw AS name, kelurahan_id").
		Where("deleted_at IS NULL").
		Order("kelurahan_id ASC, nama_rw ASC").
		Scan(&rows).Error
	return rows, err
}

func (r *templateImportRepo) GetRtsForTemplate(ctx context.Context) ([]TemplateRtRow, error) {
	var rows []TemplateRtRow
	err := r.dbSlave.WithContext(ctx).
		Table("data__rts").
		Select("id, nama_rt AS name, rw_id").
		Where("deleted_at IS NULL").
		Order("rw_id ASC, nama_rt ASC").
		Scan(&rows).Error
	return rows, err
}
