package repository

import (
	"backend/userapi/models"
	"backend/userapi/payloads"
	"backend/userapi/utils"
	"errors"
	"strconv"
	"time"

	"gorm.io/gorm"
)

type PenggunaRepo interface {
	FindRespondentByRole(payload payloads.LoginPayload) (*models.Respondent, string, error)
	FindUserByRespondentID(id int) (*models.User, error)
	CountFailedLogin(userID int) (int64, error)
	InsertFailedLogin(userID int) error
	BlockRespondent(id int) error
	CheckActiveJabatan(id int) (bool, error)
	FindUserByID(id int) (*models.User, error)
	GetMenuPermission(roleId int) ([]models.MenuPermission, error)
	ListMenus(RoleId int64) ([]models.MenuPermissionRole, error)
	GetRespondentHasJabatanActive(payload payloads.LoginPayload) (*models.Respondent, string, error)
}

type penggunaRepo struct {
	dbSlave  *gorm.DB
	dbMaster *gorm.DB
}

func NewPenggunaRepo(dbSlave, dbMaster *gorm.DB) *penggunaRepo {
	defer utils.GeneralRecover()
	return &penggunaRepo{
		dbSlave,
		dbMaster,
	}
}

func (r *penggunaRepo) ListMenus(roleId int64) ([]models.MenuPermissionRole, error) {
	defer utils.GeneralRecover()

	var menuPermissions []models.MenuPermission
	err := r.dbSlave.
		Preload("Menu").
		Joins("JOIN menus ON menus.id = menu_permissions.menu_id").
		Where("menu_permissions.role_id = ?", roleId).
		Where("menu_permissions.view_action = ?", true).
		Where("menu_permissions.show_in_menu = ?", true).
		Order("menus.sort_order ASC").
		Find(&menuPermissions).Error
	if err != nil {
		return nil, err
	}

	childrenByParentID := make(map[int][]models.Menu)
	var roots []models.Menu
	for _, mp := range menuPermissions {
		if mp.Menu.ParentID == nil {
			roots = append(roots, mp.Menu)
		} else {
			childrenByParentID[*mp.Menu.ParentID] = append(childrenByParentID[*mp.Menu.ParentID], mp.Menu)
		}
	}

	var result []models.MenuPermissionRole
	for _, root := range roots {
		result = append(result, buildMenuTree(root, childrenByParentID))
	}

	return result, nil
}

func buildMenuTree(menu models.Menu, childrenByParentID map[int][]models.Menu) models.MenuPermissionRole {
	node := models.MenuPermissionRole{
		Key:   menu.Key,
		Title: menu.MenuName,
		Icon:  menu.Icon,
	}

	for _, child := range childrenByParentID[menu.ID] {
		childNode := buildMenuTree(child, childrenByParentID)
		node.ChildMenu = append(node.ChildMenu, models.ChildMenu{
			Key:       childNode.Key,
			Title:     childNode.Title,
			Icon:      childNode.Icon,
			ChildMenu: childNode.ChildMenu,
		})
	}

	return node
}

func (r *penggunaRepo) GetMenuPermission(roleId int) ([]models.MenuPermission, error) {
	var menuPerms []models.MenuPermission

	dbSlave := r.dbSlave
	err := dbSlave.Preload("Menu").Where("role_id = ?", roleId).Find(&menuPerms).Error
	if err != nil {
		return menuPerms, err
	}

	return menuPerms, nil
}

func (r *penggunaRepo) FindRespondentByRole(payload payloads.LoginPayload) (*models.Respondent, string, error) {
	var respondent models.Respondent
	var requestEmail string

	query := r.dbSlave.Where("role_id = ?", payload.Role).Where("deleted_at IS NULL")

	if payload.Role == 8 || payload.Role == 7 || payload.Role == 9 {
		query = query.Where("email = ?", payload.Email)
		requestEmail = payload.Email
	} else {
		if payload.SelectedKecamatan != nil {
			query = query.Where("kecamatan_id = ?", *payload.SelectedKecamatan)
		}
		if payload.SelectedKelurahan != nil {
			query = query.Where("kelurahan_id = ?", *payload.SelectedKelurahan)
		}
		if payload.SelectedRW != nil {
			query = query.Where("rw_id = ?", *payload.SelectedRW)
		}
		if payload.SelectedRT != nil {
			query = query.Where("rt_id = ?", *payload.SelectedRT)
		}
	}

	err := query.First(&respondent).Error
	if err != nil {
		return nil, "", err
	}

	if respondent.Email != "" {
		requestEmail = respondent.Email
	} else if respondent.Username != "" {
		requestEmail = respondent.Username
	} else {
		requestEmail = respondent.NIK
	}

	return &respondent, requestEmail, nil
}

