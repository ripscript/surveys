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
	reportapiService := "REPORTAPI"
	docapiService := "DOCAPI"
	masterapiService := "MASTERAPI"
	surveyapiService := "SURVEYAPI"
	pyReportApiService := "PYREPORTAPI"

	// Healthy Route API

	// sicpapi
	e.GET("/sicpapi/healthy", func(c echo.Context) error { return Healthy(c) })
	// USERAPI
	e.GET("/userapi/healthy", func(c echo.Context) error { return HandleFunc(c, userapiService) })
	// DOCAPI
	e.GET("/docapi/healthy", func(c echo.Context) error { return HandleFunc(c, docapiService) })
	// REPORTAPI
	e.GET("/reportapi/healthy", func(c echo.Context) error { return HandleFunc(c, reportapiService) })
	// PYREPORTAPI
	e.GET("/py-reportapi/healthy", func(c echo.Context) error { return HandleFunc(c, pyReportApiService) })

	// Core Route API
	// USERAPI AUTH
	e.POST("/login", func(c echo.Context) error { return HandleFunc(c, userapiService) })
	e.POST("/logout", func(c echo.Context) error { return HandleFunc(c, userapiService) })
	e.PUT("/reset/password/:id", func(c echo.Context) error { return HandleFunc(c, userapiService) })
	// USERAPI Service
	e.GET("/kecamatan/options", func(c echo.Context) error { return HandleFunc(c, userapiService) })
	e.GET("/kelurahan/options", func(c echo.Context) error { return HandleFunc(c, userapiService) })
	e.GET("/rw/options", func(c echo.Context) error { return HandleFunc(c, userapiService) })
	e.GET("/rt/options", func(c echo.Context) error { return HandleFunc(c, userapiService) })
	// Responden Management
	e.GET("/respondent", func(c echo.Context) error { return HandleFunc(c, userapiService) })
	e.POST("/respondent", func(c echo.Context) error { return HandleFunc(c, userapiService) })
	e.POST("/respondent/import", func(c echo.Context) error { return HandleFunc(c, userapiService) })
	e.GET("/respondent/import", func(c echo.Context) error { return HandleFunc(c, userapiService) })
	e.GET("/respondent/:id", func(c echo.Context) error { return HandleFunc(c, userapiService) })
	e.DELETE("/respondent/:id", func(c echo.Context) error { return HandleFunc(c, userapiService) })
	e.PUT("/respondent/:id", func(c echo.Context) error { return HandleFunc(c, userapiService) })
	e.GET("/respondent/raw/:id", func(c echo.Context) error { return HandleFunc(c, userapiService) })
	e.PUT("/respondent/block", func(c echo.Context) error { return HandleFunc(c, userapiService) })
	e.GET("/respondent/options", func(c echo.Context) error { return HandleFunc(c, userapiService) })
	e.PUT("/respondent/:id/update-password", func(c echo.Context) error { return HandleFunc(c, userapiService) })
	// Users Management
	e.GET("/get-profile", func(c echo.Context) error { return HandleFunc(c, userapiService) })
	e.PUT("/update-profile", func(c echo.Context) error { return HandleFunc(c, userapiService) })
	e.GET("/users", func(c echo.Context) error { return HandleFunc(c, userapiService) })
	e.POST("/users", func(c echo.Context) error { return HandleFunc(c, userapiService) })
	e.GET("/users/:id", func(c echo.Context) error { return HandleFunc(c, userapiService) })
	e.PUT("/users/:id", func(c echo.Context) error { return HandleFunc(c, userapiService) })
	e.DELETE("/users/:id", func(c echo.Context) error { return HandleFunc(c, userapiService) })
	e.GET("/users/export", func(c echo.Context) error { return HandleFunc(c, userapiService) })
	// Users Blokir Management
	e.GET("/users/blokir", func(c echo.Context) error { return HandleFunc(c, userapiService) })
	e.PUT("/users/blokir/:id", func(c echo.Context) error { return HandleFunc(c, userapiService) })
	// Surveyor
	e.GET("/surveyor/options", func(c echo.Context) error { return HandleFunc(c, userapiService) })
	// Get List Menu Permission
	e.GET("/menus/permission", func(c echo.Context) error { return HandleFunc(c, userapiService) })
	// Report
	e.GET("/count/kecamatan", func(c echo.Context) error { return HandleFunc(c, reportapiService) })
	e.GET("/count/kelurahan", func(c echo.Context) error { return HandleFunc(c, reportapiService) })
	e.GET("/count/rw", func(c echo.Context) error { return HandleFunc(c, reportapiService) })
	e.GET("/count/rt", func(c echo.Context) error { return HandleFunc(c, reportapiService) })
	e.GET("/count/survey/ongoing", func(c echo.Context) error { return HandleFunc(c, reportapiService) })
	e.GET("/count/survey/upcoming", func(c echo.Context) error { return HandleFunc(c, reportapiService) })
	e.GET("/count/survey/finished", func(c echo.Context) error { return HandleFunc(c, reportapiService) })
	e.POST("/save/log/activities", func(c echo.Context) error { return HandleFunc(c, reportapiService) })
	e.GET("/log/activities", func(c echo.Context) error { return HandleFunc(c, reportapiService) })
	e.GET("/survey/activities", func(c echo.Context) error { return HandleFunc(c, reportapiService) })
	e.GET("/survey/activities/export", func(c echo.Context) error { return HandleFunc(c, reportapiService) })
	e.GET("/log/surveys", func(c echo.Context) error { return HandleFunc(c, reportapiService) })
	e.POST("/log/surveys", func(c echo.Context) error { return HandleFunc(c, reportapiService) })

	// DOCAPI SERVICE
	e.GET("/view-survey-image/:id", func(c echo.Context) error { return HandleFunc(c, docapiService) })
	e.GET("/view-cms-image/:path", func(c echo.Context) error { return HandleFunc(c, docapiService) })
	e.GET("/view-laporan-konten-image/:path", func(c echo.Context) error { return HandleFunc(c, docapiService) })
	e.GET("/view-foto-profil/:path", func(c echo.Context) error { return HandleFunc(c, docapiService) })

	// MASTERAPI Service
	e.GET("/masterapi/healthy", func(c echo.Context) error { return HandleFunc(c, masterapiService) })
	manajemenWilayahGroup := e.Group("/manajemen-wilayah")
	manajemenWilayahKecamatanGroup := manajemenWilayahGroup.Group("/kecamatan")
	manajemenWilayahKecamatanGroup.POST("/create", func(c echo.Context) error { return HandleFunc(c, masterapiService) })
	manajemenWilayahKecamatanGroup.GET("/list", func(c echo.Context) error { return HandleFunc(c, masterapiService) })
	manajemenWilayahKecamatanGroup.GET("/detail/:kecamatan_id", func(c echo.Context) error { return HandleFunc(c, masterapiService) })
	manajemenWilayahKecamatanGroup.PUT("/update/:kecamatan_id", func(c echo.Context) error { return HandleFunc(c, masterapiService) })
	manajemenWilayahGroup.GET("/kecamatan/options", func(c echo.Context) error { return HandleFunc(c, masterapiService) })
	manajemenWilayahKecamatanGroup.DELETE("/delete/:kecamatan_id", func(c echo.Context) error { return HandleFunc(c, masterapiService) })

	manajemenWilayahKelurahanGroup := manajemenWilayahGroup.Group("/kelurahan")
	manajemenWilayahKelurahanGroup.POST("/create", func(c echo.Context) error { return HandleFunc(c, masterapiService) })
	manajemenWilayahKelurahanGroup.GET("/detail/:kelurahan_id", func(c echo.Context) error { return HandleFunc(c, masterapiService) })
	manajemenWilayahKelurahanGroup.PUT("/update/:kelurahan_id", func(c echo.Context) error { return HandleFunc(c, masterapiService) })
	manajemenWilayahKelurahanGroup.GET("/list", func(c echo.Context) error { return HandleFunc(c, masterapiService) })
	manajemenWilayahKelurahanGroup.GET("/list/:kecamatan_id", func(c echo.Context) error { return HandleFunc(c, masterapiService) })
	manajemenWilayahKelurahanGroup.GET("/options", func(c echo.Context) error { return HandleFunc(c, masterapiService) })
	manajemenWilayahKelurahanGroup.DELETE("/delete/:kelurahan_id", func(c echo.Context) error { return HandleFunc(c, masterapiService) })

	manajemenWilayahRwGroup := manajemenWilayahGroup.Group("/rw")
	manajemenWilayahRwGroup.GET("/detail/:rw_id", func(c echo.Context) error { return HandleFunc(c, masterapiService) })
	manajemenWilayahRwGroup.PUT("/update/:rw_id", func(c echo.Context) error { return HandleFunc(c, masterapiService) })
	manajemenWilayahRwGroup.GET("/list", func(c echo.Context) error { return HandleFunc(c, masterapiService) })
	manajemenWilayahRwGroup.GET("/list/:kelurahan_id", func(c echo.Context) error { return HandleFunc(c, masterapiService) })
	manajemenWilayahRwGroup.POST("/create", func(c echo.Context) error { return HandleFunc(c, masterapiService) })
	manajemenWilayahGroup.GET("/rw/options", func(c echo.Context) error { return HandleFunc(c, masterapiService) })
	manajemenWilayahRwGroup.DELETE("/delete/:rw_id", func(c echo.Context) error { return HandleFunc(c, masterapiService) })

	manajemenWilayahRtGroup := manajemenWilayahGroup.Group("/rt")
	manajemenWilayahRtGroup.GET("/detail/:rt_id", func(c echo.Context) error { return HandleFunc(c, masterapiService) })
	manajemenWilayahRtGroup.PUT("/update/:rt_id", func(c echo.Context) error { return HandleFunc(c, masterapiService) })
	manajemenWilayahRtGroup.GET("/list", func(c echo.Context) error { return HandleFunc(c, masterapiService) })
	manajemenWilayahRtGroup.GET("/list/:rw_id", func(c echo.Context) error { return HandleFunc(c, masterapiService) })
	manajemenWilayahRtGroup.POST("/create", func(c echo.Context) error { return HandleFunc(c, masterapiService) })
	manajemenWilayahGroup.GET("/rt/options", func(c echo.Context) error { return HandleFunc(c, masterapiService) })
	manajemenWilayahRtGroup.DELETE("/delete/:rt_id", func(c echo.Context) error { return HandleFunc(c, masterapiService) })

	manajemenPejabatGroup := e.Group("/manajemen-pejabat")
	manajemenPejabatGroup.POST("/create", func(c echo.Context) error { return HandleFunc(c, masterapiService) })
	manajemenPejabatGroup.GET("/detail/:id", func(c echo.Context) error { return HandleFunc(c, masterapiService) })
	manajemenPejabatGroup.PUT("/update/:id", func(c echo.Context) error { return HandleFunc(c, masterapiService) })
	manajemenPejabatGroup.DELETE("/delete/:id", func(c echo.Context) error { return HandleFunc(c, masterapiService) })
	manajemenPejabatGroup.GET("/list", func(c echo.Context) error { return HandleFunc(c, masterapiService) })

	manajemenArtikelGroup := e.Group("/manajemen-artikel")
	manajemenArtikelKategoriGroup := manajemenArtikelGroup.Group("/kategori")
	manajemenArtikelKategoriGroup.POST("/create", func(c echo.Context) error { return HandleFunc(c, masterapiService) })
	manajemenArtikelKategoriGroup.PUT("/update/:id", func(c echo.Context) error { return HandleFunc(c, masterapiService) })
	manajemenArtikelKategoriGroup.DELETE("/delete/:id", func(c echo.Context) error { return HandleFunc(c, masterapiService) })
	manajemenArtikelKategoriGroup.GET("/detail/:id", func(c echo.Context) error { return HandleFunc(c, masterapiService) })
	manajemenArtikelKategoriGroup.GET("/list", func(c echo.Context) error { return HandleFunc(c, masterapiService) })

	e.GET("/tabel-data-kota-bandung", func(c echo.Context) error { return HandleFunc(c, masterapiService) })

	// SURVEYAPI Service
	e.GET("/surveyapi/healthy", func(c echo.Context) error { return HandleFunc(c, surveyapiService) })

	templateGroup := e.Group("/template")

	templateUcapanGroup := templateGroup.Group("/ucapan")
	templateUcapanGroup.GET("/detail/:template_ucapan_id", func(c echo.Context) error { return HandleFunc(c, surveyapiService) })
	templateUcapanGroup.GET("/variable-options", func(c echo.Context) error { return HandleFunc(c, surveyapiService) })
	templateUcapanGroup.POST("/create", func(c echo.Context) error { return HandleFunc(c, surveyapiService) })
	templateUcapanGroup.PUT("/update/:template_ucapan_id", func(c echo.Context) error { return HandleFunc(c, surveyapiService) })
	templateUcapanGroup.DELETE("/delete/:template_ucapan_id", func(c echo.Context) error { return HandleFunc(c, surveyapiService) })
	templateUcapanGroup.GET("/list", func(c echo.Context) error { return HandleFunc(c, surveyapiService) })
	templateUcapanGroup.GET("/options", func(c echo.Context) error { return HandleFunc(c, surveyapiService) })

	templateFormulirPertanyaanGroup := templateGroup.Group("/formulir-pertanyaan")
	templateFormulirPertanyaanGroup.POST("/create", func(c echo.Context) error { return HandleFunc(c, surveyapiService) })
	templateFormulirPertanyaanGroup.GET("/detail/:template_formulir_pertanyaan_code", func(c echo.Context) error { return HandleFunc(c, surveyapiService) })
	templateFormulirPertanyaanGroup.PUT("/update/:template_formulir_pertanyaan_code", func(c echo.Context) error { return HandleFunc(c, surveyapiService) })
	templateFormulirPertanyaanGroup.POST("/duplicate/:template_formulir_pertanyaan_code", func(c echo.Context) error { return HandleFunc(c, surveyapiService) })
	templateFormulirPertanyaanGroup.DELETE("/delete/:template_formulir_pertanyaan_code", func(c echo.Context) error { return HandleFunc(c, surveyapiService) })
	templateFormulirPertanyaanGroup.GET("/list", func(c echo.Context) error { return HandleFunc(c, surveyapiService) })
	templateFormulirPertanyaanGroup.GET("/question-type-options", func(c echo.Context) error { return HandleFunc(c, surveyapiService) })
	templateFormulirPertanyaanGroup.GET("/options", func(c echo.Context) error { return HandleFunc(c, surveyapiService) })
	templateFormulirPertanyaanGroup.GET("/pertanyaan-options/:form_code", func(c echo.Context) error { return HandleFunc(c, surveyapiService) })
	templateFormulirPertanyaanGroup.GET("/detail-pertanyaan/:id", func(c echo.Context) error { return HandleFunc(c, surveyapiService) })
	templateFormulirPertanyaanGroup.GET("/multiple-choice-options/:form_field_id", func(c echo.Context) error { return HandleFunc(c, surveyapiService) })

	manajemenAlurGroup := e.Group("/manajemen-alur")
	manajemenAlurGroup.POST("/create", func(c echo.Context) error { return HandleFunc(c, surveyapiService) })
	manajemenAlurGroup.GET("/detail/:code", func(c echo.Context) error { return HandleFunc(c, surveyapiService) })
	manajemenAlurGroup.PUT("/update/:code", func(c echo.Context) error { return HandleFunc(c, surveyapiService) })
	manajemenAlurGroup.DELETE("/delete/:code", func(c echo.Context) error { return HandleFunc(c, surveyapiService) })
	manajemenAlurGroup.GET("/list", func(c echo.Context) error { return HandleFunc(c, surveyapiService) })
	manajemenAlurGroup.GET("/preview-index/:code", func(c echo.Context) error { return HandleFunc(c, surveyapiService) })
	manajemenAlurGroup.GET("/preview-alur/:flow_code/:section_code", func(c echo.Context) error { return HandleFunc(c, surveyapiService) })
	manajemenAlurGroup.GET("/options", func(c echo.Context) error { return HandleFunc(c, surveyapiService) })

	surveyGroup := e.Group("/survey")
	surveyGroup.POST("/create", func(c echo.Context) error { return HandleFunc(c, surveyapiService) })
	surveyGroup.GET("/detail/:survey_code", func(c echo.Context) error { return HandleFunc(c, surveyapiService) })
	surveyGroup.GET("/periode-options", func(c echo.Context) error { return HandleFunc(c, surveyapiService) })
	surveyGroup.GET("/list", func(c echo.Context) error { return HandleFunc(c, surveyapiService) })
	surveyGroup.POST("/approval/:code", func(c echo.Context) error { return HandleFunc(c, surveyapiService) })
	surveyGroup.GET("/history-approval/survey/:survey_code", func(c echo.Context) error { return HandleFunc(c, surveyapiService) })
	surveyGroup.GET("/show-image/:survey_code/:code_wilayah/:path", func(c echo.Context) error { return HandleFunc(c, surveyapiService) })
	surveyGroup.GET("/options", func(c echo.Context) error { return HandleFunc(c, surveyapiService) })
	surveyGroup.GET("/question-options/:survey_id", func(c echo.Context) error { return HandleFunc(c, surveyapiService) })

	surveyWilayahGroup := e.Group("/survey-wilayah")
	surveyWilayahGroup.GET("/list", func(c echo.Context) error { return HandleFunc(c, surveyapiService) })
	surveyWilayahGroup.GET("/preview-index/:code", func(c echo.Context) error { return HandleFunc(c, surveyapiService) })
	surveyWilayahGroup.GET("/preview/:survey_code/:section_code", func(c echo.Context) error { return HandleFunc(c, surveyapiService) })
	surveyWilayahGroup.POST("/:survey_code/section/:section_code/submit", func(c echo.Context) error { return HandleFunc(c, surveyapiService) })
	surveyWilayahGroup.PUT("/update-status/:survey_code", func(c echo.Context) error { return HandleFunc(c, surveyapiService) })
	surveyWilayahGroup.GET("/action-required/count", func(c echo.Context) error { return HandleFunc(c, surveyapiService) })

	surveyKewilayahanAktifitasGroup := e.Group("/survey-kewilayahan-aktifitas")
	surveyKewilayahanAktifitasGroup.GET("/list", func(c echo.Context) error { return HandleFunc(c, surveyapiService) })
	surveyKewilayahanAktifitasGroup.GET("/detail/:survey_code", func(c echo.Context) error { return HandleFunc(c, surveyapiService) })

	surveyKewilayahanGroup := e.Group("/survey-kewilayahan")
	surveyKewilayahanGroup.GET("/list", func(c echo.Context) error { return HandleFunc(c, surveyapiService) })
	surveyKewilayahanGroup.GET("/detail/:survey_code", func(c echo.Context) error { return HandleFunc(c, surveyapiService) })
	surveyKewilayahanGroup.GET("/survey/:survey_code/result-index/:code", func(c echo.Context) error { return HandleFunc(c, surveyapiService) })
	surveyKewilayahanGroup.GET("/survey/:survey_code/result-index/:code/section/:section_code", func(c echo.Context) error { return HandleFunc(c, surveyapiService) })
	surveyKewilayahanGroup.PUT("/survey/:survey_code/result-index/:code/approve-verify", func(c echo.Context) error { return HandleFunc(c, surveyapiService) })
	surveyKewilayahanGroup.PUT("/survey/:survey_code/result-index/:code/reject-verify", func(c echo.Context) error { return HandleFunc(c, surveyapiService) })
	surveyKewilayahanGroup.PUT("/survey/:survey_code/result-index/:code/reject-validate", func(c echo.Context) error { return HandleFunc(c, surveyapiService) })
	surveyKewilayahanGroup.PUT("/survey/:survey_code/result-index/:code/approve-validate", func(c echo.Context) error { return HandleFunc(c, surveyapiService) })
	surveyKewilayahanGroup.GET("/approval-history/survey/:survey_code/wilayah/:code_wilayah", func(c echo.Context) error { return HandleFunc(c, surveyapiService) })
	surveyKewilayahanGroup.GET("/history-detail/survey/:survey_code/wilayah/:code_wilayah", func(c echo.Context) error { return HandleFunc(c, surveyapiService) })

	surveyKewilayahanGroup.GET("/export-excel/survey/:survey_code/wilayah/:code_wilayah", func(c echo.Context) error { return HandleFunc(c, surveyapiService) })
	surveyKewilayahanGroup.GET("/export-excel/survey/:survey_code", func(c echo.Context) error { return HandleFunc(c, surveyapiService) })

	surveyKewilayahanGroup.PUT("/reset-status-survey/:survey_code/wilayah/:code_wilayah", func(c echo.Context) error { return HandleFunc(c, surveyapiService) })

	surveyKewilayahanGroup.GET("/respondent-rejected-all", func(c echo.Context) error { return HandleFunc(c, surveyapiService) })
	surveyKewilayahanGroup.GET("/respondent-rejected-survey/:survey_code", func(c echo.Context) error { return HandleFunc(c, surveyapiService) })

	// BEGIN::MONITORING & LAPORAN ==========================
	monitoringDanLaporanGroup := e.Group("/monitoring-dan-laporan")
	// BEGIN::STATISTIK ==========================
	statistikGroup := monitoringDanLaporanGroup.Group("/statistik")
	statistikGroup.GET("", func(c echo.Context) error { return HandleFunc(c, reportapiService) })
	statistikGroup.GET("/survey/:survey_code", func(c echo.Context) error { return HandleFunc(c, reportapiService) })
	statistikGroup.GET("/survey/:survey_code/export-excel", func(c echo.Context) error { return HandleFunc(c, reportapiService) })
	// END::STATISTIK ==========================

	// BEGIN::LAPORAN ==========================
	laporanGroup := monitoringDanLaporanGroup.Group("/laporan")
	laporanGroup.GET("/list", func(c echo.Context) error { return HandleFunc(c, reportapiService) })
	laporanGroup.PATCH("/change-name/:laporan_id", func(c echo.Context) error { return HandleFunc(c, reportapiService) })
	laporanGroup.POST("/create", func(c echo.Context) error { return HandleFunc(c, reportapiService) })
	laporanGroup.PUT("/update-cover/:laporan_id", func(c echo.Context) error { return HandleFunc(c, reportapiService) })
	laporanGroup.GET("/get-cover/:laporan_id", func(c echo.Context) error { return HandleFunc(c, reportapiService) })
	laporanGroup.GET("/cetak/:laporan_id", func(c echo.Context) error { return HandleFunc(c, pyReportApiService) })

	laporanGroup.POST("/:laporan_id/section/create", func(c echo.Context) error { return HandleFunc(c, reportapiService) })
	laporanGroup.GET("/calculation-type-options", func(c echo.Context) error { return HandleFunc(c, reportapiService) })
	// END::LAPORAN ==========================
	// END::MONITORING & LAPORAN ==========================

	// BEGIN::PENGATURAN APLIKASI ==========================
	pengaturanAplikasi := e.Group("/pengaturan-aplikasi")
	// BEGIN::MANAJEMEN CMS ==========================
	manajemenCMSGroup := pengaturanAplikasi.Group("/manajemen-cms")
	manajemenCMSGroup.GET("/list-section", func(c echo.Context) error { return HandleFunc(c, masterapiService) })
	manajemenCMSGroup.PATCH("/update-status-section/:id", func(c echo.Context) error { return HandleFunc(c, masterapiService) })
	manajemenCMSGroup.PATCH("/update-name-section/:id", func(c echo.Context) error { return HandleFunc(c, masterapiService) })
	manajemenCMSGroup.DELETE("/delete-section/:id", func(c echo.Context) error { return HandleFunc(c, masterapiService) })
	manajemenCMSGroup.POST("/create-section", func(c echo.Context) error { return HandleFunc(c, masterapiService) })

	manajemenCMSGroup.GET("/section/:slug", func(c echo.Context) error { return HandleFunc(c, masterapiService) })
	manajemenCMSGroup.PATCH("/update-section/:slug", func(c echo.Context) error { return HandleFunc(c, masterapiService) })
	manajemenCMSGroup.PATCH("/reorder-section/:slug", func(c echo.Context) error { return HandleFunc(c, masterapiService) })
	// END::MANAJEMEN CMS ==========================
	// END::PENGATURAN APLIKASI ==========================
}
