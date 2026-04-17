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
	manajemenWilayahRepo  repository.ManajemenWilayahRepo  = repository.NewManajemenWilayahRepo(dbSlave, dbMaster)
	manajemenPenggunaRepo repository.ManajemenPenggunaRepo = repository.NewManajemenPenggunaRepo(dbSlave, dbMaster)
	manajemenArtikelRepo  repository.ManajemenArtikelRepo  = repository.NewManajemenArtikelRepo(dbSlave, dbMaster)
	manajemenPejabatRepo  repository.ManajemenPejabatRepo  = repository.NewManajemenPejabatRepo(dbSlave, dbMaster)
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
)

type RouteConfig struct {
	Handler func(context.Context, map[string]interface{}, models.JwtCustomClaims, url.Values, map[string]interface{}) (*pb.ProxyResponse, error)
	MenuKey string
}

// ROUTING GRPC
// Definisikan pemetaan fungsi handler dengan path dan metode HTTP
var grpcMap = map[string]map[string]RouteConfig{

	"/masterapi/healthy": {"GET": {Handler: handlers.Healthy, MenuKey: ""}},

	"/manajemen-wilayah/kecamatan/create":               {"POST": {Handler: manajemenWilayahHandler.CreateKecamatan, MenuKey: ""}},
	"/manajemen-wilayah/kecamatan/list":                 {"GET": {Handler: manajemenWilayahHandler.GetListKecamatan, MenuKey: ""}},
	"/manajemen-wilayah/kecamatan/detail/:kecamatan_id": {"GET": {Handler: manajemenWilayahHandler.GetKecamatanDetail, MenuKey: ""}},
	"/manajemen-wilayah/kecamatan/update/:kecamatan_id": {"PUT": {Handler: manajemenWilayahHandler.UpdateKecamatan, MenuKey: ""}},
	"/manajemen-wilayah/kecamatan/options":              {"GET": {Handler: manajemenWilayahHandler.OptionsKecamatan, MenuKey: ""}},
	"/manajemen-wilayah/kecamatan/delete/:kecamatan_id": {"DELETE": {Handler: manajemenWilayahHandler.DeleteKecamatan, MenuKey: ""}},

	"/manajemen-wilayah/kelurahan/create":               {"POST": {Handler: manajemenWilayahHandler.CreateKelurahan, MenuKey: ""}},
	"/manajemen-wilayah/kelurahan/detail/:kelurahan_id": {"GET": {Handler: manajemenWilayahHandler.GetKelurahanDetail, MenuKey: ""}},
	"/manajemen-wilayah/kelurahan/update/:kelurahan_id": {"PUT": {Handler: manajemenWilayahHandler.UpdateKelurahan, MenuKey: ""}},
	"/manajemen-wilayah/kelurahan/list":                 {"GET": {Handler: manajemenWilayahHandler.GetListKelurahan, MenuKey: ""}},
	"/manajemen-wilayah/kelurahan/list/:kecamatan_id":   {"GET": {Handler: manajemenWilayahHandler.GetListKelurahan, MenuKey: ""}},
	"/manajemen-wilayah/kelurahan/options":              {"GET": {Handler: manajemenWilayahHandler.OptionsKelurahan, MenuKey: ""}},
	"/manajemen-wilayah/kelurahan/delete/:kelurahan_id": {"DELETE": {Handler: manajemenWilayahHandler.DeleteKelurahan, MenuKey: ""}},

	"/manajemen-wilayah/rw/detail/:rw_id":      {"GET": {Handler: manajemenWilayahHandler.GetRwDetail, MenuKey: ""}},
	"/manajemen-wilayah/rw/update/:rw_id":      {"PUT": {Handler: manajemenWilayahHandler.UpdateRw, MenuKey: ""}},
	"/manajemen-wilayah/rw/list":               {"GET": {Handler: manajemenWilayahHandler.GetListRw, MenuKey: ""}},
	"/manajemen-wilayah/rw/list/:kelurahan_id": {"GET": {Handler: manajemenWilayahHandler.GetListRwByKelurahan, MenuKey: ""}},
	"/manajemen-wilayah/rw/create":             {"POST": {Handler: manajemenWilayahHandler.CreateRw, MenuKey: ""}},
	"/manajemen-wilayah/rw/options":            {"GET": {Handler: manajemenWilayahHandler.OptionsRw, MenuKey: ""}},
	"/manajemen-wilayah/rw/delete/:rw_id":      {"DELETE": {Handler: manajemenWilayahHandler.DeleteRw, MenuKey: ""}},

	"/manajemen-wilayah/rt/detail/:rt_id": {"GET": {Handler: manajemenWilayahHandler.GetRtDetail, MenuKey: ""}},
	"/manajemen-wilayah/rt/update/:rt_id": {"PUT": {Handler: manajemenWilayahHandler.UpdateRt, MenuKey: ""}},
	"/manajemen-wilayah/rt/list":          {"GET": {Handler: manajemenWilayahHandler.GetListRt, MenuKey: ""}},
	"/manajemen-wilayah/rt/list/:rw_id":   {"GET": {Handler: manajemenWilayahHandler.GetListRtByRw, MenuKey: ""}},
	"/manajemen-wilayah/rt/create":        {"POST": {Handler: manajemenWilayahHandler.CreateRt, MenuKey: ""}},
	"/manajemen-wilayah/rt/options":       {"GET": {Handler: manajemenWilayahHandler.OptionsRt, MenuKey: ""}},
	"/manajemen-wilayah/rt/delete/:rt_id": {"DELETE": {Handler: manajemenWilayahHandler.DeleteRT, MenuKey: ""}},

	"/manajemen-pejabat/create":     {"POST": {Handler: manajemenPejabatHandler.CreatePejabat, MenuKey: ""}},
	"/manajemen-pejabat/detail/:id": {"GET": {Handler: manajemenPejabatHandler.DetailPejabat, MenuKey: ""}},
	"/manajemen-pejabat/update/:id": {"PUT": {Handler: manajemenPejabatHandler.UpdatePejabat, MenuKey: ""}},
	"/manajemen-pejabat/delete/:id": {"DELETE": {Handler: manajemenPejabatHandler.DeletePejabat, MenuKey: ""}},
	"/manajemen-pejabat/list":       {"GET": {Handler: manajemenPejabatHandler.GetListPejabat, MenuKey: ""}},

	"/manajemen-pengguna/role/options":     {"GET": {Handler: manajemenPenggunaHandler.OptionsRole, MenuKey: ""}},
	"/manajemen-pengguna/responden/create": {"POST": {Handler: manajemenPenggunaHandler.CreateResponden, MenuKey: ""}},

	"/manajemen-artikel/kategori/create":     {"POST": {Handler: manajemenArtikelHandler.CreateKategoriArtikel, MenuKey: ""}},
	"/manajemen-artikel/kategori/update/:id": {"PUT": {Handler: manajemenArtikelHandler.UpdateKategoriArtikel, MenuKey: ""}},
	"/manajemen-artikel/kategori/delete/:id": {"DELETE": {Handler: manajemenArtikelHandler.DeleteKategoriArtikel, MenuKey: ""}},
	"/manajemen-artikel/kategori/detail/:id": {"GET": {Handler: manajemenArtikelHandler.GetKategoriArtikel, MenuKey: ""}},
	"/manajemen-artikel/kategori/list":       {"GET": {Handler: manajemenArtikelHandler.GetListKategoriArtikel, MenuKey: ""}},
}

// Metode untuk menangani permintaan yang masuk
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
		allowed := CheckPermission(int(userLogin.Role), menuKey, method)
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
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: expiredAt,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	encryptedToken, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
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
func CheckPermission(roleID int, menuKey string, method string) bool {
	if menuKey == "" {
		return true
	}

	action := methodToAction(method)
	if action == "" {
		return false
	}

	var count int64

	query := dbSlave.Table("menu_permissions mp").
		Joins("JOIN menus m ON m.id = mp.menu_id").
		Where("mp.role_id = ?", roleID).
		Where("m.key = ?", menuKey)

	switch action {
	case "view":
		query = query.Where("mp.view_action = ?", true)
	case "create":
		query = query.Where("mp.create_action = ?", true)
	case "update":
		query = query.Where("mp.update_action = ?", true)
	case "delete":
		query = query.Where("mp.delete_action = ?", true)
	}

	query.Count(&count)

	return count > 0
}
