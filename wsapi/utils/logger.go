// utils/logger.go
package utils

import (
	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
)

var Logger *logrus.Logger

func init() {
	Logger = NewLogger()
}

func NewLogger() *logrus.Logger {
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})
	return logger
}

// LogErrors mencatat pesan error umum (dipanggil dari recover() manual).
func LogErrors(message string) {
	Logger.Error(message)
}

// ErrorLogger middleware Echo untuk menangani panic & error tak tertangani.
func ErrorLogger() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			defer func() {
				if r := recover(); r != nil {
					Logger.WithFields(logrus.Fields{
						"panic": r,
						"path":  c.Path(),
					}).Error("Panic recovered")
				}
			}()

			err := next(c)
			if err != nil {
				Logger.WithFields(logrus.Fields{
					"error": err,
					"path":  c.Path(),
				}).Error("Unhandled error")
			}
			return err
		}
	}
}
