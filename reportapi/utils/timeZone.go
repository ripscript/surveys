package utils

import (
	"fmt"
	"time"
)

func TimeNow() time.Time {
	defer func() {
		if r := recover(); r != nil {
			message := fmt.Sprintf("Terjadi kendala pada service yang sedang anda akses: %v", r)
			LogErrors(message)
		}
	}()
	location, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		return time.Now().UTC()
	}
	return time.Now().In(location)
}

func ParseToWIB(t time.Time) time.Time {
	loc, _ := time.LoadLocation("Asia/Jakarta")
	return time.Date(
		t.Year(),
		t.Month(),
		t.Day(),
		t.Hour(),
		t.Minute(),
		t.Second(),
		0, loc,
	)
}

func NowWIB() time.Time {
	loc, _ := time.LoadLocation("Asia/Jakarta")
	return time.Now().In(loc)
}
