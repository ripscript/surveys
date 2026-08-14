package utils

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

func LogErrorsOld(message string) {
	_, file, line, _ := runtime.Caller(1)
	formattedMessage := fmt.Sprintf("%s:%d - %s", filepath.Base(file), line, message)
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Terjadi Kesalahan")
		}
	}()

	// Print to console
	fmt.Println(formattedMessage)

	// Create the logs directory if it doesn't exist
	logDir := "logs"
	if _, err := os.Stat(logDir); os.IsNotExist(err) {
		err := os.Mkdir(logDir, 0755)
		if err != nil {
			log.Fatalf("failed to create log directory: %v", err)
		}
	}

	// Create or rotate log file
	currentTime := time.Now()
	logFileName := fmt.Sprintf("%s/app.log", logDir)
	archiveLogFileName := fmt.Sprintf("%s/app_%d-%02d-%02d.log", logDir, currentTime.Year(), currentTime.Month(), currentTime.Day())

	// Check if app.log exists
	if stat, err := os.Stat(logFileName); err == nil {
		// Check if the modification date is today
		if !isSameDay(stat.ModTime(), currentTime) {
			// Rename the existing app.log to app_YYYY-MM-DD.log
			err := os.Rename(logFileName, archiveLogFileName)
			if err != nil {
				log.Fatalf("failed to rename log file: %v", err)
			}
		}
	}

	// Open the log file (app.log)
	logFile, err := os.OpenFile(logFileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatalf("failed to open error log file: %v", err)
	}
	defer logFile.Close()

	// Create a logger and log the formatted message
	logger := log.New(logFile, "", log.LstdFlags)
	logger.Println(formattedMessage)
}

// isSameDay checks if two times are on the same day
func isSameDay(t1, t2 time.Time) bool {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Terjadi Kesalahan")
		}
	}()
	y1, m1, d1 := t1.Date()
	y2, m2, d2 := t2.Date()
	return y1 == y2 && m1 == m2 && d1 == d2
}
