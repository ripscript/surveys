package routingGrpc

import (
	"backend/masterapi/configs"
	"backend/masterapi/handlers"
	"backend/masterapi/models"
	"backend/masterapi/repository"
	"backend/masterapi/service"
	"backend/masterapi/utils"
	pb "backend/siccore/pb"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"google.golang.org/grpc/metadata"
	"gorm.io/gorm"
)

type GRPCServer struct {
	pb.UnimplementedProxyServer
}

var (
	dbMaster *gorm.DB = configs.SetupDatabaseMasterConnection()
	dbSlave  *gorm.DB = configs.SetupDatabaseSlaveConnection()
)

var (
	manajemenWilayahRepo  repository.ManajemenWilayahRepo      = repository.NewManajemenWilayahRepo(dbSlave, dbMaster)
	manajemenPenggunaRepo repository.ManajemenPenggunaRepo     = repository.NewManajemenPenggunaRepo(dbSlave, dbMaster)
	manajemenArtikelRepo  repository.ManajemenArtikelRepo      = repository.NewManajemenArtikelRepo(dbSlave, dbMaster)
	manajemenPejabatRepo  repository.ManajemenPejabatRepo      = repository.NewManajemenPejabatRepo(dbSlave, dbMaster)
	manajamenCMSRepo      repository.ManajemenCMSRepo          = repository.NewManajemenCMSRepo(dbSlave, dbMaster)
	fileRepo              repository.FileRepo                  = repository.NewFileRepo(dbSlave, dbMaster)
	artikelCategoryRepo   repository.ArtikelCategoryRepository = repository.NewArtikelCategoryRepository(dbSlave, dbMaster)
	artikelRepo           repository.ArtikelRepository         = repository.NewArtikelRepository(dbSlave, dbMaster)
	artikelPromoteRepo    repository.ArtikelPromoteRepository  = repository.NewArtikelPromoteRepository(dbSlave, dbMaster)
)

var (
	manajemenWilayahService service.ManajemenWilayahService = service.NewManajemenWilayahService(
		manajemenWilayahRepo,
	)
	manajemenPenggunaService service.ManajemenPenggunaService = service.NewManajemenPenggunaService(
		manajemenPenggunaRepo,
	)
	manajemenArtikelService service.ManajemenArtikelService = service.NewManajemenArtikelService(
		manajemenArtikelRepo,
	)
	manajemenPejabatService service.ManajemenPejabatService = service.NewManajemenPejabatService(
		manajemenPejabatRepo,
	)
	manajemenCMSService service.ManajemenCMSService = service.NewManajemenCMSService(
		manajamenCMSRepo,
		fileRepo,
	)
	artikelCategoryService service.ArtikelCategoryService = service.NewArtikelCategoryService(
		artikelCategoryRepo,
	)
	artikelService service.ArtikelService = service.NewArtikelService(
		fileRepo,
		artikelRepo,
		artikelCategoryRepo,
	)
	artikelPromoteService service.ArtikelPromoteService = service.NewArtikelPromoteService(
		artikelRepo,
		artikelPromoteRepo,
	)
)

var (
	manajemenWilayahHandler handlers.ManajemenWilayahHandler = handlers.NewManajemenWilayahHandler(
		manajemenWilayahService,
	)
	manajemenPenggunaHandler handlers.ManajemenPenggunaHandler = handlers.NewManajemenPenggunaHandler(
		manajemenPenggunaService,
	)
	manajemenArtikelHandler handlers.ManajemenArtikelHandler = handlers.NewManajemenArtikelHandler(
		manajemenArtikelService,
	)
	manajemenPejabatHandler handlers.ManajemenPejabatHandler = handlers.NewManajemenPejabatHandler(
		manajemenPejabatService,
	)
	manajemenCMSHandler handlers.ManajemenCMSHandler = handlers.NewManajemenCMSHandler(
		manajemenCMSService,
	)
	artikelCategoryHandler handlers.ArtikelCategoryHandler = handlers.NewArtikelCategoryHandler(
		artikelCategoryService,
	)
	artikelHandler handlers.ArtikelHandler = handlers.NewArtikelHandler(
		artikelService,
	)
	artikelPromoteHandler handlers.ArtikelPromoteHandler = handlers.NewArtikelPromoteHandler(
		artikelPromoteService,
	)
)

