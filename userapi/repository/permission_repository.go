package repository

import (
	"backend/userapi/models"
	"backend/userapi/utils"

	"gorm.io/gorm"
)

type PermissionRepository interface {
	GetMenuPermission(roleID int) ([]models.MenuPermission, error)
}

type permissionRepository struct {
	dbSlave  *gorm.DB
	dbMaster *gorm.DB
}

func NewPermissionRepository(dbSlave, dbMaster *gorm.DB) *permissionRepository {
	defer utils.GeneralRecover()
	return &permissionRepository{
		dbSlave,
		dbMaster,
	}
}

func (r *permissionRepository) GetMenuPermission(roleID int) ([]models.MenuPermission, error) {
	var menuPerms []models.MenuPermission
	err := r.dbSlave.
		Preload("Menu").
		Joins("JOIN menus ON menus.id = menu_permissions.menu_id").
		Where("menu_permissions.role_id = ?", roleID).
		Order("menus.sort_order ASC").
		Find(&menuPerms).Error
	if err != nil {
		return nil, err
	}
	return menuPerms, nil
}
