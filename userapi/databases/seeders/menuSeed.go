package seeders

import (
	"backend/userapi/models"
	"backend/userapi/utils"

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
			MenuName:  "Pengelola Survey",
			Key:       "pengelola-survey",
			Icon:      "questionnaire-tablet",
			CreatedBy: 0,
			UpdatedBy: 0,
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			MenuName:  "Beranda",
			Key:       "beranda",
			Icon:      "abstract-26",
			CreatedBy: 0,
			UpdatedBy: 0,
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			MenuName:  "Pengaturan Pengguna",
			Key:       "user",
			Icon:      "profile-user",
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
		{
			MenuName:  "Survey Kewilayahan",
			Key:       "survey-kewilayahan",
			Icon:      "geolocation",
			CreatedBy: 0,
			UpdatedBy: 0,
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			MenuName:  "user",
			Key:       "user",
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
	err = db.Where("key IN ?", []string{"pengelola-survey", "template", "management-alur", "beranda", "user", "laporan", "monitoring", "admin"}).Find(&parentMenus).Error
	if err != nil {
		return err
	}
	menuIDMap := make(map[string]uint)
	for _, menu := range parentMenus {
		menuIDMap[menu.Key] = uint(menu.ID)
	}
	// templateID := int(menuIDMap["template"])
	pengelolaSurvey := int(menuIDMap["pengelola-survey"])
	// surveyID := int(menuIDMap["survey"])
	pengaturanID := int(menuIDMap["user"])
	monitoringID := int(menuIDMap["monitoring"])
	adminID := int(menuIDMap["admin"])
	ChildMenus := []models.Menu{
		{
			MenuName:  "Template Pertanyaan",
			Key:       "template-pertanyaan",
			ParentID:  &pengelolaSurvey,
			Icon:      "-",
			CreatedBy: 0,
			UpdatedBy: 0,
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			MenuName:  "Manajemen Alur",
			Key:       "management-alur",
			ParentID:  &pengelolaSurvey,
			Icon:      "-",
			CreatedBy: 0,
			UpdatedBy: 0,
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			MenuName:  "Template Ucapan",
			Key:       "template-ucapan",
			ParentID:  &pengelolaSurvey,
			Icon:      "-",
			CreatedBy: 0,
			UpdatedBy: 0,
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			MenuName:  "Survey",
			Key:       "survey",
			ParentID:  &pengelolaSurvey,
			Icon:      "-",
			CreatedBy: 0,
			UpdatedBy: 0,
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			MenuName:  "Manajemen Pengguna",
			Key:       "management-pengguna",
			ParentID:  &pengaturanID,
			Icon:      "-",
			CreatedBy: 0,
			UpdatedBy: 0,
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			MenuName:  "Manajemen Wilayah",
			Key:       "manage-wilayah",
			ParentID:  &pengaturanID,
			Icon:      "-",
			CreatedBy: 0,
			UpdatedBy: 0,
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			MenuName:  "Manajemen CMS",
			Key:       "management-cms",
			ParentID:  &pengaturanID,
			Icon:      "-",
			CreatedBy: 0,
			UpdatedBy: 0,
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			MenuName:  "Manajemen Artikel",
			Key:       "management-artikel",
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
		{
			MenuName:  "Survey",
			Key:       "master-data",
			ParentID:  &pengelolaSurvey,
			Icon:      "-",
			CreatedBy: 0,
			UpdatedBy: 0,
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			MenuName:  "Profile",
			Key:       "profile",
			ParentID:  &pengelolaSurvey,
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
	err = db.Where("key IN ?", []string{"management-pengguna", "manage-wilayah", "management-artikel", "master-data"}).Find(&childMenus).Error
	if err != nil {
		return err
	}
	childMenuIDMap := make(map[string]uint)
	for _, menu := range childMenus {
		childMenuIDMap[menu.Key] = uint(menu.ID)
	}
	manajemenPenggunaID := int(childMenuIDMap["management-pengguna"])
	manajemenWilayahID := int(childMenuIDMap["manage-wilayah"])
	manajemenArtikelID := int(childMenuIDMap["management-artikel"])
	masterDataID := int(childMenuIDMap["master-data"])

	GrindChildMenus := []models.Menu{
		{
			MenuName:  "Manajemen Responden",
			Key:       "management-responden",
			ParentID:  &manajemenPenggunaID,
			Icon:      "-",
			CreatedBy: 0,
			UpdatedBy: 0,
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			MenuName:  "Manajemen User",
			Key:       "management-user",
			ParentID:  &manajemenPenggunaID,
			Icon:      "-",
			CreatedBy: 0,
			UpdatedBy: 0,
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			MenuName:  "Manajemen Blokir",
			Key:       "management-blokir",
			ParentID:  &manajemenPenggunaID,
			Icon:      "-",
			CreatedBy: 0,
			UpdatedBy: 0,
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			MenuName:  "Manajemen Wilayah",
			Key:       "management-wilayah",
			ParentID:  &manajemenWilayahID,
			Icon:      "-",
			CreatedBy: 0,
			UpdatedBy: 0,
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			MenuName:  "Manajemen Pejabat",
			Key:       "management-pejabat",
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
		{
			MenuName:  "Hasil Survey",
			Key:       "hasil-survey",
			ParentID:  &masterDataID,
			Icon:      "-",
			CreatedBy: 0,
			UpdatedBy: 0,
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			MenuName:  "List Survey",
			Key:       "list-survey",
			ParentID:  &masterDataID,
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
