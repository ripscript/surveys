package seeders

import (
	"backend/userapi/models"
	"backend/userapi/utils"
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func MenuSeed(db *gorm.DB) error {
	now := utils.TimeNow()

	// Create Parent Menus
	Menus := []models.Menu{
		{
			MenuName:  "Template",
			Key:       "template",
			Icon:      "-",
			CreatedBy: 0,
			UpdatedBy: 0,
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			MenuName:  "Manajemen Alur",
			Key:       "manajemen-alur",
			Icon:      "-",
			CreatedBy: 0,
			UpdatedBy: 0,
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			MenuName:  "Survey",
			Key:       "survey",
			Icon:      "-",
			CreatedBy: 0,
			UpdatedBy: 0,
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			MenuName:  "Dashboard",
			Key:       "dashboard",
			Icon:      "-",
			CreatedBy: 0,
			UpdatedBy: 0,
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			MenuName:  "Pengaturan",
			Key:       "pengaturan",
			Icon:      "-",
			CreatedBy: 0,
			UpdatedBy: 0,
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			MenuName:  "Laporan",
			Key:       "laporan",
			Icon:      "-",
			CreatedBy: 0,
			UpdatedBy: 0,
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			MenuName:  "Monitoring",
			Key:       "monitoring",
			Icon:      "-",
			CreatedBy: 0,
			UpdatedBy: 0,
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			MenuName:  "Admin",
			Key:       "admin",
			Icon:      "-",
			CreatedBy: 0,
			UpdatedBy: 0,
			CreatedAt: now,
			UpdatedAt: now,
		},
	}

	err := db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "key"}},
		DoNothing: true,
	}).Create(&Menus).Error

	if err != nil {
		return err
	}

	// Create Child Menus
	parentMenus := []models.Menu{}
	err = db.Where("key IN ?", []string{"template", "manajemen-alur", "survey", "dashboard", "pengaturan", "laporan", "monitoring", "admin"}).Find(&parentMenus).Error
	if err != nil {
		return err
	}
	menuIDMap := make(map[string]uint)
	for _, menu := range parentMenus {
		menuIDMap[menu.Key] = uint(menu.ID)
	}
	templateID := int(menuIDMap["template"])
	surveyID := int(menuIDMap["survey"])
	pengaturanID := int(menuIDMap["pengaturan"])
	monitoringID := int(menuIDMap["monitoring"])
	adminID := int(menuIDMap["admin"])
	ChildMenus := []models.Menu{
		{
			MenuName:  "Formulir Pertanyaan",
			Key:       "formulir-pertanyaan",
			ParentID:  &templateID,
			Icon:      "-",
			CreatedBy: 0,
			UpdatedBy: 0,
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			MenuName:  "Ucapan",
			Key:       "ucapan",
			ParentID:  &templateID,
			Icon:      "-",
			CreatedBy: 0,
			UpdatedBy: 0,
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			MenuName:  "List Survey",
			Key:       "list-survey",
			ParentID:  &surveyID,
			Icon:      "-",
			CreatedBy: 0,
			UpdatedBy: 0,
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			MenuName:  "Hasil",
			Key:       "hasil",
			ParentID:  &surveyID,
			Icon:      "-",
			CreatedBy: 0,
			UpdatedBy: 0,
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			MenuName:  "Hasil",
			Key:       "hasil",
			ParentID:  &surveyID,
			Icon:      "-",
			CreatedBy: 0,
			UpdatedBy: 0,
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			MenuName:  "Manajemen Pengguna",
			Key:       "manajemen-pengguna",
			ParentID:  &pengaturanID,
			Icon:      "-",
			CreatedBy: 0,
			UpdatedBy: 0,
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			MenuName:  "Manajemen Wilayah",
			Key:       "manajemen-wilayah",
			ParentID:  &pengaturanID,
			Icon:      "-",
			CreatedBy: 0,
			UpdatedBy: 0,
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			MenuName:  "Manajemen CMS",
			Key:       "manajemen-cms",
			ParentID:  &pengaturanID,
			Icon:      "-",
			CreatedBy: 0,
			UpdatedBy: 0,
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			MenuName:  "Manajemen Artikel",
			Key:       "manajemen-artikel",
			ParentID:  &pengaturanID,
			Icon:      "-",
			CreatedBy: 0,
			UpdatedBy: 0,
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			MenuName:  "Rating",
			Key:       "rating",
			ParentID:  &pengaturanID,
			Icon:      "-",
			CreatedBy: 0,
			UpdatedBy: 0,
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			MenuName:  "Statistik",
			Key:       "statistik",
			ParentID:  &monitoringID,
			Icon:      "-",
			CreatedBy: 0,
			UpdatedBy: 0,
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			MenuName:  "Aktifitas Survey",
			Key:       "aktifitas-survey",
			ParentID:  &monitoringID,
			Icon:      "-",
			CreatedBy: 0,
			UpdatedBy: 0,
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			MenuName:  "Profil Saya",
			Key:       "profil-saya",
			ParentID:  &adminID,
			Icon:      "-",
			CreatedBy: 0,
			UpdatedBy: 0,
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			MenuName:  "Keluar",
			Key:       "keluar",
			ParentID:  &adminID,
			Icon:      "-",
			CreatedBy: 0,
			UpdatedBy: 0,
			CreatedAt: now,
			UpdatedAt: now,
		},
	}

	err = db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "key"}},
		DoNothing: true,
	}).Create(&ChildMenus).Error

	if err != nil {
		return err
	}

	// Create Grandchild xixixi

	childMenus := []models.Menu{}
	err = db.Where("key IN ?", []string{"manajemen-pengguna", "manajemen-wilayah", "manajemen-artikel"}).Find(&childMenus).Error
	if err != nil {
		return err
	}
	childMenuIDMap := make(map[string]uint)
	for _, menu := range childMenus {
		childMenuIDMap[menu.Key] = uint(menu.ID)
	}
	manajemenPenggunaID := int(childMenuIDMap["manajemen-pengguna"])
	manajemenWilayahID := int(childMenuIDMap["manajemen-wilayah"])
	manajemenArtikelID := int(childMenuIDMap["manajemen-artikel"])

	GrindChildMenus := []models.Menu{
		{
			MenuName:  "Manajemen Responden",
			Key:       "manajemen-responden",
			ParentID:  &manajemenPenggunaID,
			Icon:      "-",
			CreatedBy: 0,
			UpdatedBy: 0,
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			MenuName:  "Manajemen User",
			Key:       "manajemen-user",
			ParentID:  &manajemenPenggunaID,
			Icon:      "-",
			CreatedBy: 0,
			UpdatedBy: 0,
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			MenuName:  "Manajemen Blokir",
			Key:       "manajemen-blokir",
			ParentID:  &manajemenPenggunaID,
			Icon:      "-",
			CreatedBy: 0,
			UpdatedBy: 0,
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			MenuName:  "Manajemen Wilayah",
			Key:       "manajemen-wilayah-child",
			ParentID:  &manajemenWilayahID,
			Icon:      "-",
			CreatedBy: 0,
			UpdatedBy: 0,
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			MenuName:  "Manajemen Pejabat",
			Key:       "manajemen-pejabat",
			ParentID:  &manajemenWilayahID,
			Icon:      "-",
			CreatedBy: 0,
			UpdatedBy: 0,
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			MenuName:  "Artikel",
			Key:       "artikel",
			ParentID:  &manajemenArtikelID,
			Icon:      "-",
			CreatedBy: 0,
			UpdatedBy: 0,
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			MenuName:  "Promote",
			Key:       "promote",
			ParentID:  &manajemenArtikelID,
			Icon:      "-",
			CreatedBy: 0,
			UpdatedBy: 0,
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			MenuName:  "Kategori",
			Key:       "kategori",
			ParentID:  &manajemenArtikelID,
			Icon:      "-",
			CreatedBy: 0,
			UpdatedBy: 0,
			CreatedAt: now,
			UpdatedAt: now,
		},
	}

	err = db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "key"}},
		DoNothing: true,
	}).Create(&GrindChildMenus).Error

	if err != nil {
		return err
	}

	return nil
}

