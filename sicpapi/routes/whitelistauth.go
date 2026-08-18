package routes

import "fmt"

// Whitelist menyimpan daftar URL dan metode HTTP yang akan di-white list dari auth.
var Whitelist = map[string]map[string]bool{
	"/login":             {"POST": true},
	"/userapi/healthy":   {"GET": true},
	"/reportapi/healthy": {"GET": true},

	"/kecamatan/options": {"GET": true},
	"/kelurahan/options": {"GET": true},
	"/rw/options":        {"GET": true},
	"/rt/options":        {"GET": true},

	// DOCAPI SERVICE
	"/docapi/healthy":                  {"GET": true},
	"/view-public-survey-image/:path":  {"GET": true},
	"/view-cms-image/:path":            {"GET": true},
	"/view-laporan-konten-image/:path": {"GET": true},
	"/view-foto-profil/:path":          {"GET": true},

	// MASTERAPI SERVICE
	"/masterapi/healthy":                                        {"GET": true},
	"/tabel-data-kota-bandung":                                  {"GET": true},
	"/pengaturan-aplikasi/manajemen-cms/landing-page":           {"GET": true},
	"/pengaturan-aplikasi/geojson-kota-bandung-level-kecamatan": {"GET": true},

	// SURVEYAPI SERVICE
	"/surveyapi/healthy": {"GET": true},
	"/survey/show-image/:survey_code/:code_wilayah/:path": {"GET": true},
	"/survey/public-options":                              {"GET": true},
	"/survey/public-question-options":                     {"GET": true},
	"/survey/heat-point-wilayah-kota-bandung":             {"POST": true},

	"/rating/save": {"POST": true},

	// PYREPORTAPI SERVICE
	"/py-reportapi/healthy": {"GET": true},
	// WEBSOCKET SERVICE
	"/wsapi/healthy": {"GET": true},
}

// IsInWhitelist adalah fungsi untuk memeriksa apakah sebuah permintaan ada di dalam whitelist.
func IsInWhitelist(url, method string) bool {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Terjadi kesalahan")
		}
	}()
	if _, ok := Whitelist[url][method]; ok {
		return true
	}
	return false
}
