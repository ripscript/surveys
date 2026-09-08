package utils

import (
	"fmt"
	"time"
	_ "time/tzdata"
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

func ParseToWIB(t time.Time) time.Time {
	loc, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		loc = time.FixedZone("WIB", 7*60*60)
	}
	return time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), loc)
}

func NowWIB() time.Time {
	loc, _ := time.LoadLocation("Asia/Jakarta")
	return time.Now().In(loc)
}
