package routes

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"
)

func SetupRoutes(e *echo.Echo) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Terjadi kesalahan")
		}
	}()

	// Middleware untuk mengubah respons error
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		defer func() {
			if r := recover(); r != nil {
				fmt.Println("Terjadi kesalahan")
			}
		}()
		return func(c echo.Context) error {
			defer func() error {
				if r := recover(); r != nil {
					resp := struct {
						Data    interface{} `json:"data"`
						Message string      `json:"message"`
						Success bool        `json:"success"`
						Code    int         `json:"code"`
					}{
						Data:    "",
						Message: "Terjadi kendala pada service yang sedang anda akses",
						Success: false,
						Code:    500,
					}
					return c.JSON(int(resp.Code), resp)
				}
				return nil
			}()

			err := next(c)

			if err != nil {

				errString := err.Error()

				// Memisahkan string menjadi kode dan pesan
				parts := strings.Split(errString, ", ")

				// Mendapatkan nilai kode dan pesan
				codeStr := strings.Split(parts[0], "=")[1]
				message := strings.Split(parts[1], "=")[1]

				// Mengonversi kode menjadi integer
				code, _ := strconv.Atoi(codeStr)

				// Mengubah respons default
				var data []byte
				success := false
				return c.JSON(code, map[string]interface{}{
					"data":    data,
					"message": message,
					"success": success,
					"code":    code,
				})
			}
			return nil
		}
	})

	// deklarasi service
	userapiService := "USERAPI"
	docapiService := "DOCAPI"

	// Healthy Route API

	// sicpapi
	e.GET("/sicpapi/healthy", func(c echo.Context) error { return Healthy(c) })
	// USERAPI
	e.GET("/userapi/healthy", func(c echo.Context) error { return HandleFunc(c, userapiService) })
	// DOCAPI
	e.GET("/docapi/healthy", func(c echo.Context) error { return HandleFunc(c, docapiService) })

	// Core Route API

	// USERAPI AUTH
	e.POST("/login", func(c echo.Context) error { return HandleFunc(c, userapiService) })

	// USERAPI Service
	e.GET("/kecamatan/options", func(c echo.Context) error { return HandleFunc(c, userapiService) })
	e.GET("/kelurahan/options", func(c echo.Context) error { return HandleFunc(c, userapiService) })
	e.GET("/rw/options", func(c echo.Context) error { return HandleFunc(c, userapiService) })
	e.GET("/rt/options", func(c echo.Context) error { return HandleFunc(c, userapiService) })
}
