package utils

import (
	"fmt"
	"time"
)

func TimeNow() *time.Location {
	defer func() {
		if r := recover(); r != nil {
			message := fmt.Sprintf("Terjadi kendala pada service yang sedang anda akses: %v", r)
			LogErrors(message)
		}
	}()
	// Gunakan Asia/Jakarta sebagai zona waktu
	location, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		// Jika gagal, kembalikan zona waktu default UTC
		return time.UTC
	}
	return location
}