// && Key Menu && \\
// template
// management-alur
// survey
// dashboard
// pengaturan
// laporan
// monitoring
// admin
// formulir-pertanyaan
// ucapan
// master-data
// hasil
// management-pengguna
// management-wilayah
// management-cms
// management-artikel
// rating
// statistik
// aktifitas-survey
// profil-saya
// keluar
// management-responden
// user
// management-blokir
// management-wilayah-child
// management-pejabat
// artikel
// promote
// kategori

// ROUTING GRPC
type RouteConfig struct {
	Handler func(context.Context, map[string]interface{}, models.JwtCustomClaims, url.Values, map[string]interface{}) (*pb.ProxyResponse, error)
	MenuKey string
}

// Definisikan pemetaan fungsi handler dengan path dan metode HTTP
var grpcMap = map[string]map[string]RouteConfig{

	"/masterapi/healthy": {"GET": {Handler: handlers.Healthy, MenuKey: ""}},

	"/manajemen-wilayah/kecamatan/create":               {"POST": {Handler: manajemenWilayahHandler.CreateKecamatan, MenuKey: "management-wilayah"}},
	"/manajemen-wilayah/kecamatan/list":                 {"GET": {Handler: manajemenWilayahHandler.GetListKecamatan, MenuKey: "management-wilayah"}},
	"/manajemen-wilayah/kecamatan/detail/:kecamatan_id": {"GET": {Handler: manajemenWilayahHandler.GetKecamatanDetail, MenuKey: "management-wilayah"}},
	"/manajemen-wilayah/kecamatan/update/:kecamatan_id": {"PUT": {Handler: manajemenWilayahHandler.UpdateKecamatan, MenuKey: "management-wilayah"}},
	"/manajemen-wilayah/kecamatan/options":              {"GET": {Handler: manajemenWilayahHandler.OptionsKecamatan, MenuKey: "management-wilayah"}},
	"/manajemen-wilayah/kecamatan/delete/:kecamatan_id": {"DELETE": {Handler: manajemenWilayahHandler.DeleteKecamatan, MenuKey: "management-wilayah"}},

	"/manajemen-wilayah/kelurahan/create":               {"POST": {Handler: manajemenWilayahHandler.CreateKelurahan, MenuKey: "management-wilayah"}},
	"/manajemen-wilayah/kelurahan/detail/:kelurahan_id": {"GET": {Handler: manajemenWilayahHandler.GetKelurahanDetail, MenuKey: "management-wilayah"}},
	"/manajemen-wilayah/kelurahan/update/:kelurahan_id": {"PUT": {Handler: manajemenWilayahHandler.UpdateKelurahan, MenuKey: "management-wilayah"}},
	"/manajemen-wilayah/kelurahan/list":                 {"GET": {Handler: manajemenWilayahHandler.GetListKelurahan, MenuKey: "management-wilayah"}},
	"/manajemen-wilayah/kelurahan/list/:kecamatan_id":   {"GET": {Handler: manajemenWilayahHandler.GetListKelurahan, MenuKey: "management-wilayah"}},
	"/manajemen-wilayah/kelurahan/options":              {"GET": {Handler: manajemenWilayahHandler.OptionsKelurahan, MenuKey: "management-wilayah"}},
	"/manajemen-wilayah/kelurahan/delete/:kelurahan_id": {"DELETE": {Handler: manajemenWilayahHandler.DeleteKelurahan, MenuKey: "management-wilayah"}},

	"/manajemen-wilayah/rw/detail/:rw_id":      {"GET": {Handler: manajemenWilayahHandler.GetRwDetail, MenuKey: "management-wilayah"}},
	"/manajemen-wilayah/rw/update/:rw_id":      {"PUT": {Handler: manajemenWilayahHandler.UpdateRw, MenuKey: "management-wilayah"}},
	"/manajemen-wilayah/rw/list":               {"GET": {Handler: manajemenWilayahHandler.GetListRw, MenuKey: "management-wilayah"}},
	"/manajemen-wilayah/rw/list/:kelurahan_id": {"GET": {Handler: manajemenWilayahHandler.GetListRwByKelurahan, MenuKey: "management-wilayah"}},
	"/manajemen-wilayah/rw/create":             {"POST": {Handler: manajemenWilayahHandler.CreateRw, MenuKey: "management-wilayah"}},
	"/manajemen-wilayah/rw/options":            {"GET": {Handler: manajemenWilayahHandler.OptionsRw, MenuKey: "management-wilayah"}},
	"/manajemen-wilayah/rw/delete/:rw_id":      {"DELETE": {Handler: manajemenWilayahHandler.DeleteRw, MenuKey: "management-wilayah"}},

	"/manajemen-wilayah/rt/detail/:rt_id": {"GET": {Handler: manajemenWilayahHandler.GetRtDetail, MenuKey: "management-wilayah"}},
	"/manajemen-wilayah/rt/update/:rt_id": {"PUT": {Handler: manajemenWilayahHandler.UpdateRt, MenuKey: "management-wilayah"}},
	"/manajemen-wilayah/rt/list":          {"GET": {Handler: manajemenWilayahHandler.GetListRt, MenuKey: "management-wilayah"}},
	"/manajemen-wilayah/rt/list/:rw_id":   {"GET": {Handler: manajemenWilayahHandler.GetListRtByRw, MenuKey: "management-wilayah"}},
	"/manajemen-wilayah/rt/create":        {"POST": {Handler: manajemenWilayahHandler.CreateRt, MenuKey: "management-wilayah"}},
	"/manajemen-wilayah/rt/options":       {"GET": {Handler: manajemenWilayahHandler.OptionsRt, MenuKey: "management-wilayah"}},
	"/manajemen-wilayah/rt/delete/:rt_id": {"DELETE": {Handler: manajemenWilayahHandler.DeleteRT, MenuKey: "management-wilayah"}},

	"/manajemen-pejabat/create": {"POST": {
		Handler: manajemenPejabatHandler.CreatePejabat,
		MenuKey: "management-pejabat",
	}},
	"/manajemen-pejabat/detail/:id": {"GET": {
		Handler: manajemenPejabatHandler.DetailPejabat,
		MenuKey: "management-pejabat",
	}},
	"/manajemen-pejabat/update/:id": {"PUT": {
		Handler: manajemenPejabatHandler.UpdatePejabat,
		MenuKey: "management-pejabat",
	}},
	"/manajemen-pejabat/delete/:id": {"DELETE": {
		Handler: manajemenPejabatHandler.DeletePejabat,
		MenuKey: "belum-digunakan",
	}},
	"/manajemen-pejabat/list": {"GET": {
		Handler: manajemenPejabatHandler.GetListPejabat,
		MenuKey: "management-pejabat",
	}},

	"/manajemen-artikel/kategori/create": {"POST": {
		Handler: manajemenArtikelHandler.CreateKategoriArtikel,
		MenuKey: "",
	}},
	"/manajemen-artikel/kategori/update/:id": {"PUT": {
		Handler: manajemenArtikelHandler.UpdateKategoriArtikel,
		MenuKey: "",
	}},
	"/manajemen-artikel/kategori/delete/:id": {"DELETE": {
		Handler: manajemenArtikelHandler.DeleteKategoriArtikel,
		MenuKey: "",
	}},
	"/manajemen-artikel/kategori/detail/:id": {"GET": {
		Handler: manajemenArtikelHandler.GetKategoriArtikel,
		MenuKey: "",
	}},
	"/manajemen-artikel/kategori/list": {"GET": {
		Handler: manajemenArtikelHandler.GetListKategoriArtikel,
		MenuKey: "",
	}},

	"/tabel-data-kota-bandung": {"GET": {
		Handler: manajemenWilayahHandler.TabelDataKotaBandung,
		MenuKey: "management-wilayah",
	}},

	// BEGIN::PENGATURAN APLIKASI ===============================
	// BEGIN::MANAJEMEN CMS ===============================
	"/pengaturan-aplikasi/manajemen-cms/list-section": {
		"GET": {
			Handler: manajemenCMSHandler.ListSection,
			MenuKey: "manajemen-cms",
		},
	},
	"/pengaturan-aplikasi/manajemen-cms/update-status-section/:id": {
		"PATCH": {
			Handler: manajemenCMSHandler.UpdateStatusSection,
			MenuKey: "manajemen-cms",
		},
	},
	"/pengaturan-aplikasi/manajemen-cms/update-name-section/:id": {
		"PATCH": {
			Handler: manajemenCMSHandler.UpdateNameSection,
			MenuKey: "manajemen-cms",
		},
	},
	"/pengaturan-aplikasi/manajemen-cms/delete-section/:id": {
		"DELETE": {
			Handler: manajemenCMSHandler.DeleteSection,
			MenuKey: "manajemen-cms",
		},
	},
	"/pengaturan-aplikasi/manajemen-cms/create-section": {
		"POST": {
			Handler: manajemenCMSHandler.CreateSection,
			MenuKey: "manajemen-cms",
		},
	},

	"/pengaturan-aplikasi/manajemen-cms/section/:slug": {
		"GET": {
			Handler: manajemenCMSHandler.GetSectionBySlug,
			MenuKey: "manajemen-cms",
		},
	},
	"/pengaturan-aplikasi/manajemen-cms/update-section/:slug": {
		"PATCH": {
			Handler: manajemenCMSHandler.UpdateSectionBySlug,
			MenuKey: "manajemen-cms",
		},
	},
	"/pengaturan-aplikasi/manajemen-cms/reorder-section/:slug": {
		"PATCH": {
			Handler: manajemenCMSHandler.UpdateOrderSection,
			MenuKey: "manajemen-cms",
		},
	},

	"/pengaturan-aplikasi/manajemen-cms/landing-page": {
		"GET": {
			Handler: manajemenCMSHandler.GetLandingPage,
			MenuKey: "",
		},
	},

	"/pengaturan-aplikasi/geojson-kota-bandung-level-kecamatan": {
		"GET": {
			Handler: manajemenCMSHandler.GeoJsonKotaBandungLevelKecamatan,
			MenuKey: "",
		},
	},
	// END::MANAJEMEN CMS ===============================

	// BEGIN::MANAJEMEN ARTIKEL ===============================

	// BEGIN::ARTIKEL ===============================
	"/pengaturan-aplikasi/manajemen-artikel/artikel/create": {
		"POST": {
			Handler: artikelHandler.Create,
			MenuKey: "artikel",
		},
	},
	"/pengaturan-aplikasi/manajemen-artikel/artikel/detail/:id": {
		"GET": {
			Handler: artikelHandler.Detail,
			MenuKey: "artikel",
		},
	},
	"/pengaturan-aplikasi/manajemen-artikel/artikel/detail-public/:id": {
		"GET": {
			Handler: artikelHandler.DetailPublic,
			MenuKey: "artikel",
		},
	},
	"/pengaturan-aplikasi/manajemen-artikel/artikel/update/:id": {
		"PUT": {
			Handler: artikelHandler.Update,
			MenuKey: "artikel",
		},
	},
	"/pengaturan-aplikasi/manajemen-artikel/artikel/delete/:id": {
		"DELETE": {
			Handler: artikelHandler.Delete,
			MenuKey: "artikel",
		},
	},
	"/pengaturan-aplikasi/manajemen-artikel/artikel/get-options": {
		"GET": {
			Handler: artikelHandler.GetOptions,
			MenuKey: "artikel",
		},
	},
	"/pengaturan-aplikasi/manajemen-artikel/artikel/list": {
		"GET": {
			Handler: artikelHandler.List,
			MenuKey: "artikel",
		},
	},
	"/pengaturan-aplikasi/manajemen-artikel/artikel/public-list": {
		"GET": {
			Handler: artikelHandler.PublicList,
			MenuKey: "artikel",
		},
	},
	// END::ARTIKEL ===============================

	// BEGIN::ARTIKEL KATEGORI ===============================
	"/pengaturan-aplikasi/manajemen-artikel/kategori/create": {
		"POST": {
			Handler: artikelCategoryHandler.Create,
			MenuKey: "kategori",
		},
	},
	"/pengaturan-aplikasi/manajemen-artikel/kategori/delete/:id": {
		"DELETE": {
			Handler: artikelCategoryHandler.Delete,
			MenuKey: "kategori",
		},
	},
	"/pengaturan-aplikasi/manajemen-artikel/kategori/detail/:id": {
		"GET": {
			Handler: artikelCategoryHandler.Detail,
			MenuKey: "kategori",
		},
	},
	"/pengaturan-aplikasi/manajemen-artikel/kategori/update/:id": {
		"PUT": {
			Handler: artikelCategoryHandler.Update,
			MenuKey: "kategori",
		},
	},
	"/pengaturan-aplikasi/manajemen-artikel/kategori/get-options": {
		"GET": {
			Handler: artikelCategoryHandler.GetOptions,
			MenuKey: "kategori",
		},
	},
	"/pengaturan-aplikasi/manajemen-artikel/kategori/list": {
		"GET": {
			Handler: artikelCategoryHandler.List,
			MenuKey: "kategori",
		},
	},
	"/pengaturan-aplikasi/manajemen-artikel/kategori/public-list": {
		"GET": {
			Handler: artikelCategoryHandler.PublicList,
			MenuKey: "kategori",
		},
	},
	// BEGIN::ARTIKEL KATEGORI ===============================

	// BEGIN::ARTIKEL PROMOTE ===============================
	"/pengaturan-aplikasi/manajemen-artikel/promote/create": {
		"POST": {
			Handler: artikelPromoteHandler.Create,
			MenuKey: "promote",
		},
	},
	"/pengaturan-aplikasi/manajemen-artikel/promote/delete/:id": {
		"DELETE": {
			Handler: artikelPromoteHandler.Delete,
			MenuKey: "promote",
		},
	},
	"/pengaturan-aplikasi/manajemen-artikel/promote/update-status/:id": {
		"PATCH": {
			Handler: artikelPromoteHandler.UpdateStatus,
			MenuKey: "promote",
		},
	},
	"/pengaturan-aplikasi/manajemen-artikel/promote/list": {
		"GET": {
			Handler: artikelPromoteHandler.List,
			MenuKey: "promote",
		},
	},
	// END::ARTIKEL PROMOTE ===============================

	// END::MANAJEMEN ARTIKEL ===============================
	// END::PENGATURAN APLIKASI ===============================
}

