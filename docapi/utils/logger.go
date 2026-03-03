// utils/logger.go
package utils

import (
	"fmt"

	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
)

var Logger *logrus.Logger

func init() {
	defer func() {
		if r := recover(); r != nil {
			message := fmt.Sprintf("Terjadi kendala pada service yang sedang anda akses: %v", r)
			LogErrors(message)
		}
	}()
	Logger = NewLogger()
}

func NewLogger() *logrus.Logger {
	defer func() {
		if r := recover(); r != nil {
			message := fmt.Sprintf("Terjadi kendala pada service yang sedang anda akses: %v", r)
			LogErrors(message)
		}
	}()
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})

	return logger
}

func ErrorLogger() echo.MiddlewareFunc {
	defer func() {
		if r := recover(); r != nil {
			message := fmt.Sprintf("Terjadi kendala pada service yang sedang anda akses: %v", r)
			LogErrors(message)
		}
	}()
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		defer func() {
			if r := recover(); r != nil {
				message := fmt.Sprintf("Terjadi kendala pada service yang sedang anda akses: %v", r)
				LogErrors(message)
			}
		}()
		return func(c echo.Context) error {
			defer func() {
				if r := recover(); r != nil {
					message := fmt.Sprintf("Terjadi kendala pada service yang sedang anda akses: %v", r)
					LogErrors(message)
				}
			}()
			if err := next(c); err != nil {
				Logger.WithFields(logrus.Fields{
					"error": err,
				}).Error("Unhandled error")
			}
			return nil
		}
	}
}
