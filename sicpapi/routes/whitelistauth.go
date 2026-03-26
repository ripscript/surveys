package routes

import "fmt"

// Whitelist menyimpan daftar URL dan metode HTTP yang akan di-white list dari auth.
var Whitelist = map[string]map[string]bool{
	"/login":             {"POST": true},
	"/userapi/healthy":   {"GET": true},
	"/docapi/healthy":    {"GET": true},
	"/kecamatan/options": {"GET": true},
	"/kelurahan/options": {"GET": true},
	"/rw/options":        {"GET": true},
	"/rt/options":        {"GET": true},

	// MASTERAPI SERVICE
	"/masterapi/healthy":                                {"GET": true},
	"/manajemen-wilayah/kecamatan/list":                 {"GET": true},
	"/manajemen-wilayah/kecamatan/detail/:kecamatan_id": {"GET": true},
	"/manajemen-wilayah/kecamatan/update/:kecamatan_id": {"PUT": true},
	"/manajemen-wilayah/kecamatan/options":              {"GET": true},

	"/manajemen-wilayah/kelurahan/list":                 {"GET": true},
	"/manajemen-wilayah/kelurahan/list/:kecamatan_id":   {"GET": true},
	"/manajemen-wilayah/kelurahan/detail/:kelurahan_id": {"GET": true},
	"/manajemen-wilayah/kelurahan/update/:kelurahan_id": {"PUT": true},
	"/manajemen-wilayah/kelurahan/options":              {"GET": true},

	"/manajemen-wilayah/rw/list":               {"GET": true},
	"/manajemen-wilayah/rw/list/:kelurahan_id": {"GET": true},
	"/manajemen-wilayah/rw/detail/:rw_id":      {"GET": true},
	"/manajemen-wilayah/rw/update/:rw_id":      {"PUT": true},
	"/manajemen-wilayah/rw/create":             {"POST": true},
	"/manajemen-wilayah/rw/options":            {"GET": true},

	"/manajemen-wilayah/rt/list":          {"GET": true},
	"/manajemen-wilayah/rt/list/:rw_id":   {"GET": true},
	"/manajemen-wilayah/rt/detail/:rt_id": {"GET": true},
	"/manajemen-wilayah/rt/update/:rt_id": {"PUT": true},
	"/manajemen-wilayah/rt/create":        {"POST": true},
	"/manajemen-wilayah/rt/options":       {"GET": true},

	"/manajemen-pengguna/role/options":     {"GET": true},
	"/manajemen-pengguna/responden/create": {"POST": true},

	"/manajemen-artikel/kategori/create":     {"POST": true},
	"/manajemen-artikel/kategori/update/:id": {"PUT": true},
	"/manajemen-artikel/kategori/delete/:id": {"DELETE": true},
	"/manajemen-artikel/kategori/detail/:id": {"GET": true},
	"/manajemen-artikel/kategori/list":       {"GET": true},
	// Tambahkan URL dan metode lainnya sesuai kebutuhan Anda
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