func (s *GRPCServer) SendData(ctx context.Context, req *pb.ProxyRequest) (*pb.ProxyResponse, error) {
	defer func() {
		if r := recover(); r != nil {
			message := fmt.Sprintf("Terjadi kendala pada service yang sedang anda akses: %v", r)
			utils.LogErrors(message)
		}
	}()

	path := req.GetPath()
	method := req.GetMethod()

	success, mess, code, userLogin, newToken := ValidasiToken(ctx, req)
	if !success {
		utils.LogErrors(mess)
		return utils.SetResponseData([]byte{}, success, mess, code, nil, ""), nil
	}

	methodMap, ok := grpcMap[path]
	if !ok {
		message := "Path grpc tidak ditemukan"
		return utils.SetResponseData([]byte{}, false, message, http.StatusNotFound, nil, ""), nil
	}

	routeConfig, ok := methodMap[method]
	if !ok {
		message := "Method grpc tidak ditemukan"
		return utils.SetResponseData([]byte{}, false, message, http.StatusMethodNotAllowed, nil, ""), nil
	}

	handler := routeConfig.Handler
	menuKey := routeConfig.MenuKey

	if req.GetIsSecure() {
		allowed := CheckPermission(userLogin, menuKey, method)
		if !allowed {
			message := "Anda tidak memiliki hak akses"
			return utils.SetResponseData([]byte{}, false, message, http.StatusForbidden, nil, ""), nil
		}
	}

	var reqs map[string]interface{}
	if len(req.Data) > 0 {
		if err := json.Unmarshal(req.Data, &reqs); err != nil {
			message := "Terjadi kesalahan saat unmarshal request data : " + err.Error()
			return utils.SetResponseData([]byte{}, false, message, http.StatusInternalServerError, nil, ""), nil
		}
	}
	var slug map[string]interface{}
	if len(req.Slug) > 0 {
		if err := json.Unmarshal(req.Slug, &slug); err != nil {
			message := "Terjadi kesalahan saat unmarshal slug : " + err.Error()
			return utils.SetResponseData([]byte{}, false, message, http.StatusInternalServerError, nil, ""), nil
		}
	}

	paramString := string(req.Param)
	queryValues, err := url.ParseQuery(paramString)
	if err != nil {
		message := "Terjadi kesalahan saat parsing query param : " + err.Error()
		return utils.SetResponseData([]byte{}, false, message, http.StatusInternalServerError, nil, ""), nil
	}
	response, err := handler(ctx, reqs, userLogin, queryValues, slug)
	if err != nil {
		return nil, err
	}
	if response != nil && newToken != "" {
		response.Token = newToken
	}

	return response, nil
}