var (
	ErrRespondentNotFound = errors.New("respondent not found")
	ErrJabatanNotActive   = errors.New("jabatan tidak aktif")

	ErrJabatanTidakAktif   = errors.New("jabatan tidak aktif")
	ErrPeriodeJabatanHabis = errors.New("periode jabatan telah berakhir")
)

func (r *penggunaRepo) GetRespondentHasJabatanActive(payload payloads.LoginPayload) (*models.Respondent, string, error) {
	var respondent models.Respondent
	var requestEmail string

	baseQuery := r.dbSlave.
		Where("respondents.role_id = ?", payload.Role).
		Where("respondents.deleted_at IS NULL")

	if payload.Role == 8 || payload.Role == 7 || payload.Role == 9 {
		baseQuery = baseQuery.Where("respondents.email = ?", payload.Email)
	} else {
		if payload.SelectedKecamatan != nil {
			baseQuery = baseQuery.Where("respondents.kecamatan_id = ?", *payload.SelectedKecamatan)
		}
		if payload.SelectedKelurahan != nil {
			baseQuery = baseQuery.Where("respondents.kelurahan_id = ?", *payload.SelectedKelurahan)
		}
		if payload.SelectedRW != nil {
			baseQuery = baseQuery.Where("respondents.rw_id = ?", *payload.SelectedRW)
		}
		if payload.SelectedRT != nil {
			baseQuery = baseQuery.Where("respondents.rt_id = ?", *payload.SelectedRT)
		}
	}

	err := baseQuery.First(&respondent).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, "", ErrRespondentNotFound
		}
		return nil, "", err
	}

	var activeRespondent models.Respondent
	jabatanQuery := r.dbSlave.
		Joins("JOIN pejabat__wilayahs pw ON pw.id_responden = respondents.id").
		Where("respondents.id = ?", respondent.ID).
		Where("pw.status_jabat = ?", 1).
		Where("pw.periode_akhir IS NULL OR pw.periode_akhir >= ?", time.Now())

	err = jabatanQuery.First(&activeRespondent).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, "", ErrJabatanNotActive
		}
		return nil, "", err
	}

	respondent = activeRespondent

	if respondent.Email != "" {
		requestEmail = respondent.Email
	} else if respondent.Username != "" {
		requestEmail = respondent.Username
	} else {
		requestEmail = respondent.NIK
	}

	return &respondent, requestEmail, nil
}

func (r *penggunaRepo) FindUserByRespondentID(id int) (*models.User, error) {
	var user models.User
	err := r.dbSlave.Where("respondent_id = ?", id).First(&user).Error
	return &user, err
}

func (r *penggunaRepo) CountFailedLogin(userID int) (int64, error) {
	var count int64

	err := r.dbSlave.Model(&models.LogBlockLogin{}).
		Where("user_credential = ?", strconv.Itoa(userID)).
		Count(&count).Error

	return count, err
}

func (r *penggunaRepo) InsertFailedLogin(userID int) error {
	var userIdStr = strconv.Itoa(userID)

	return r.dbMaster.Create(&models.LogBlockLogin{
		UserCredential: userIdStr,
	}).Error
}

func (r *penggunaRepo) BlockRespondent(id int) error {
	return r.dbMaster.Model(&models.Respondent{}).
		Where("id = ?", id).
		Update("is_blocked", true).Error
}

func (r *penggunaRepo) CheckActiveJabatan(id int) (bool, error) {
	var data models.PejabatWilayah

	err := r.dbSlave.Where("id_responden = ?", id).
		Where("status_jabat = ?", 1).
		Where("periode_akhir IS NULL OR periode_akhir >= ?", time.Now()).
		First(&data).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

func (r *penggunaRepo) FindUserByID(id int) (*models.User, error) {
	var user models.User
	err := r.dbSlave.Where("id = ?", id).First(&user).Error
	return &user, err
}