func PermissionSeed(db *gorm.DB) error {
	var menuKeys = []string{"template", "manajemen-alur", "survey", "dashboard", "pengaturan", "laporan", "monitoring",
		"admin", "formulir-pertanyaan", "ucapan", "list-survey", "hasil", "manajemen-pengguna", "manajemen-wilayah",
		"manajemen-cms", "manajemen-artikel", "rating", "statistik", "aktifitas-survey", "profil-saya", "keluar",
		"manajemen-responden", "manajemen-user", "manajemen-blokir", "manajemen-wilayah-child", "manajemen-pejabat",
		"artikel", "promote", "kategori",
	}
	menuMap, err := GetMenuIDMap(db, menuKeys)
	if err != nil {
		return err
	}
	menuIDs := MapToStruct(menuMap)

	var roleName = []string{"public", "rt", "rw", "lurah", "camat", "pemkot", "admin", "surveyor", "walikota"}
	roleMap, err := GetRoleIDMap(db, roleName)
	if err != nil {
		return err
	}
	roleIDs := RoleMapToStruct(roleMap)

	// Mapping Menu ID By Key
	ManajemenUserID := menuIDs.ManajemenUser
	ManajemenWilayahID := menuIDs.ManajemenWilayahChild
	ManajemenPejabatID := menuIDs.ManajemenPejabat
	ManajemenArtikelID := menuIDs.ManajemenArtikel
	PromoteID := menuIDs.Promote
	CategoryID := menuIDs.Kategori
	ListSurveyID := menuIDs.ListSurvey
	HasilSurvey := menuIDs.Hasil
	ManajemenAlur := menuIDs.ManajemenAlur
	FormulirPertanyaan := menuIDs.FormulirPertanyaan
	Ucapan := menuIDs.Ucapan

	// Mapping Role Id By Name
	RolePublicId := roleIDs.Public
	RoleRtId := roleIDs.Rt
	RoleRwId := roleIDs.Rw
	RoleLurahId := roleIDs.Lurah
	RoleCamatId := roleIDs.Camat
	RolePemkotId := roleIDs.Pemkot
	RoleAdminId := roleIDs.Admin
	RoleSurveyorId := roleIDs.Surveyor
	RoleWalikotaId := roleIDs.Walikota

	MenuPermission := []models.MenuPermission{
		// Manajemen Ucapan
		{
			MenuID:       Ucapan,
			RoleID:       RolePublicId,
			ViewAction:   false,
			CreateAction: false,
			UpdateAction: false,
			DeleteAction: false,
		},
		{
			MenuID:       Ucapan,
			RoleID:       RoleRtId,
			ViewAction:   false,
			CreateAction: false,
			UpdateAction: false,
			DeleteAction: false,
		},
		{
			MenuID:       Ucapan,
			RoleID:       RoleRwId,
			ViewAction:   false,
			CreateAction: false,
			UpdateAction: false,
			DeleteAction: false,
		},
		{
			MenuID:       Ucapan,
			RoleID:       RoleLurahId,
			ViewAction:   false,
			CreateAction: false,
			UpdateAction: false,
			DeleteAction: false,
		},
		{
			MenuID:       Ucapan,
			RoleID:       RoleCamatId,
			ViewAction:   false,
			CreateAction: false,
			UpdateAction: false,
			DeleteAction: false,
		},
		{
			MenuID:       Ucapan,
			RoleID:       RolePemkotId,
			ViewAction:   false,
			CreateAction: false,
			UpdateAction: false,
			DeleteAction: false,
		},
		{
			MenuID:       Ucapan,
			RoleID:       RoleAdminId,
			ViewAction:   true,
			CreateAction: true,
			UpdateAction: true,
			DeleteAction: true,
		},
		{
			MenuID:       Ucapan,
			RoleID:       RoleSurveyorId,
			ViewAction:   false,
			CreateAction: false,
			UpdateAction: false,
			DeleteAction: false,
		},
		{
			MenuID:       Ucapan,
			RoleID:       RoleWalikotaId,
			ViewAction:   false,
			CreateAction: false,
			UpdateAction: false,
			DeleteAction: false,
		},
		// Manajemen Formulir Pertanyaan
		{
			MenuID:       FormulirPertanyaan,
			RoleID:       RolePublicId,
			ViewAction:   false,
			CreateAction: false,
			UpdateAction: false,
			DeleteAction: false,
		},
		{
			MenuID:       FormulirPertanyaan,
			RoleID:       RoleRtId,
			ViewAction:   false,
			CreateAction: false,
			UpdateAction: false,
			DeleteAction: false,
		},
		{
			MenuID:       FormulirPertanyaan,
			RoleID:       RoleRwId,
			ViewAction:   false,
			CreateAction: false,
			UpdateAction: false,
			DeleteAction: false,
		},
		{
			MenuID:       FormulirPertanyaan,
			RoleID:       RoleLurahId,
			ViewAction:   false,
			CreateAction: false,
			UpdateAction: false,
			DeleteAction: false,
		},
		{
			MenuID:       FormulirPertanyaan,
			RoleID:       RoleCamatId,
			ViewAction:   false,
			CreateAction: false,
			UpdateAction: false,
			DeleteAction: false,
		},
		{
			MenuID:       FormulirPertanyaan,
			RoleID:       RolePemkotId,
			ViewAction:   false,
			CreateAction: false,
			UpdateAction: false,
			DeleteAction: false,
		},
		{
			MenuID:       FormulirPertanyaan,
			RoleID:       RoleAdminId,
			ViewAction:   true,
			CreateAction: true,
			UpdateAction: true,
			DeleteAction: true,
		},
		{
			MenuID:       FormulirPertanyaan,
			RoleID:       RoleSurveyorId,
			ViewAction:   false,
			CreateAction: false,
			UpdateAction: false,
			DeleteAction: false,
		},
		{
			MenuID:       FormulirPertanyaan,
			RoleID:       RoleWalikotaId,
			ViewAction:   false,
			CreateAction: false,
			UpdateAction: false,
			DeleteAction: false,
		},
		// Manajemen Alur
		{
			MenuID:       ManajemenAlur,
			RoleID:       RolePublicId,
			ViewAction:   false,
			CreateAction: false,
			UpdateAction: false,
			DeleteAction: false,
		},
		{
			MenuID:       ManajemenAlur,
			RoleID:       RoleRtId,
			ViewAction:   false,
			CreateAction: false,
			UpdateAction: false,
			DeleteAction: false,
		},
		{
			MenuID:       ManajemenAlur,
			RoleID:       RoleRwId,
			ViewAction:   false,
			CreateAction: false,
			UpdateAction: false,
			DeleteAction: false,
		},
		{
			MenuID:       ManajemenAlur,
			RoleID:       RoleLurahId,
			ViewAction:   false,
			CreateAction: false,
			UpdateAction: false,
			DeleteAction: false,
		},
		{
			MenuID:       ManajemenAlur,
			RoleID:       RoleCamatId,
			ViewAction:   false,
			CreateAction: false,
			UpdateAction: false,
			DeleteAction: false,
		},
		{
			MenuID:       ManajemenAlur,
			RoleID:       RolePemkotId,
			ViewAction:   false,
			CreateAction: false,
			UpdateAction: false,
			DeleteAction: false,
		},
		{
			MenuID:       ManajemenAlur,
			RoleID:       RoleAdminId,
			ViewAction:   true,
			CreateAction: true,
			UpdateAction: true,
			DeleteAction: true,
		},
		{
			MenuID:       ManajemenAlur,
			RoleID:       RoleSurveyorId,
			ViewAction:   false,
			CreateAction: false,
			UpdateAction: false,
			DeleteAction: false,
		},
		{
			MenuID:       ManajemenAlur,
			RoleID:       RoleWalikotaId,
			ViewAction:   false,
			CreateAction: false,
			UpdateAction: false,
			DeleteAction: false,
		},
		// Manajemen Hasil Survey
		{
			MenuID:       HasilSurvey,
			RoleID:       RolePublicId,
			ViewAction:   false,
			CreateAction: false,
			UpdateAction: false,
			DeleteAction: false,
		},
		{
			MenuID:       HasilSurvey,
			RoleID:       RoleRtId,
			ViewAction:   false,
			CreateAction: false,
			UpdateAction: false,
			DeleteAction: false,
		},
		{
			MenuID:       HasilSurvey,
			RoleID:       RoleRwId,
			ViewAction:   false,
			CreateAction: false,
			UpdateAction: false,
			DeleteAction: false,
		},
		{
			MenuID:       HasilSurvey,
			RoleID:       RoleLurahId,
			ViewAction:   false,
			CreateAction: false,
			UpdateAction: false,
			DeleteAction: false,
		},
		{
			MenuID:       HasilSurvey,
			RoleID:       RoleCamatId,
			ViewAction:   false,
			CreateAction: false,
			UpdateAction: false,
			DeleteAction: false,
		},
		{
			MenuID:       HasilSurvey,
			RoleID:       RolePemkotId,
			ViewAction:   false,
			CreateAction: false,
			UpdateAction: false,
			DeleteAction: false,
		},
		{
			MenuID:       HasilSurvey,
			RoleID:       RoleAdminId,
			ViewAction:   true,
			CreateAction: true,
			UpdateAction: true,
			DeleteAction: true,
		},
		{
			MenuID:       HasilSurvey,
			RoleID:       RoleSurveyorId,
			ViewAction:   false,
			CreateAction: false,
			UpdateAction: false,
			DeleteAction: false,
		},
		{
			MenuID:       HasilSurvey,
			RoleID:       RoleWalikotaId,
			ViewAction:   true,
			CreateAction: true,
			UpdateAction: true,
			DeleteAction: true,
		},
		// Manajemen List Survey
		{
			MenuID:       ListSurveyID,
			RoleID:       RolePublicId,
			ViewAction:   true,
			CreateAction: false,
			UpdateAction: false,
			DeleteAction: false,
		},
		{
			MenuID:       ListSurveyID,
			RoleID:       RoleRtId,
			ViewAction:   true,
			CreateAction: false,
			UpdateAction: false,
			DeleteAction: false,
		},
		{
			MenuID:       ListSurveyID,
			RoleID:       RoleRwId,
			ViewAction:   true,
			CreateAction: false,
			UpdateAction: false,
			DeleteAction: false,
		},
		{
			MenuID:       ListSurveyID,
			RoleID:       RoleLurahId,
			ViewAction:   true,
			CreateAction: true,
			UpdateAction: true,
			DeleteAction: true,
		},
		{
			MenuID:       ListSurveyID,
			RoleID:       RoleCamatId,
			ViewAction:   true,
			CreateAction: true,
			UpdateAction: true,
			DeleteAction: true,
		},
		{
			MenuID:       ListSurveyID,
			RoleID:       RolePemkotId,
			ViewAction:   true,
			CreateAction: false,
			UpdateAction: false,
			DeleteAction: false,
		},
		{
			MenuID:       ListSurveyID,
			RoleID:       RoleAdminId,
			ViewAction:   true,
			CreateAction: true,
			UpdateAction: true,
			DeleteAction: true,
		},
		{
			MenuID:       ListSurveyID,
			RoleID:       RoleSurveyorId,
			ViewAction:   true,
			CreateAction: false,
			UpdateAction: false,
			DeleteAction: false,
		},
		{
			MenuID:       ListSurveyID,
			RoleID:       RoleWalikotaId,
			ViewAction:   true,
			CreateAction: false,
			UpdateAction: false,
			DeleteAction: false,
		},
		// Manajemen Kategori
		{
			MenuID:       CategoryID,
			RoleID:       RolePublicId,
			ViewAction:   false,
			CreateAction: false,
			UpdateAction: false,
			DeleteAction: false,
		},
		{
			MenuID:       CategoryID,
			RoleID:       RoleRtId,
			ViewAction:   false,
			CreateAction: false,
			UpdateAction: false,
			DeleteAction: false,
		},
		{
			MenuID:       CategoryID,
			RoleID:       RoleRwId,
			ViewAction:   false,
			CreateAction: false,
			UpdateAction: false,
			DeleteAction: false,
		},
		{
			MenuID:       CategoryID,
			RoleID:       RoleLurahId,
			ViewAction:   false,
			CreateAction: false,
			UpdateAction: false,
			DeleteAction: false,
		},
		{
			MenuID:       CategoryID,
			RoleID:       RoleCamatId,
			ViewAction:   true,
			CreateAction: true,
			UpdateAction: true,
			DeleteAction: true,
		},
		{
			MenuID:       CategoryID,
			RoleID:       RolePemkotId,
			ViewAction:   false,
			CreateAction: false,
			UpdateAction: false,
			DeleteAction: false,
		},
		{
			MenuID:       CategoryID,
			RoleID:       RoleAdminId,
			ViewAction:   true,
			CreateAction: true,
			UpdateAction: true,
			DeleteAction: true,
		},
		{
			MenuID:       CategoryID,
			RoleID:       RoleSurveyorId,
			ViewAction:   false,
			CreateAction: false,
			UpdateAction: false,
			DeleteAction: false,
		},
		{
			MenuID:       CategoryID,
			RoleID:       RoleWalikotaId,
			ViewAction:   false,
			CreateAction: false,
			UpdateAction: false,
			DeleteAction: false,
		},
		// Manajemen Promote
		{
			MenuID:       PromoteID,
			RoleID:       RolePublicId,
			ViewAction:   false,
			CreateAction: false,
			UpdateAction: false,
			DeleteAction: false,
		},
		{
			MenuID:       PromoteID,
			RoleID:       RoleRtId,
			ViewAction:   false,
			CreateAction: false,
			UpdateAction: false,
			DeleteAction: false,
		},
		{
			MenuID:       PromoteID,
			RoleID:       RoleRwId,
			ViewAction:   false,
			CreateAction: false,
			UpdateAction: false,
			DeleteAction: false,
		},
		{
			MenuID:       PromoteID,
			RoleID:       RoleLurahId,
			ViewAction:   false,
			CreateAction: false,
			UpdateAction: false,
			DeleteAction: false,
		},
		{
			MenuID:       PromoteID,
			RoleID:       RoleCamatId,
			ViewAction:   true,
			CreateAction: true,
			UpdateAction: true,
			DeleteAction: true,
		},
		{
			MenuID:       PromoteID,
			RoleID:       RolePemkotId,
			ViewAction:   false,
			CreateAction: false,
			UpdateAction: false,
			DeleteAction: false,
		},
		{
			MenuID:       PromoteID,
			RoleID:       RoleAdminId,
			ViewAction:   true,
			CreateAction: true,
			UpdateAction: true,
			DeleteAction: true,
		},
		{
			MenuID:       PromoteID,
			RoleID:       RoleSurveyorId,
			ViewAction:   false,
			CreateAction: false,
			UpdateAction: false,
			DeleteAction: false,
		},
		{
			MenuID:       PromoteID,
			RoleID:       RoleWalikotaId,
			ViewAction:   false,
			CreateAction: false,
			UpdateAction: false,
			DeleteAction: false,
		},
		// Promote
		{
			MenuID:       PromoteID,
			RoleID:       RolePublicId,
			ViewAction:   false,
			CreateAction: false,
			UpdateAction: false,
			DeleteAction: false,
		},
		{
			MenuID:       PromoteID,
			RoleID:       RoleRtId,
			ViewAction:   false,
			CreateAction: false,
			UpdateAction: false,
			DeleteAction: false,
		},
		{
			MenuID:       PromoteID,
			RoleID:       RoleRwId,
			ViewAction:   false,
			CreateAction: false,
			UpdateAction: false,
			DeleteAction: false,
		},
		{
			MenuID:       PromoteID,
			RoleID:       RoleLurahId,
			ViewAction:   false,
			CreateAction: false,
			UpdateAction: false,
			DeleteAction: false,
		},
		{
			MenuID:       PromoteID,
			RoleID:       RoleCamatId,
			ViewAction:   true,
			CreateAction: true,
			UpdateAction: true,
			DeleteAction: true,
		},
		{
			MenuID:       PromoteID,
			RoleID:       RolePemkotId,
			ViewAction:   false,
			CreateAction: false,
			UpdateAction: false,
			DeleteAction: false,
		},
		{
			MenuID:       PromoteID,
			RoleID:       RoleAdminId,
			ViewAction:   true,
			CreateAction: true,
			UpdateAction: true,
			DeleteAction: true,
		},
		{
			MenuID:       PromoteID,
			RoleID:       RoleSurveyorId,
			ViewAction:   false,
			CreateAction: false,
			UpdateAction: false,
			DeleteAction: false,
		},
		{
			MenuID:       PromoteID,
			RoleID:       RoleWalikotaId,
			ViewAction:   false,
			CreateAction: false,
			UpdateAction: false,
			DeleteAction: false,
		},
		// Manajemen Artikel
		{
			MenuID:       ManajemenArtikelID,
			RoleID:       RolePublicId,
			ViewAction:   false,
			CreateAction: false,
			UpdateAction: false,
			DeleteAction: false,
		},
		{
			MenuID:       ManajemenArtikelID,
			RoleID:       RoleRtId,
			ViewAction:   false,
			CreateAction: false,
			UpdateAction: false,
			DeleteAction: false,
		},
		{
			MenuID:       ManajemenArtikelID,
			RoleID:       RoleRwId,
			ViewAction:   false,
			CreateAction: false,
			UpdateAction: false,
			DeleteAction: false,
		},
		{
			MenuID:       ManajemenArtikelID,
			RoleID:       RoleLurahId,
			ViewAction:   false,
			CreateAction: false,
			UpdateAction: false,
			DeleteAction: false,
		},
		{
			MenuID:       ManajemenArtikelID,
			RoleID:       RoleCamatId,
			ViewAction:   true,
			CreateAction: true,
			UpdateAction: true,
			DeleteAction: true,
		},
		{
			MenuID:       ManajemenArtikelID,
			RoleID:       RolePemkotId,
			ViewAction:   false,
			CreateAction: false,
			UpdateAction: false,
			DeleteAction: false,
		},
		{
			MenuID:       ManajemenArtikelID,
			RoleID:       RoleAdminId,
			ViewAction:   true,
			CreateAction: true,
			UpdateAction: true,
			DeleteAction: true,
		},
		{
			MenuID:       ManajemenArtikelID,
			RoleID:       RoleSurveyorId,
			ViewAction:   false,
			CreateAction: false,
			UpdateAction: false,
			DeleteAction: false,
		},
		{
			MenuID:       ManajemenArtikelID,
			RoleID:       RoleWalikotaId,
			ViewAction:   false,
			CreateAction: false,
			UpdateAction: false,
			DeleteAction: false,
		},
		// Manajemen Pejabat
		{
			MenuID:       ManajemenPejabatID,
			RoleID:       RolePublicId,
			ViewAction:   false,
			CreateAction: false,
			UpdateAction: false,
			DeleteAction: false,
		},
		{
			MenuID:       ManajemenPejabatID,
			RoleID:       RoleRtId,
			ViewAction:   false,
			CreateAction: false,
			UpdateAction: false,
			DeleteAction: false,
		},
		{
			MenuID:       ManajemenPejabatID,
			RoleID:       RoleRwId,
			ViewAction:   false,
			CreateAction: false,
			UpdateAction: false,
			DeleteAction: false,
		},
		{
			MenuID:       ManajemenPejabatID,
			RoleID:       RoleLurahId,
			ViewAction:   false,
			CreateAction: false,
			UpdateAction: false,
			DeleteAction: false,
		},
		{
			MenuID:       ManajemenPejabatID,
			RoleID:       RoleCamatId,
			ViewAction:   false,
			CreateAction: false,
			UpdateAction: false,
			DeleteAction: false,
		},
		{
			MenuID:       ManajemenPejabatID,
			RoleID:       RolePemkotId,
			ViewAction:   false,
			CreateAction: false,
			UpdateAction: false,
			DeleteAction: false,
		},
		{
			MenuID:       ManajemenPejabatID,
			RoleID:       RoleAdminId,
			ViewAction:   true,
			CreateAction: true,
			UpdateAction: true,
			DeleteAction: true,
		},
		{
			MenuID:       ManajemenPejabatID,
			RoleID:       RoleSurveyorId,
			ViewAction:   false,
			CreateAction: false,
			UpdateAction: false,
			DeleteAction: false,
		},
		{
			MenuID:       ManajemenPejabatID,
			RoleID:       RoleWalikotaId,
			ViewAction:   false,
			CreateAction: false,
			UpdateAction: false,
			DeleteAction: false,
		},
		// Manajemen User Permission
		{
			MenuID:       ManajemenUserID,
			RoleID:       RolePublicId,
			ViewAction:   false,
			CreateAction: false,
			UpdateAction: false,
			DeleteAction: false,
		},
		{
			MenuID:       ManajemenUserID,
			RoleID:       RoleRtId,
			ViewAction:   false,
			CreateAction: false,
			UpdateAction: false,
			DeleteAction: false,
		},
		{
			MenuID:       ManajemenUserID,
			RoleID:       RoleRwId,
			ViewAction:   false,
			CreateAction: false,
			UpdateAction: false,
			DeleteAction: false,
		},
		{
			MenuID:       ManajemenUserID,
			RoleID:       RoleLurahId,
			ViewAction:   false,
			CreateAction: false,
			UpdateAction: false,
			DeleteAction: false,
		},
		{
			MenuID:       ManajemenUserID,
			RoleID:       RoleCamatId,
			ViewAction:   false,
			CreateAction: false,
			UpdateAction: false,
			DeleteAction: false,
		},
		{
			MenuID:       ManajemenUserID,
			RoleID:       RolePemkotId,
			ViewAction:   false,
			CreateAction: false,
			UpdateAction: false,
			DeleteAction: false,
		},
		{
			MenuID:       ManajemenUserID,
			RoleID:       RoleAdminId,
			ViewAction:   true,
			CreateAction: true,
			UpdateAction: true,
			DeleteAction: true,
		},
		{
			MenuID:       ManajemenUserID,
			RoleID:       RoleSurveyorId,
			ViewAction:   false,
			CreateAction: false,
			UpdateAction: false,
			DeleteAction: false,
		},
		{
			MenuID:       ManajemenUserID,
			RoleID:       RoleWalikotaId,
			ViewAction:   false,
			CreateAction: false,
			UpdateAction: false,
			DeleteAction: false,
		},
		// Manajemen Wilayah Permission
		{
			MenuID:       ManajemenWilayahID,
			RoleID:       RolePublicId,
			ViewAction:   false,
			CreateAction: false,
			UpdateAction: false,
			DeleteAction: false,
		},
		{
			MenuID:       ManajemenWilayahID,
			RoleID:       RoleRwId,
			ViewAction:   false,
			CreateAction: false,
			UpdateAction: false,
			DeleteAction: false,
		},
		{
			MenuID:       ManajemenWilayahID,
			RoleID:       RoleRtId,
			ViewAction:   false,
			CreateAction: false,
			UpdateAction: false,
			DeleteAction: false,
		},
		{
			MenuID:       ManajemenWilayahID,
			RoleID:       RoleLurahId,
			ViewAction:   true,
			CreateAction: true,
			UpdateAction: true,
			DeleteAction: true,
		},
		{
			MenuID:       ManajemenWilayahID,
			RoleID:       RoleCamatId,
			ViewAction:   false,
			CreateAction: false,
			UpdateAction: false,
			DeleteAction: false,
		},
		{
			MenuID:       ManajemenWilayahID,
			RoleID:       RolePemkotId,
			ViewAction:   false,
			CreateAction: false,
			UpdateAction: false,
			DeleteAction: false,
		},
		{
			MenuID:       ManajemenWilayahID,
			RoleID:       RoleAdminId,
			ViewAction:   true,
			CreateAction: true,
			UpdateAction: true,
			DeleteAction: true,
		},
		{
			MenuID:       ManajemenWilayahID,
			RoleID:       RoleSurveyorId,
			ViewAction:   false,
			CreateAction: false,
			UpdateAction: false,
			DeleteAction: false,
		},
		{
			MenuID:       ManajemenWilayahID,
			RoleID:       RoleWalikotaId,
			ViewAction:   false,
			CreateAction: false,
			UpdateAction: false,
			DeleteAction: false,
		},
	}

	err = db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "menu_id"}, {Name: "role_id"}},
		DoNothing: true,
	}).Create(&MenuPermission).Error

	if err != nil {
		return err
	}

	return nil
}