func ValidasiToken(ctx context.Context, req *pb.ProxyRequest) (bool, string, int, models.JwtCustomClaims, string) {
	defer func() {
		if r := recover(); r != nil {
			message := fmt.Sprintf("Terjadi kendala pada service yang sedang anda akses: %v", r)
			utils.LogErrors(message)
		}
	}()

	var withToken bool = true
	var userData models.JwtCustomClaims
	var newToken string

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		message := "Terjadi kesalahan saat mengambil metadata"
		utils.LogErrors(message)
		return false, message, int(http.StatusInternalServerError), userData, newToken
	}

	token := ""
	if val, ok := md["authorization"]; ok {
		parts := strings.Split(val[0], " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			withToken = false
			if req.GetIsSecure() {
				return false, "Format header token tidak valid", int(http.StatusUnauthorized), userData, newToken
			}
		}

		if withToken {
			token = parts[1]
		}
	} else {
		withToken = false
		if req.GetIsSecure() {
			return false, "Header token tidak ditemukan", int(http.StatusUnauthorized), userData, newToken
		}
	}

	if withToken {
		claims := &models.JwtCustomClaims{}
		_, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
			return []byte(os.Getenv("JWT_SECRET_KEY")), nil
		})
		if err != nil {
			if req.GetIsSecure() {
				return false, "Token Tidak Valid", int(http.StatusUnauthorized), userData, newToken
			}
			return true, "Tervalidasi", int(http.StatusOK), userData, newToken
		}

		userData = models.JwtCustomClaims{
			ID:           int64(claims.ID),
			RespondentID: claims.RespondentID,
			Name:         claims.Name,
			Email:        claims.Email,
			Role:         claims.Role,
			Permissions:  claims.Permissions,
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: claims.ExpiresAt,
			},
		}

		if claims.ExpiresAt != nil {
			refreshedToken, err := GenerateJWTToken(userData)
			if err != nil {
				utils.LogErrors("Gagal generate token baru: " + err.Error())
			} else {
				newToken = refreshedToken
			}
		}

		return true, "Tervalidasi", int(http.StatusOK), userData, newToken
	}

	return true, "Tervalidasi", int(http.StatusOK), userData, newToken
}

