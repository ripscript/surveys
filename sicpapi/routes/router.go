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
						Token   string      `json:"token"`
					}{
						Data:    "",
						Message: "Terjadi kendala pada service yang sedang anda akses",
						Success: false,
						Code:    500,
						Token:   "",
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
	masterapiService := "MASTERAPI"

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
	e.POST("/logout", func(c echo.Context) error { return HandleFunc(c, userapiService) })

	// USERAPI Service
	e.GET("/kecamatan/options", func(c echo.Context) error { return HandleFunc(c, userapiService) })
	e.GET("/kelurahan/options", func(c echo.Context) error { return HandleFunc(c, userapiService) })
	e.GET("/rw/options", func(c echo.Context) error { return HandleFunc(c, userapiService) })
	e.GET("/rt/options", func(c echo.Context) error { return HandleFunc(c, userapiService) })

	// Responden Management
	e.GET("/respondent", func(c echo.Context) error { return HandleFunc(c, userapiService) })
	e.GET("/respondent/:id", func(c echo.Context) error { return HandleFunc(c, userapiService) })
	e.DELETE("/respondent/:id", func(c echo.Context) error { return HandleFunc(c, userapiService) })
	e.PUT("/respondent/:id", func(c echo.Context) error { return HandleFunc(c, userapiService) })

	// Users Management
	e.GET("/users", func(c echo.Context) error { return HandleFunc(c, userapiService) })
	e.GET("/users/:id", func(c echo.Context) error { return HandleFunc(c, userapiService) })
	e.PUT("/users/:id", func(c echo.Context) error { return HandleFunc(c, userapiService) })
	e.DELETE("/users/:id", func(c echo.Context) error { return HandleFunc(c, userapiService) })

	e.PUT("/reset/password/:id", func(c echo.Context) error { return HandleFunc(c, userapiService) })

	// MASTERAPI Service
	e.GET("/masterapi/healthy", func(c echo.Context) error { return HandleFunc(c, masterapiService) })
	manajemenWilayahGroup := e.Group("/manajemen-wilayah")
	manajemenWilayahKecamatanGroup := manajemenWilayahGroup.Group("/kecamatan")
	manajemenWilayahKecamatanGroup.GET("/list", func(c echo.Context) error { return HandleFunc(c, masterapiService) })
	manajemenWilayahKecamatanGroup.GET("/detail/:kecamatan_id", func(c echo.Context) error { return HandleFunc(c, masterapiService) })
	manajemenWilayahKecamatanGroup.PUT("/update/:kecamatan_id", func(c echo.Context) error { return HandleFunc(c, masterapiService) })
	manajemenWilayahGroup.GET("/kecamatan/options", func(c echo.Context) error { return HandleFunc(c, masterapiService) })

	manajemenWilayahKelurahanGroup := manajemenWilayahGroup.Group("/kelurahan")
	manajemenWilayahKelurahanGroup.GET("/detail/:kelurahan_id", func(c echo.Context) error { return HandleFunc(c, masterapiService) })
	manajemenWilayahKelurahanGroup.PUT("/update/:kelurahan_id", func(c echo.Context) error { return HandleFunc(c, masterapiService) })
	manajemenWilayahKelurahanGroup.GET("/list", func(c echo.Context) error { return HandleFunc(c, masterapiService) })
	manajemenWilayahKelurahanGroup.GET("/list/:kecamatan_id", func(c echo.Context) error { return HandleFunc(c, masterapiService) })
	manajemenWilayahKelurahanGroup.GET("/options", func(c echo.Context) error { return HandleFunc(c, masterapiService) })

	manajemenWilayahRwGroup := manajemenWilayahGroup.Group("/rw")
	manajemenWilayahRwGroup.GET("/detail/:rw_id", func(c echo.Context) error { return HandleFunc(c, masterapiService) })
	manajemenWilayahRwGroup.PUT("/update/:rw_id", func(c echo.Context) error { return HandleFunc(c, masterapiService) })
	manajemenWilayahRwGroup.GET("/list", func(c echo.Context) error { return HandleFunc(c, masterapiService) })
	manajemenWilayahRwGroup.GET("/list/:kelurahan_id", func(c echo.Context) error { return HandleFunc(c, masterapiService) })
	manajemenWilayahRwGroup.POST("/create", func(c echo.Context) error { return HandleFunc(c, masterapiService) })
	manajemenWilayahGroup.GET("/rw/options", func(c echo.Context) error { return HandleFunc(c, masterapiService) })

	manajemenWilayahRtGroup := manajemenWilayahGroup.Group("/rt")
	manajemenWilayahRtGroup.GET("/detail/:rt_id", func(c echo.Context) error { return HandleFunc(c, masterapiService) })
	manajemenWilayahRtGroup.PUT("/update/:rt_id", func(c echo.Context) error { return HandleFunc(c, masterapiService) })
	manajemenWilayahRtGroup.GET("/list", func(c echo.Context) error { return HandleFunc(c, masterapiService) })
	manajemenWilayahRtGroup.GET("/list/:rw_id", func(c echo.Context) error { return HandleFunc(c, masterapiService) })
	manajemenWilayahRtGroup.POST("/create", func(c echo.Context) error { return HandleFunc(c, masterapiService) })
	manajemenWilayahGroup.GET("/rt/options", func(c echo.Context) error { return HandleFunc(c, masterapiService) })

	manajemenPenggunaGroup := e.Group("/manajemen-pengguna")
	manajemenPenggunaGroup.GET("/role/options", func(c echo.Context) error { return HandleFunc(c, masterapiService) })
	manajemenPenggunaRespondenGroup := manajemenPenggunaGroup.Group("/responden")
	manajemenPenggunaRespondenGroup.POST("/create", func(c echo.Context) error { return HandleFunc(c, masterapiService) })

	manajemenArtikelGroup := e.Group("/manajemen-artikel")
	manajemenArtikelKategoriGroup := manajemenArtikelGroup.Group("/kategori")
	manajemenArtikelKategoriGroup.POST("/create", func(c echo.Context) error { return HandleFunc(c, masterapiService) })
	manajemenArtikelKategoriGroup.PUT("/update/:id", func(c echo.Context) error { return HandleFunc(c, masterapiService) })
	manajemenArtikelKategoriGroup.DELETE("/delete/:id", func(c echo.Context) error { return HandleFunc(c, masterapiService) })
	manajemenArtikelKategoriGroup.GET("/detail/:id", func(c echo.Context) error { return HandleFunc(c, masterapiService) })
	manajemenArtikelKategoriGroup.GET("/list", func(c echo.Context) error { return HandleFunc(c, masterapiService) })
}
