package seeders

import (
	"backend/userapi/models"
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func PermissionSeed(db *gorm.DB) error {
	var menuKeys = []string{"template", "management-alur", "master-data", "beranda", "user", "laporan", "monitoring",
		"admin", "template-pertanyaan", "template-ucapan", "survey", "management-pengguna", "manage-wilayah",
		"management-cms", "management-artikel", "rating", "statistik", "aktifitas-survey", "profil-saya", "keluar",
		"management-responden", "management-user", "management-blokir", "management-wilayah", "management-pejabat",
		"artikel", "promote", "kategori", "list-survey", "profile", "survey-kewilayahan", "pengelola-survey",
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

	full := func(menuID, roleID int) models.MenuPermission {
		return models.MenuPermission{
			MenuID: menuID, RoleID: roleID,
			ViewAction: true, CreateAction: true, UpdateAction: true, DeleteAction: true,
		}
	}
	none := func(menuID, roleID int) models.MenuPermission {
		return models.MenuPermission{
			MenuID: menuID, RoleID: roleID,
			ViewAction: false, CreateAction: false, UpdateAction: false, DeleteAction: false,
		}
	}
	viewOnly := func(menuID, roleID int) models.MenuPermission {
		return models.MenuPermission{
			MenuID: menuID, RoleID: roleID,
			ViewAction: true, CreateAction: false, UpdateAction: false, DeleteAction: false,
		}
	}

	type P = models.MenuPermission
	var (
		beranda            = menuIDs.Beranda
		masterData         = menuIDs.MasterData
		pengaturan         = menuIDs.Pengaturan
		respondent         = menuIDs.ManajemenResponden
		pengguna           = menuIDs.ManajemenPengguna
		ucapan             = menuIDs.Ucapan
		formulir           = menuIDs.FormulirPertanyaan
		alur               = menuIDs.ManajemenAlur
		ManajemenWilayah   = menuIDs.ManajemenWilayah
		ManajemenResponden = menuIDs.ManajemenResponden
		Surveys            = menuIDs.Surveys
		category           = menuIDs.Kategori
		ListSurvey         = menuIDs.ListSurvey
		promote            = menuIDs.Promote
		artikel            = menuIDs.ManajemenArtikel
		pejabat            = menuIDs.ManajemenPejabat
		mUser              = menuIDs.ManajemenUser
		mWilayah           = menuIDs.ManajemenWilayahChild
		profile            = menuIDs.Profile
		SurveyKewilayahan  = menuIDs.SurveyKewilayahan
		PengelolaSurvey    = menuIDs.PengelolaSurvey

		public   = roleIDs.Public
		rt       = roleIDs.Rt
		rw       = roleIDs.Rw
		lurah    = roleIDs.Lurah
		camat    = roleIDs.Camat
		pemkot   = roleIDs.Pemkot
		admin    = roleIDs.Admin
		surveyor = roleIDs.Surveyor
		walikota = roleIDs.Walikota
	)

	MenuPermission := []P{

		viewOnly(PengelolaSurvey, public), viewOnly(PengelolaSurvey, rt), viewOnly(PengelolaSurvey, rw), viewOnly(PengelolaSurvey, lurah),
		full(PengelolaSurvey, camat), viewOnly(PengelolaSurvey, pemkot), full(PengelolaSurvey, admin),
		viewOnly(PengelolaSurvey, surveyor), viewOnly(PengelolaSurvey, walikota),

		// Survey Kewilayahan
		viewOnly(SurveyKewilayahan, public), viewOnly(SurveyKewilayahan, rt), viewOnly(SurveyKewilayahan, rw), viewOnly(SurveyKewilayahan, lurah),
		viewOnly(SurveyKewilayahan, camat), viewOnly(SurveyKewilayahan, pemkot), full(SurveyKewilayahan, admin),
		viewOnly(SurveyKewilayahan, surveyor), viewOnly(SurveyKewilayahan, walikota),

		// List Survey
		full(profile, public), full(profile, rt), full(profile, rw), full(profile, lurah),
		full(profile, camat), full(profile, pemkot), full(profile, admin),
		full(profile, surveyor), full(profile, walikota),

		// List Survey
		none(ListSurvey, public), none(ListSurvey, rt), none(ListSurvey, rw), none(ListSurvey, lurah),
		full(ListSurvey, camat), none(ListSurvey, pemkot), full(ListSurvey, admin),
		none(ListSurvey, surveyor), none(ListSurvey, walikota),

		// Manajemen Responden
		none(ManajemenResponden, public), viewOnly(ManajemenResponden, rt), viewOnly(ManajemenResponden, rw), viewOnly(ManajemenResponden, lurah),
		viewOnly(ManajemenResponden, camat), none(ManajemenResponden, pemkot), full(ManajemenResponden, admin),
		none(ManajemenResponden, surveyor), none(ManajemenResponden, walikota),

		// Manajemen Wilayah
		viewOnly(ManajemenWilayah, public), viewOnly(ManajemenWilayah, rt), viewOnly(ManajemenWilayah, rw), full(ManajemenWilayah, lurah),
		full(ManajemenWilayah, camat), viewOnly(ManajemenWilayah, pemkot), full(ManajemenWilayah, admin),
		viewOnly(ManajemenWilayah, surveyor), viewOnly(ManajemenWilayah, walikota),

		// Beranda
		none(beranda, public), none(beranda, rt), none(beranda, rw), full(beranda, lurah),
		none(beranda, camat), none(beranda, pemkot), full(beranda, admin),
		none(beranda, surveyor), none(beranda, walikota),

		// Pengelola Survey (Master Data)
		none(masterData, public), none(masterData, rt), none(masterData, rw), full(masterData, lurah),
		none(masterData, camat), none(masterData, pemkot), full(masterData, admin),
		none(masterData, surveyor), none(masterData, walikota),

		// Pengaturan
		none(pengaturan, public), none(pengaturan, rt), none(pengaturan, rw), full(pengaturan, lurah),
		none(pengaturan, camat), none(pengaturan, pemkot), full(pengaturan, admin),
		none(pengaturan, surveyor), none(pengaturan, walikota),

		// Manajemen Responden
		none(respondent, public), none(respondent, rt), none(respondent, rw), none(respondent, lurah),
		none(respondent, camat), none(respondent, pemkot), full(respondent, admin),
		none(respondent, surveyor), none(respondent, walikota),

		// Ucapan
		none(ucapan, public), none(ucapan, rt), none(ucapan, rw), none(ucapan, lurah),
		none(ucapan, camat), none(ucapan, pemkot), full(ucapan, admin),
		none(ucapan, surveyor), none(ucapan, walikota),

		// Formulir Pertanyaan
		none(formulir, public), none(formulir, rt), none(formulir, rw), none(formulir, lurah),
		none(formulir, camat), none(formulir, pemkot), full(formulir, admin),
		none(formulir, surveyor), none(formulir, walikota),

		// Manajemen Alur
		none(alur, public), none(alur, rt), none(alur, rw), none(alur, lurah),
		none(alur, camat), none(alur, pemkot), full(alur, admin),
		none(alur, surveyor), none(alur, walikota),

		// List Survey
		viewOnly(Surveys, public), viewOnly(Surveys, rt), viewOnly(Surveys, rw), full(Surveys, lurah),
		full(Surveys, camat), viewOnly(Surveys, pemkot), full(Surveys, admin),
		viewOnly(Surveys, surveyor), viewOnly(Surveys, walikota),

		// Kategori
		none(category, public), none(category, rt), none(category, rw), none(category, lurah),
		full(category, camat), none(category, pemkot), full(category, admin),
		none(category, surveyor), none(category, walikota),

		// Promote
		none(promote, public), none(promote, rt), none(promote, rw), none(promote, lurah),
		full(promote, camat), none(promote, pemkot), full(promote, admin),
		none(promote, surveyor), none(promote, walikota),

		// Manajemen Artikel
		none(artikel, public), none(artikel, rt), none(artikel, rw), none(artikel, lurah),
		full(artikel, camat), none(artikel, pemkot), full(artikel, admin),
		none(artikel, surveyor), none(artikel, walikota),

		// Manajemen Pejabat
		none(pejabat, public), none(pejabat, rt), none(pejabat, rw), none(pejabat, lurah),
		none(pejabat, camat), none(pejabat, pemkot), full(pejabat, admin),
		none(pejabat, surveyor), none(pejabat, walikota),

		// Manajemen Pengguna
		none(pengguna, public), none(pengguna, rt), none(pengguna, rw), none(pengguna, lurah),
		none(pengguna, camat), none(pengguna, pemkot), full(pengguna, admin),
		none(pengguna, surveyor), none(pengguna, walikota),

		// Manajemen User
		none(mUser, public), none(mUser, rt), none(mUser, rw), none(mUser, lurah),
		none(mUser, camat), none(mUser, pemkot), full(mUser, admin),
		none(mUser, surveyor), none(mUser, walikota),

		// Manajemen Wilayah
		none(mWilayah, public), none(mWilayah, rt), none(mWilayah, rw), full(mWilayah, lurah),
		none(mWilayah, camat), none(mWilayah, pemkot), full(mWilayah, admin),
		none(mWilayah, surveyor), none(mWilayah, walikota),
	}

	err = db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "menu_id"}, {Name: "role_id"}},
		DoNothing: true,
	}).Create(&MenuPermission).Error

	return err
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
		Template:           m["template"],
		ManajemenAlur:      m["management-alur"],
		Beranda:            m["beranda"],
		Pengaturan:         m["user"],
		Laporan:            m["laporan"],
		Monitoring:         m["monitoring"],
		Admin:              m["admin"],
		FormulirPertanyaan: m["template-pertanyaan"],
		Ucapan:             m["template-ucapan"],
		Surveys:            m["survey"],
		// Hasil:                 m["hasil"],
		ManajemenPengguna:     m["management-pengguna"],
		ManajemenWilayah:      m["manage-wilayah"],
		ManajemenCMS:          m["management-cms"],
		ManajemenArtikel:      m["management-artikel"],
		Rating:                m["rating"],
		Statistik:             m["statistik"],
		AktifitasSurvey:       m["aktifitas-survey"],
		ProfilSaya:            m["profil-saya"],
		Keluar:                m["keluar"],
		ManajemenResponden:    m["management-responden"],
		ManajemenUser:         m["management-user"],
		ManajemenBlokir:       m["management-blokir"],
		ManajemenWilayahChild: m["management-wilayah"],
		ManajemenPejabat:      m["management-pejabat"],
		Artikel:               m["artikel"],
		Promote:               m["promote"],
		Kategori:              m["kategori"],
		ListSurvey:            m["list-survey"],
		MasterData:            m["master-data"],
		Profile:               m["profile"],
		SurveyKewilayahan:     m["survey-kewilayahan"],
		PengelolaSurvey:       m["pengelola-survey"],
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