func GenerateJWTToken(user models.JwtCustomClaims) (string, error) {
	defer utils.GeneralRecover()

	expiredAt := user.ExpiresAt

	if user.ExpiresAt != nil {
		remaining := user.ExpiresAt.Time.Sub(time.Now())
		if remaining < 30*time.Minute {
			expiredAt = jwt.NewNumericDate(time.Now().Add(3 * time.Hour))
		}
	} else {
		expiredAt = jwt.NewNumericDate(time.Now().Add(3 * time.Hour))
	}

	claims := models.JwtCustomClaims{
		RespondentID: user.RespondentID,
		ID:           int64(user.ID),
		Name:         user.Name,
		Email:        user.Email,
		Role:         user.Role,
		Permissions:  user.Permissions,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: expiredAt,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	encryptedToken, err := token.SignedString([]byte(os.Getenv("JWT_SECRET_KEY")))
	if err != nil {
		return "", err
	}
	return encryptedToken, nil
}

func NormalizePath(path string) string {
	re := regexp.MustCompile(`/\d+`)
	return re.ReplaceAllString(path, "/:id")
}

func methodToAction(method string) string {
	switch method {
	case "GET":
		return "view"
	case "POST":
		return "create"
	case "PUT", "PATCH":
		return "update"
	case "DELETE":
		return "delete"
	default:
		return ""
	}
}
func CheckPermission(claims models.JwtCustomClaims, menuKey string, method string) bool {
	action := methodToAction(method)
	if menuKey == "" {
		return true
	}
	perm, ok := claims.Permissions[menuKey]
	if !ok {
		return false
	}

	switch action {
	case "view":
		return perm.V
	case "create":
		return perm.C
	case "update":
		return perm.U
	case "delete":
		return perm.D
	}

	return false
}
