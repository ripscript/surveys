package seeders

import (
	"backend/userapi/models"
	"backend/userapi/utils"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func MenuSeed(db *gorm.DB) error {
	now := utils.TimeNow()

	return db.Transaction(func(tx *gorm.DB) error {
		// Create Parent Menus
		Menus := []models.Menu{
			{
				MenuName:  "Beranda",
				Key:       "beranda",
				Icon:      "abstract-26",
				SortOrder: 1,
				CreatedBy: 0,
				UpdatedBy: 0,
				CreatedAt: now,
				UpdatedAt: now,
			},
			{
				MenuName:  "Dasbor",
				Key:       "dashboard",
				Icon:      "category",
				SortOrder: 2,
				CreatedBy: 0,
				UpdatedBy: 0,
				CreatedAt: now,
				UpdatedAt: now,
			},
			{
				MenuName:  "Pengaturan Pengguna",
				Key:       "user",
				Icon:      "profile-user",
				SortOrder: 3,
				CreatedBy: 0,
				UpdatedBy: 0,
				CreatedAt: now,
				UpdatedAt: now,
			},
			{
				MenuName:  "Pengelola Survey",
				Key:       "pengelola-survey",
				Icon:      "questionnaire-tablet",
				SortOrder: 4,
				CreatedBy: 0,
				UpdatedBy: 0,
				CreatedAt: now,
				UpdatedAt: now,
			},
			{
				MenuName:  "Monitoring Dan Laporan",
				Key:       "monitoring-dan-laporan",
				Icon:      "note-2",
				SortOrder: 5,
				CreatedBy: 0,
				UpdatedBy: 0,
				CreatedAt: now,
				UpdatedAt: now,
			},
			{
				MenuName:  "Pengaturan Aplikasi",
				Key:       "pengaturan-aplikasi",
				Icon:      "setting-2",
				SortOrder: 6,
				CreatedBy: 0,
				UpdatedBy: 0,
				CreatedAt: now,
				UpdatedAt: now,
			},
			{
				MenuName:  "Template",
				Key:       "template",
				Icon:      "-",
				SortOrder: 7,
				CreatedBy: 0,
				UpdatedBy: 0,
				CreatedAt: now,
				UpdatedAt: now,
			},

			{
				MenuName:  "Monitoring",
				Key:       "monitoring",
				Icon:      "-",
				SortOrder: 8,
				CreatedBy: 0,
				UpdatedBy: 0,
				CreatedAt: now,
				UpdatedAt: now,
			},
			{
				MenuName:  "Admin",
				Key:       "admin",
				Icon:      "-",
				SortOrder: 9,
				CreatedBy: 0,
				UpdatedBy: 0,
				CreatedAt: now,
				UpdatedAt: now,
			},
			{
				MenuName:  "Survey Kewilayahan",
				Key:       "survey-kewilayahan",
				Icon:      "geolocation",
				SortOrder: 10,
				CreatedBy: 0,
				UpdatedBy: 0,
				CreatedAt: now,
				UpdatedAt: now,
			},
		}

		if err := upsertMenus(tx, Menus); err != nil {
			return err
		}

		parentMenus, err := findMenusByKeys(tx, keysOf(Menus))
		if err != nil {
			return err
		}
		menuIDMap := toIDMap(parentMenus)

		pengelolaSurvey := int(menuIDMap["pengelola-survey"])
		pengaturanID := int(menuIDMap["user"])
		monitoringDanLaporanID := int(menuIDMap["monitoring-dan-laporan"])
		pengaturanAplikasiID := int(menuIDMap["pengaturan-aplikasi"])
		dashboardID := int(menuIDMap["dashboard"])
		// monitoringID := int(menuIDMap["monitoring"])
		// adminID := int(menuIDMap["admin"])

		// Create Child Menus
		ChildMenus := []models.Menu{
			// BEGIN: PENGATURAN PENGGUNA =======================================
			{
				MenuName:  "Manajemen Pengguna",
				Key:       "management-pengguna",
				ParentID:  &pengaturanID,
				Icon:      "-",
				SortOrder: 1,
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
				SortOrder: 2,
				CreatedBy: 0,
				UpdatedBy: 0,
				CreatedAt: now,
				UpdatedAt: now,
			},
			// END: PENGATURAN PENGGUNA =======================================

			// BEGIN: PENGELOLA SURVEY =======================================
			{
				MenuName:  "Template Ucapan",
				Key:       "template-ucapan",
				ParentID:  &pengelolaSurvey,
				Icon:      "-",
				SortOrder: 1,
				CreatedBy: 0,
				UpdatedBy: 0,
				CreatedAt: now,
				UpdatedAt: now,
			},
			{
				MenuName:  "Template Pertanyaan",
				Key:       "template-pertanyaan",
				ParentID:  &pengelolaSurvey,
				Icon:      "-",
				SortOrder: 2,
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
				SortOrder: 3,
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
				SortOrder: 4,
				CreatedBy: 0,
				UpdatedBy: 0,
				CreatedAt: now,
				UpdatedAt: now,
			},
			// END: PENGELOLA SURVEY =======================================

			// BEGIN: MONITORING DAN LAPORAN =======================================
			{
				MenuName:  "Statistik",
				Key:       "statistik",
				ParentID:  &monitoringDanLaporanID,
				Icon:      "-",
				SortOrder: 1,
				CreatedBy: 0,
				UpdatedBy: 0,
				CreatedAt: now,
				UpdatedAt: now,
			},
			{
				MenuName:  "Aktifitas Survey",
				Key:       "aktifitas-survey",
				ParentID:  &monitoringDanLaporanID,
				Icon:      "-",
				SortOrder: 2,
				CreatedBy: 0,
				UpdatedBy: 0,
				CreatedAt: now,
				UpdatedAt: now,
			},
			{
				MenuName:  "Laporan",
				Key:       "laporan",
				ParentID:  &monitoringDanLaporanID,
				Icon:      "-",
				SortOrder: 3,
				CreatedBy: 0,
				UpdatedBy: 0,
				CreatedAt: now,
				UpdatedAt: now,
			},
			// END: MONITORING DAN LAPORAN =======================================

			// BEGIN: PENGATURAN APLIKASI =======================================
			{
				MenuName:  "Manajemen CMS",
				Key:       "management-cms",
				ParentID:  &pengaturanAplikasiID,
				Icon:      "-",
				SortOrder: 1,
				CreatedBy: 0,
				UpdatedBy: 0,
				CreatedAt: now,
				UpdatedAt: now,
			},
			{
				MenuName:  "Manajemen Artikel",
				Key:       "management-artikel",
				ParentID:  &pengaturanAplikasiID,
				Icon:      "-",
				SortOrder: 2,
				CreatedBy: 0,
				UpdatedBy: 0,
				CreatedAt: now,
				UpdatedAt: now,
			},
			// END: PENGATURAN APLIKASI =======================================

			// BEGIN: DASHBOARD =======================================
			{
				MenuName:  "Dasbor Utama",
				Key:       "dashboard-utama",
				ParentID:  &dashboardID,
				Icon:      "-",
				SortOrder: 1,
				CreatedBy: 0,
				UpdatedBy: 0,
				CreatedAt: now,
				UpdatedAt: now,
			},
			{
				MenuName:  "Dasbor Wilayah",
				Key:       "dashboard-wilayah",
				ParentID:  &dashboardID,
				Icon:      "-",
				SortOrder: 2,
				CreatedBy: 0,
				UpdatedBy: 0,
				CreatedAt: now,
				UpdatedAt: now,
			},
			{
				MenuName:  "Dasbor RT",
				Key:       "dashboard-rt",
				ParentID:  &dashboardID,
				Icon:      "-",
				SortOrder: 3,
				CreatedBy: 0,
				UpdatedBy: 0,
				CreatedAt: now,
				UpdatedAt: now,
			},
			// END: DASHBOARD =======================================
		}

		if err := upsertMenus(tx, ChildMenus); err != nil {
			return err
		}

		childMenus, err := findMenusByKeys(tx, keysOf(ChildMenus))
		if err != nil {
			return err
		}
		childMenuIDMap := toIDMap(childMenus)

		manajemenPenggunaID := int(childMenuIDMap["management-pengguna"])
		manajemenWilayahID := int(childMenuIDMap["manage-wilayah"])
		manajemenArtikelID := int(childMenuIDMap["management-artikel"])
		masterDataID := int(childMenuIDMap["master-data"])

		// Create Grandchild
		GrindChildMenus := []models.Menu{
			// BEGIN: MANAJEMEN PENGGUNA =======================================
			{
				MenuName:  "Manajemen Responden",
				Key:       "management-responden",
				ParentID:  &manajemenPenggunaID,
				Icon:      "-",
				SortOrder: 1,
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
				SortOrder: 2,
				CreatedBy: 0,
				UpdatedBy: 0,
				CreatedAt: now,
				UpdatedAt: now,
			},
			// {
			// 	MenuName:  "Manajemen Blokir",
			// 	Key:       "management-blokir",
			// 	ParentID:  &manajemenPenggunaID,
			// 	Icon:      "-",
			// 	SortOrder: 3,
			// 	CreatedBy: 0,
			// 	UpdatedBy: 0,
			// 	CreatedAt: now,
			// 	UpdatedAt: now,
			// },
			// END: MANAJEMEN PENGGUNA =======================================

			// BEGIN: MANAJEMEN WILAYAH =======================================
			{
				MenuName:  "Manajemen Wilayah",
				Key:       "management-wilayah",
				ParentID:  &manajemenWilayahID,
				Icon:      "-",
				SortOrder: 1,
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
				SortOrder: 2,
				CreatedBy: 0,
				UpdatedBy: 0,
				CreatedAt: now,
				UpdatedAt: now,
			},
			// END: MANAJEMEN WILAYAH =======================================

			// BEGIN: SURVEY =======================================
			{
				MenuName:  "List Survey",
				Key:       "list-survey",
				ParentID:  &masterDataID,
				Icon:      "-",
				SortOrder: 1,
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
				SortOrder: 2,
				CreatedBy: 0,
				UpdatedBy: 0,
				CreatedAt: now,
				UpdatedAt: now,
			},
			// END: SURVEY =======================================

			// BEGIN: MANAJEMEN ARTIKEL =======================================
			{
				MenuName:  "Artikel",
				Key:       "artikel",
				ParentID:  &manajemenArtikelID,
				Icon:      "-",
				SortOrder: 1,
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
				SortOrder: 2,
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
				SortOrder: 3,
				CreatedBy: 0,
				UpdatedBy: 0,
				CreatedAt: now,
				UpdatedAt: now,
			},
			// END: MANAJEMEN ARTIKEL =======================================
		}

		if err := upsertMenus(tx, GrindChildMenus); err != nil {
			return err
		}

		// Sinkronisasi: hapus menu yang ada di DB tapi sudah tidak ada di kode (di semua level).
		// PERINGATAN: destructive. Menu yang dihapus dari kode akan otomatis
		// terhapus dari database (beserta menu_permissions terkait) pada run berikutnya.
		allKeys := append(append(keysOf(Menus), keysOf(ChildMenus)...), keysOf(GrindChildMenus)...)
		return syncDeleteRemovedMenus(tx, allKeys)
	})
}

func upsertMenus(tx *gorm.DB, menus []models.Menu) error {
	return tx.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "key"}},
		DoUpdates: clause.AssignmentColumns([]string{"menu_name", "icon", "parent_id", "sort_order", "updated_by", "updated_at"}),
	}).Create(&menus).Error
}