func GetMenuIDMap(db *gorm.DB, keys []string) (map[string]int, error) {
	var menus []models.Menu

	err := db.Where("key IN ?", keys).Find(&menus).Error
	if err != nil {
		return nil, err
	}

	menuMap := make(map[string]int)

	for _, m := range menus {
		menuMap[m.Key] = m.ID
	}

	for _, key := range keys {
		if _, ok := menuMap[key]; !ok {
			return nil, fmt.Errorf("menu key '%s' tidak ditemukan di database", key)
		}
	}

	return menuMap, nil
}

func MapToStruct(m map[string]int) models.MenuIDs {
	return models.MenuIDs{
		Template:              m["template"],
		ManajemenAlur:         m["manajemen-alur"],
		Survey:                m["survey"],
		Dashboard:             m["dashboard"],
		Pengaturan:            m["pengaturan"],
		Laporan:               m["laporan"],
		Monitoring:            m["monitoring"],
		Admin:                 m["admin"],
		FormulirPertanyaan:    m["formulir-pertanyaan"],
		Ucapan:                m["ucapan"],
		ListSurvey:            m["list-survey"],
		Hasil:                 m["hasil"],
		ManajemenPengguna:     m["manajemen-pengguna"],
		ManajemenWilayah:      m["manajemen-wilayah"],
		ManajemenCMS:          m["manajemen-cms"],
		ManajemenArtikel:      m["manajemen-artikel"],
		Rating:                m["rating"],
		Statistik:             m["statistik"],
		AktifitasSurvey:       m["aktifitas-survey"],
		ProfilSaya:            m["profil-saya"],
		Keluar:                m["keluar"],
		ManajemenResponden:    m["manajemen-responden"],
		ManajemenUser:         m["manajemen-user"],
		ManajemenBlokir:       m["manajemen-blokir"],
		ManajemenWilayahChild: m["manajemen-wilayah-child"],
		ManajemenPejabat:      m["manajemen-pejabat"],
		Artikel:               m["artikel"],
		Promote:               m["promote"],
		Kategori:              m["kategori"],
	}
}

func GetRoleIDMap(db *gorm.DB, keys []string) (map[string]int, error) {
	var role []models.Role

	err := db.Where("name IN ?", keys).Find(&role).Error
	if err != nil {
		return nil, err
	}

	menuMap := make(map[string]int)

	for _, m := range role {
		menuMap[m.Name] = m.ID
	}

	for _, key := range keys {
		if _, ok := menuMap[key]; !ok {
			return nil, fmt.Errorf("Nama Role '%s' tidak ditemukan di database", key)
		}
	}

	return menuMap, nil
}

func RoleMapToStruct(m map[string]int) models.RoleIDs {
	return models.RoleIDs{
		Public:   m["public"],
		Rt:       m["rt"],
		Rw:       m["rw"],
		Lurah:    m["lurah"],
		Camat:    m["camat"],
		Pemkot:   m["pemkot"],
		Admin:    m["admin"],
		Surveyor: m["surveyor"],
		Walikota: m["walikota"],
	}
}
