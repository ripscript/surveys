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