func findMenusByKeys(tx *gorm.DB, keys []string) ([]models.Menu, error) {
	var menus []models.Menu
	err := tx.Where("key IN ?", keys).Find(&menus).Error
	return menus, err
}

func toIDMap(menus []models.Menu) map[string]uint {
	m := make(map[string]uint)
	for _, menu := range menus {
		m[menu.Key] = uint(menu.ID)
	}
	return m
}

func keysOf(menus []models.Menu) []string {
	keys := make([]string, len(menus))
	for i, m := range menus {
		keys[i] = m.Key
	}
	return keys
}

// syncDeleteRemovedMenus menghapus row menus (beserta menu_permissions terkait)
// yang key-nya tidak lagi ada di validKeys.
func syncDeleteRemovedMenus(tx *gorm.DB, validKeys []string) error {
	var staleMenus []models.Menu
	if err := tx.Where("key NOT IN ?", validKeys).Find(&staleMenus).Error; err != nil {
		return err
	}
	if len(staleMenus) == 0 {
		return nil
	}

	staleIDs := make([]int, len(staleMenus))
	for i, m := range staleMenus {
		staleIDs[i] = int(m.ID)
	}

	if err := tx.Where("menu_id IN ?", staleIDs).Delete(&models.MenuPermission{}).Error; err != nil {
		return err
	}

	return tx.Where("id IN ?", staleIDs).Delete(&models.Menu{}).Error
}
